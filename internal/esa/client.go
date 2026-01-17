package esa

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// EsaClient はesa.io APIクライアント
type EsaClient struct {
	teamName    string
	accessToken string
	httpClient  *http.Client
}

// NewEsaClient は新しいEsaClientを作成します
func NewEsaClient(teamName, accessToken string) *EsaClient {
	// セキュアなHTTPクライアント設定
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12, // TLS 1.2以上を要求
		},
		Proxy: nil, // HTTPプロキシを無効化（トークン漏洩防止）
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// リダイレクトを禁止（ホスト変更やHTTPダウングレード防止）
			return http.ErrUseLastResponse
		},
	}

	return &EsaClient{
		teamName:    teamName,
		accessToken: accessToken,
		httpClient:  client,
	}
}

// CreatePost は新規記事を作成します
func (c *EsaClient) CreatePost(post *PostInput) (*Post, error) {
	url := fmt.Sprintf("https://api.esa.io/v1/teams/%s/posts", c.teamName)
	return c.doRequestWithRetry("POST", url, post)
}

// UpdatePost は既存記事を更新します
func (c *EsaClient) UpdatePost(postNumber int, post *PostInput) (*Post, error) {
	url := fmt.Sprintf("https://api.esa.io/v1/teams/%s/posts/%d", c.teamName, postNumber)
	return c.doRequestWithRetry("PATCH", url, post)
}

// GetPost は記事を取得します
func (c *EsaClient) GetPost(postNumber int) (*Post, error) {
	url := fmt.Sprintf("https://api.esa.io/v1/teams/%s/posts/%d", c.teamName, postNumber)
	return c.doRequestWithRetry("GET", url, nil)
}

// doRequestWithRetry はリトライ付きでHTTPリクエストを実行します
func (c *EsaClient) doRequestWithRetry(method, url string, payload interface{}) (*Post, error) {
	maxRetries := 3
	backoff := 1 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		post, err := c.doRequest(method, url, payload)
		if err == nil {
			return post, nil
		}

		lastErr = err

		// 最後の試行でない場合はバックオフ
		if i < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2 // 指数バックオフ
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

// doRequest はHTTPリクエストを実行します
func (c *EsaClient) doRequest(method, url string, payload interface{}) (*Post, error) {
	var body io.Reader
	if payload != nil {
		// esa.io APIは {"post": {...}} 形式を要求
		wrapped := map[string]interface{}{
			"post": payload,
		}
		jsonData, err := json.Marshal(wrapped)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを制限付きで読み込み（10MB上限）
	limitedReader := io.LimitReader(resp.Body, 10*1024*1024)
	respBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// ステータスコードチェック
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// エラーメッセージをサニタイズ（最大500文字、制御文字除去）
		errMsg := sanitizeErrorMessage(string(respBody))
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, errMsg)
	}

	// レスポンスをパース
	var result struct {
		Post `json:",inline"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result.Post, nil
}

// sanitizeErrorMessage はエラーメッセージをサニタイズします
func sanitizeErrorMessage(msg string) string {
	// 最大500文字に制限
	if len(msg) > 500 {
		msg = msg[:500] + "..."
	}

	// 制御文字を除去
	var sb strings.Builder
	for _, r := range msg {
		if r >= 32 && r != 127 { // 制御文字以外
			sb.WriteRune(r)
		}
	}

	return sb.String()
}
