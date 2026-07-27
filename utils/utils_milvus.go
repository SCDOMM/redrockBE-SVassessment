package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func CheckCollectionsName(ctx context.Context) error {
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
func CheckCollections(ctx context.Context, collectionName string) error {

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
func BuildFilterExpr(filterPaths []string) string {
	if len(filterPaths) == 0 {
		return ""
	}
	var conds []string
	for _, fp := range filterPaths {
		fp = strings.TrimSpace(fp)
		if fp == "" {
			continue
		}
		var expr string
		if strings.HasSuffix(fp, "/") {
			// 目录过滤：like "path%"
			expr = fmt.Sprintf(`file_path like "%s%s"`, fp, "%") // 如 "src/%"
		} else if strings.HasPrefix(fp, ".") {
			// 文件后缀：like "%.suffix"
			expr = fmt.Sprintf(`file_path like "%%%s"`, fp) // 如 "%.go"
		} else {
			expr = fmt.Sprintf(`file_path == "%s"`, fp)
		}
		conds = append(conds, expr)
	}
	if len(conds) == 0 {
		return ""
	}
	return strings.Join(conds, " and ")
}
