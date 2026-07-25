package sv

import (
	"Main/sv/intake"
	"Main/sv/rag"
	"Main/utils"
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
