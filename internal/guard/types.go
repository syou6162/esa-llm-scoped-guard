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
	ID          string     `yaml:"id"`
	Title       string     `yaml:"title"`
	Status      TaskStatus `yaml:"status"`
	Summary     []string   `yaml:"summary"`
	Description string     `yaml:"description"`
	GitHubURLs  []string   `yaml:"github_urls,omitempty"`
	DependsOn   []string   `yaml:"depends_on,omitempty"`
}

// Body は本文の構造体
type Body struct {
	Background   string   `yaml:"background"`
	RelatedLinks []string `yaml:"related_links,omitempty"`
	Instructions []string `yaml:"instructions,omitempty"`
	Tasks        []Task   `yaml:"tasks"`
}

// PostInput は入力YAMLの構造体
type PostInput struct {
	CreateNew  bool   `yaml:"create_new,omitempty"`  // 新規作成フラグ
	PostNumber *int   `yaml:"post_number,omitempty"` // 更新時に指定
	Name       string `yaml:"name"`                  // 必須
	Category   string `yaml:"category"`              // 必須
	Body       Body   `yaml:"body"`                  // 必須
}
