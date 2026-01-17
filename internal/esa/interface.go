package esa

// EsaClientInterface はesa.io APIクライアントのインターフェース
type EsaClientInterface interface {
	// CreatePost は新規記事を作成します
	CreatePost(post *PostInput) (*Post, error)

	// UpdatePost は既存記事を更新します
	UpdatePost(postNumber int, post *PostInput) (*Post, error)

	// GetPost は記事を取得します（カテゴリ検証用）
	GetPost(postNumber int) (*Post, error)
}
