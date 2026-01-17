package guard

// Body は本文の構造体
type Body struct {
	Background string `json:"background"`
}

// PostInput は入力JSONの構造体
type PostInput struct {
	PostNumber *int   `json:"post_number,omitempty"` // 省略時は新規作成
	Name       string `json:"name"`                  // 必須
	Category   string `json:"category"`              // 必須
	Body       Body   `json:"body"`                  // 必須
}
