package model

type IngestNewModel struct {
	RepoUrl         string   `json:"repo_url"`
	IncludePatterns []string `json:"include_patterns"`
	ExcludePatterns []string `json:"exclude_patterns"`
}

type IngestNewDTO struct {
	TaskId int64 `json:"task_id"`
}
type IngestCheckDTO struct {
	TaskId      int64    `json:"task_id"`
	Status      string   `json:"status"`
	Progress    float64  `json:"progress"`
	TotalFile   int      `json:"total_file"`
	IndexedFile int      `json:"indexed_file"`
	Error       []string `json:"error"`
}
type IngestCancelModel struct {
	TaskId int64 `json:"task_id"`
}
