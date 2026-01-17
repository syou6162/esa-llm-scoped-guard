package main

// PostInput は入力JSONの構造体
type PostInput struct {
	PostNumber *int     `json:"post_number,omitempty"` // 省略時は新規作成
	Name       string   `json:"name"`                  // 必須
	Category   string   `json:"category"`              // 必須
	Tags       []string `json:"tags,omitempty"`        // オプション
	BodyMD     string   `json:"body_md"`               // 必須
}
