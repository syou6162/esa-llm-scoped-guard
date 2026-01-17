package guard

// TaskStatus はタスクのステータスを表す型
type TaskStatus string

const (
	TaskStatusNotStarted TaskStatus = "not_started"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusInReview   TaskStatus = "in_review"
	TaskStatusCompleted  TaskStatus = "completed"
)

// Task はタスクの構造体
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Status      TaskStatus `json:"status"`
	Description string     `json:"description"`
}

// Body は本文の構造体
type Body struct {
	Background   string   `json:"background"`
	RelatedLinks []string `json:"related_links,omitempty"`
	Tasks        []Task   `json:"tasks"`
}

// PostInput は入力JSONの構造体
type PostInput struct {
	PostNumber *int   `json:"post_number,omitempty"` // 省略時は新規作成
	Name       string `json:"name"`                  // 必須
	Category   string `json:"category"`              // 必須
	Body       Body   `json:"body"`                  // 必須
}
