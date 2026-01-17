package esa

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// StubEsaClient はテスト用のスタブクライアント
type StubEsaClient struct {
	createPostFunc func(*PostInput) (*Post, error)
	updatePostFunc func(int, *PostInput) (*Post, error)
	getPostFunc    func(int) (*Post, error)
}

// CreatePost はスタブの実装
func (s *StubEsaClient) CreatePost(post *PostInput) (*Post, error) {
	if s.createPostFunc != nil {
		return s.createPostFunc(post)
	}
	return &Post{Number: 123, Name: post.Name, Category: post.Category}, nil
}

// UpdatePost はスタブの実装
func (s *StubEsaClient) UpdatePost(postNumber int, post *PostInput) (*Post, error) {
	if s.updatePostFunc != nil {
		return s.updatePostFunc(postNumber, post)
	}
	return &Post{Number: postNumber, Name: post.Name, Category: post.Category}, nil
}

// GetPost はスタブの実装
func (s *StubEsaClient) GetPost(postNumber int) (*Post, error) {
	if s.getPostFunc != nil {
		return s.getPostFunc(postNumber)
	}
	return &Post{Number: postNumber, Name: "Test Post", Category: "LLM/Tasks"}, nil
}

func TestStubEsaClient(t *testing.T) {
	stub := &StubEsaClient{}

	// CreatePostのテスト
	input := &PostInput{
		Name:     "Test Post",
		Category: "LLM/Tasks",
		BodyMD:   "## Content",
		WIP:      false,
	}
	post, err := stub.CreatePost(input)
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if post.Name != "Test Post" {
		t.Errorf("Post.Name = %v, want Test Post", post.Name)
	}

	// UpdatePostのテスト
	post, err = stub.UpdatePost(123, input)
	if err != nil {
		t.Fatalf("UpdatePost() error = %v", err)
	}
	if post.Number != 123 {
		t.Errorf("Post.Number = %v, want 123", post.Number)
	}

	// GetPostのテスト
	post, err = stub.GetPost(123)
	if err != nil {
		t.Fatalf("GetPost() error = %v", err)
	}
	if post.Number != 123 {
		t.Errorf("Post.Number = %v, want 123", post.Number)
	}
}

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{
			name: "通常のメッセージ",
			msg:  "Error: something went wrong",
			want: "Error: something went wrong",
		},
		{
			name: "500文字超過",
			msg:  strings.Repeat("a", 600),
			want: strings.Repeat("a", 500) + "...",
		},
		{
			name: "制御文字を含む",
			msg:  "Error\nwith\tcontrol\rchars",
			want: "Errorwithcontrolchars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeErrorMessage(tt.msg)
			if got != tt.want {
				t.Errorf("sanitizeErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreatePostRequestFormat(t *testing.T) {
	// モックHTTPサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストボディを読み取る
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// {"post": {...}} 形式であることを検証
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to unmarshal request: %v", err)
		}

		// "post" キーが存在することを確認
		postData, ok := req["post"]
		if !ok {
			t.Errorf("Request does not contain 'post' key. Got: %v", req)
		}

		// postの中身を検証
		postMap, ok := postData.(map[string]interface{})
		if !ok {
			t.Errorf("'post' value is not an object")
		}

		// フィールドを検証
		if postMap["name"] != "Test Post" {
			t.Errorf("name = %v, want Test Post", postMap["name"])
		}
		if postMap["category"] != "LLM/Tasks" {
			t.Errorf("category = %v, want LLM/Tasks", postMap["category"])
		}

		// 成功レスポンスを返す
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"number": 123, "name": "Test Post", "category": "LLM/Tasks", "url": "https://example.esa.io/posts/123"}`))
	}))
	defer server.Close()

	// クライアントを作成（モックサーバーを使用）
	client := NewEsaClient("test-team", "test-token")
	// モックサーバーのURLを使用するためにdoRequestを直接呼び出す
	input := &PostInput{
		Name:     "Test Post",
		Category: "LLM/Tasks",
		Tags:     []string{"test"},
		BodyMD:   "## Test",
		WIP:      false,
	}

	post, err := client.doRequest("POST", server.URL, input)
	if err != nil {
		t.Fatalf("doRequest() error = %v", err)
	}

	if post.Name != "Test Post" {
		t.Errorf("Post.Name = %v, want Test Post", post.Name)
	}
}
