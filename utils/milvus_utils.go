package utils

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func CheckCollectionsName() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	milvusAddr := "localhost:19530"
	token := "root:Milvus"
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: milvusAddr,
		APIKey:  token,
	})
	if err != nil {
		return fmt.Errorf("utils:check error:%v", err)
	}
	defer client.Close(ctx)

	collectionNames, err := client.ListCollections(ctx, milvusclient.NewListCollectionOption())
	if err != nil {
		return fmt.Errorf("utils:check error:%v", err)
	}

	fmt.Println(collectionNames)
	return nil
}
func CheckCollections(collectionName string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	milvusAddr := "localhost:19530"
	token := "root:Milvus"
	client, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: milvusAddr,
		APIKey:  token,
	})
	if err != nil {
		return fmt.Errorf("utils:check error:%v", err)
	}
	defer client.Close(ctx)
	collection, err := client.DescribeCollection(ctx, milvusclient.NewDescribeCollectionOption(collectionName))
	if err != nil {
		return fmt.Errorf("utils:check error:%v", err)
	}

	fmt.Println(collection)

	fmt.Printf("集合名称: %s\n", collection.Name)
	fmt.Printf("集合ID: %d\n", collection.ID)
	fmt.Printf("分片数: %d\n", collection.ShardNum)
	// 遍历字段
	for _, f := range collection.Schema.Fields {
		dim, _ := f.GetDim()
		fmt.Printf("  字段: %s, 类型: %v, 维度: %d, AutoID: %v, 主键: %v\n",
			f.Name, f.DataType, dim, f.AutoID, f.PrimaryKey)
	}
	return nil
}
