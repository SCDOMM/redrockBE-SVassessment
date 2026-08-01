package intake

import (
	"Main/sv/rag"
	"Main/utils"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

var extToLang = map[string]string{
	".go":    "Go",
	".py":    "Python",
	".js":    "JavaScript",
	".ts":    "TypeScript",
	".java":  "Java",
	".c":     "C",
	".cpp":   "C++",
	".h":     "C",
	".hpp":   "C++",
	".cs":    "C#",
	".rb":    "Ruby",
	".php":   "PHP",
	".swift": "Swift",
	".kt":    "Kotlin",
	".scala": "Scala",
	".rs":    "Rust",
	".sh":    "Shell",
	".yaml":  "YAML",
	".yml":   "YAML",
	".json":  "JSON",
	".xml":   "XML",
	".toml":  "TOML",
	".md":    "Markdown",
	".txt":   "Text",
	".sql":   "SQL",
	".r":     "R",
	".m":     "Objective-C",
	".mm":    "Objective-C++",
}

func getLanguage(filePath string) string {
	//转换文件扩展名
	ext := strings.ToLower(filepath.Ext(filePath))
	if lang, ok := extToLang[ext]; ok {
		return lang
	}
	return "Unknown"
}

func HandleIntakeTask(ctx context.Context, task *IntakeTask, pipeline *rag.Pipeline) error {
	defer os.RemoveAll(task.TempDir)
	task.SafeSetStatus("running")
	resource := task.RepoURL

	//加载索引和集合
	if err := pipeline.CreateIndex(ctx, false, true); err != nil {
		task.SafeSetError(fmt.Errorf("创建索引失败: %v", err))
		return err
	}
	err := pipeline.LoadCollections(ctx)
	if err != nil {
		task.SafeSetError(fmt.Errorf("加载集合失败: %w", err))
		return err
	}
	//获取映射
	existingFiles, err := pipeline.GetFileRecords(ctx, resource)
	if err != nil {
		task.SafeSetError(fmt.Errorf("查询已有文件记录失败: %w", err))
		return err
	}
	localFiles := make(map[string]string)
	var toInsert []string // 新增的文件
	var toUpdate []string // 更新的文件
	var toDelete []string // 需要删除的文件

	// 遍历临时仓库，分类
	err = filepath.WalkDir(task.TempDir, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			return err
		}
		//跳过过滤目录
		if d.IsDir() {
			filterPathMap := make(map[string]bool, len(task.FilterPath))
			for _, e := range task.FilterPath {
				filterPathMap[strings.ToLower(e)] = true
			}

			dirName := d.Name()
			if filterPathMap[dirName] {
				return filepath.SkipDir
			}
			return nil
		}

		//能获取吗
		if !isIndexableFile(path, task.Include, task.Exclude, d) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			task.SafeSetError(fmt.Errorf("读取文件失败: %s: %v", path, err))
			return nil
		}
		relPath, err := filepath.Rel(task.TempDir, path)
		if err != nil {
			task.SafeSetError(fmt.Errorf("计算相对路径失败: %v", err))
			return nil
		}

		//用哈希值摁造文件更新
		hash := utils.Sha256Hex(data)
		localFiles[relPath] = hash
		existingHash, exists := existingFiles[relPath]
		if !exists {
			toInsert = append(toInsert, relPath) // 新增
		} else if existingHash != hash {
			toDelete = append(toDelete, relPath) // 旧记录删除
			toUpdate = append(toUpdate, relPath) // 重新插入
		}
		// hash 相同
		return nil
	})
	if errors.Is(err, context.Canceled) {
		task.SafeSetStatus("canceled")
		task.SafeSetError(fmt.Errorf("intake task cancelled"))
		return err
	}
	if err != nil {
		task.SafeSetError(fmt.Errorf("遍历仓库失败: %v", err))
		return err
	}
	// 找出已删除的文件
	for path := range existingFiles {
		if _, exists := localFiles[path]; !exists {
			toDelete = append(toDelete, path)
		}
	}

	task.mu.Lock()
	task.TotalFiles = len(toInsert) + len(toUpdate) + len(toDelete)
	task.mu.Unlock()

	for _, path := range toDelete {
		if err := pipeline.DeleteFile(ctx, resource, path); err != nil {
			task.SafeSetError(fmt.Errorf("删除文件失败: %s: %v", path, err))
			continue
		}
		task.SafeIncrementProgress()
	}
	// 执行新增和更新操作,重新遍历本地文件,只处理toInsert和toUpdate
	err = filepath.WalkDir(task.TempDir, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			return err
		}
		if d.IsDir() {
			filterPathMap := make(map[string]bool, len(task.FilterPath))
			for _, e := range task.FilterPath {
				filterPathMap[strings.ToLower(e)] = true
			}

			dirName := d.Name()
			if filterPathMap[dirName] {
				return filepath.SkipDir
			}
			return nil
		}
		if !isIndexableFile(path, task.Include, task.Exclude, d) {
			return nil
		}

		relPath, _ := filepath.Rel(task.TempDir, path)
		// 只处理toInsert或toUpdate中的文件
		if !contains(toInsert, relPath) && !contains(toUpdate, relPath) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			task.SafeSetError(fmt.Errorf("读取文件失败: %s: %v", path, err))
			return nil
		}
		docText := string(data)
		hash := utils.Sha256Hex(data)
		language := getLanguage(relPath)
		// fileHash插入文档
		if err := pipeline.InsertDocument(ctx, docText, resource, relPath, hash, language); err != nil {
			task.SafeSetError(fmt.Errorf("插入文档失败: %s: %v", relPath, err))
			return nil
		}
		task.SafeIncrementProgress()
		return nil
	})

	//错误处理，刷入
	if errors.Is(err, context.Canceled) {
		task.SafeSetStatus("canceled")
		task.SafeSetError(fmt.Errorf("intake task cancelled"))
		return err
	}
	if err != nil {
		task.SafeSetError(fmt.Errorf("第二次遍历插入失败: %v", err))
		return err
	}

	flushTask, err := pipeline.MilvusClient.Flush(ctx, milvusclient.NewFlushOption(pipeline.CollectionName))
	if err != nil {
		task.SafeSetError(fmt.Errorf("刷入失败: %v", err))
		return fmt.Errorf("rag: fail to flush: %w", err)
	}
	if err = flushTask.Await(ctx); err != nil {
		task.SafeSetError(fmt.Errorf("刷入超时: %v", err))
		return fmt.Errorf("rag: flush timeout: %w", err)
	}
	task.SafeSetStatus("completed")
	task.mu.Lock()
	task.Progress = 1.0
	task.mu.Unlock()
	return nil
}
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
