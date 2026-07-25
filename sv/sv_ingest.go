package sv

import (
	"Main/model"
	"Main/sv/intake"
	"Main/sv/rag"
	"Main/utils"
	"context"
	"fmt"
	"sync"
)

var (
	taskManager = &intake.TaskManager{}
	pipeline    = &rag.Pipeline{}
	once        sync.Once
)

func init() {
	once.Do(func() {
		taskManager = intake.NewTaskManager()
		var err error
		pipeline, err = rag.NewPipeline(utils.GetCollectionsName())
		if err != nil {
			fmt.Println(err)
		}
	})
}

func NewIngest(repoUrl string) (model.IngestNewDTO, error) {
	task, err := intake.CreateIntakeTask(repoUrl)
	if err != nil {
		return model.IngestNewDTO{}, err
	}
	taskManager.AddTask(&task)
	ctx, cancel := context.WithCancel(context.Background())
	taskManager.RegisterCancel(task.ID, cancel)
	go func() {
		defer taskManager.CancelTask(task.ID)
		err := intake.HandleIntakeTask(ctx, &task, pipeline)
		if err != nil {
			fmt.Println(err)
		}
		cancel()
	}()
	return model.IngestNewDTO{
		TaskId: task.ID,
	}, err
}
func CheckIngest(taskId int64) (model.IngestCheckDTO, error) {
	task, ok := taskManager.GetTaskSnap(taskId)
	if !ok {
		return model.IngestCheckDTO{}, fmt.Errorf("task not found")
	}
	return model.IngestCheckDTO{
		TaskId:      task.ID,
		Status:      task.Status,
		Progress:    task.Progress,
		TotalFile:   task.TotalFiles,
		IndexedFile: task.IndexedFiles,
	}, nil
}
func CancelIngest(taskId int64) error {
	cancel, ok := taskManager.CancelTask(taskId)
	if !ok {
		return fmt.Errorf("task not found")
	}
	cancel()
	task := taskManager.GetTask(taskId)
	if task != nil {
		task.SafeSetStatus("cancelled")
	}
	return nil
}
