package sv

import (
	"Main/model"
	"Main/sv/intake"
	"context"
	"fmt"
)

func NewIngest(newIngestModel model.IngestNewModel) (model.IngestNewDTO, error) {
	task, err := intake.CreateIntakeTask(newIngestModel)
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
		Error:       task.Error,
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
