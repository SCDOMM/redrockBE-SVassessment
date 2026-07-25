package model

type IngestNewModel struct {
	RepoUrl string `json:"repo_url"`
}

type IngestNewDTO struct {
	TaskId string `json:"task_id"`
}
type IngestCheckDTO struct {
	TaskId      string  `json:"task_id"`
	Status      string  `json:"status"`
	Progress    float64 `json:"progress"`
	TotalFile   int     `json:"total_file"`
	IndexedFile int     `json:"indexed_file"`
}
