package esa

import (
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
