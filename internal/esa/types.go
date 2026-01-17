package esa

// PostInput はesa.io APIへの投稿リクエスト
type PostInput struct {
	Name     string   `json:"name"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	BodyMD   string   `json:"body_md"`
	WIP      bool     `json:"wip"`
	Message  string   `json:"message,omitempty"`
}

// Post はesa.io APIからのレスポンス
type Post struct {
	Number   int      `json:"number"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	BodyMD   string   `json:"body_md"`
	WIP      bool     `json:"wip"`
	URL      string   `json:"url"`
}
