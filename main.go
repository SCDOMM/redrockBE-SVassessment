package main

import (
	"Main/router"
	"Main/sv/rag"
	"Main/utils"
	"context"
	"fmt"
)

func main() {
	pipeline, _ := rag.NewPipeline(utils.GetCollectionsName())
	err := pipeline.CreateCollection(context.Background())
	if err != nil {
		fmt.Printf("Create collection failed, err:%v\n", err)
	}
	//出现冲突时再调用这个函数
	//pipeline.DropCollection(context.Background())
	pipeline.Close()
	//utils.CheckCollections(context.Background(), utils.GetCollectionsName())
	router.InitRouter()
}
