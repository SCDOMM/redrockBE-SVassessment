package main

import "Main/router"

func main() {
	//pipeline, _ := rag.NewPipeline(utils.GetCollectionsName())
	//pipeline.DropCollection(context.Background())
	//pipeline.CreateCollection(context.Background())
	//utils.CheckCollections(context.Background(), utils.GetCollectionsName())
	router.InitRouter()

}
