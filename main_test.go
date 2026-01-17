package main

import (
	"testing"

	"github.com/syou6162/esa-llm-scoped-guard/internal/esa"
	"github.com/syou6162/esa-llm-scoped-guard/internal/guard"
)

// stubEsaClient はテスト用のスタブクライアント
type stubEsaClient struct {
	createPostFunc func(*esa.PostInput) (*esa.Post, error)
	updatePostFunc func(int, *esa.PostInput) (*esa.Post, error)
	getPostFunc    func(int) (*esa.Post, error)
}

func (s *stubEsaClient) CreatePost(post *esa.PostInput) (*esa.Post, error) {
	if s.createPostFunc != nil {
		return s.createPostFunc(post)
	}
	return &esa.Post{Number: 123, Name: post.Name, Category: post.Category}, nil
}

func (s *stubEsaClient) UpdatePost(postNumber int, post *esa.PostInput) (*esa.Post, error) {
	if s.updatePostFunc != nil {
		return s.updatePostFunc(postNumber, post)
	}
	return &esa.Post{Number: postNumber, Name: post.Name, Category: post.Category}, nil
}

func (s *stubEsaClient) GetPost(postNumber int) (*esa.Post, error) {
	if s.getPostFunc != nil {
		return s.getPostFunc(postNumber)
	}
	return &esa.Post{Number: postNumber, Name: "Test Post", Category: "LLM/Tasks"}, nil
}

// TestUpdatePost_ExistingCategoryNotAllowed は既存記事のカテゴリが許可範囲外の場合のテスト
func TestUpdatePost_ExistingCategoryNotAllowed(t *testing.T) {
	// スタブクライアントを作成
	stub := &stubEsaClient{
		getPostFunc: func(postNumber int) (*esa.Post, error) {
			// 既存記事は許可範囲外のカテゴリ
			return &esa.Post{
				Number:   postNumber,
				Name:     "Existing Post",
				Category: "Unauthorized/Category",
			}, nil
		},
	}

	config := &Config{
		AllowedCategories: []string{"LLM/Tasks"},
	}

	input := &PostInput{
		PostNumber: intPtr(123),
		Name:       "Updated Post",
		Category:   "LLM/Tasks",
		BodyMD:     "## Content",
	}

	// 既存カテゴリが許可範囲外なので拒否されるべき
	_, err := updatePostWithClient(stub, config, input)
	if err == nil {
		t.Errorf("Expected error when existing category is not allowed, got nil")
	}
}

// TestUpdatePost_CategoryChange はカテゴリ変更を試みた場合のテスト
func TestUpdatePost_CategoryChange(t *testing.T) {
	// スタブクライアントを作成
	stub := &stubEsaClient{
		getPostFunc: func(postNumber int) (*esa.Post, error) {
			// 既存記事のカテゴリ
			return &esa.Post{
				Number:   postNumber,
				Name:     "Existing Post",
				Category: "LLM/Tasks/Old",
			}, nil
		},
	}

	config := &Config{
		AllowedCategories: []string{"LLM/Tasks"},
	}

	input := &PostInput{
		PostNumber: intPtr(123),
		Name:       "Updated Post",
		Category:   "LLM/Tasks/New", // 異なるカテゴリ
		BodyMD:     "## Content",
	}

	// カテゴリ変更は拒否されるべき
	_, err := updatePostWithClient(stub, config, input)
	if err == nil {
		t.Errorf("Expected error when category change is attempted, got nil")
	}
}

// TestUpdatePost_Success は正常な更新のテスト
func TestUpdatePost_Success(t *testing.T) {
	// スタブクライアントを作成
	stub := &stubEsaClient{
		getPostFunc: func(postNumber int) (*esa.Post, error) {
			return &esa.Post{
				Number:   postNumber,
				Name:     "Existing Post",
				Category: "LLM/Tasks",
			}, nil
		},
		updatePostFunc: func(postNumber int, post *esa.PostInput) (*esa.Post, error) {
			return &esa.Post{
				Number:   postNumber,
				Name:     post.Name,
				Category: post.Category,
				URL:      "https://example.esa.io/posts/123",
			}, nil
		},
	}

	config := &Config{
		AllowedCategories: []string{"LLM/Tasks"},
	}

	input := &PostInput{
		PostNumber: intPtr(123),
		Name:       "Updated Post",
		Category:   "LLM/Tasks", // 既存と同じカテゴリ
		BodyMD:     "## Content",
	}

	// 正常に更新されるべき
	post, err := updatePostWithClient(stub, config, input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if post.Name != "Updated Post" {
		t.Errorf("Post.Name = %v, want Updated Post", post.Name)
	}
}

// updatePostWithClient はテスト用のヘルパー関数
func updatePostWithClient(client esa.EsaClientInterface, config *Config, input *PostInput) (*esa.Post, error) {
	// 既存記事を取得
	existingPost, err := client.GetPost(*input.PostNumber)
	if err != nil {
		return nil, err
	}

	// 更新リクエストの妥当性を検証
	if err := guard.ValidateUpdateRequest(existingPost.Category, input.Category, config.AllowedCategories); err != nil {
		return nil, err
	}

	esaInput := &esa.PostInput{
		Name:     input.Name,
		Category: input.Category,
		Tags:     []string{},
		BodyMD:   input.BodyMD,
		WIP:      false,
	}

	return client.UpdatePost(*input.PostNumber, esaInput)
}
