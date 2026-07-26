package intake

import (
	"Main/sv/rag"
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
	resource := task.RepoURL
	task.SafeSetStatus("running")

	err := filepath.WalkDir(task.TempDir, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirName := d.Name()
			if dirName == ".git" || dirName == "vendor" || dirName == "node_modules" ||
				dirName == ".idea" || dirName == "bin" || dirName == "obj" {
				return filepath.SkipDir
			}
			return nil
		}
		//能读吗
		if !isIndexableFile(path, task.Include, task.Exclude, d) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			task.SafeSetError(fmt.Errorf("读取文件失败: " + path + ": " + err.Error()))
			return nil
		}
		docText := string(data)
		// 计算相对路径
		relPath, err := filepath.Rel(task.TempDir, path)
		if err != nil {
			task.SafeSetError(fmt.Errorf("计算相对路径失败: " + err.Error()))
			return nil
		}
		// 获取语言
		language := getLanguage(relPath)
		// 插入到 Milvus
		err = pipeline.InsertDocument(ctx, docText, resource, relPath, language)
		if err != nil {
			task.SafeSetError(fmt.Errorf("插入文档失败: " + relPath + ": " + err.Error()))
			return nil
		}
		// 更新进度
		task.SafeIncrementProgress()
		return nil
	})
	if errors.Is(err, context.Canceled) {
		task.SafeSetStatus("canceled")
		task.SafeSetError(fmt.Errorf("intake task cancelled"))
		return err
	}
	if err != nil {
		task.SafeSetError(fmt.Errorf("摄取目录失败: " + err.Error()))
		return err
	}

	err = pipeline.CreateIndex(ctx, false, true)
	if err != nil {
		task.SafeSetError(fmt.Errorf("创建索引失败: " + err.Error()))
		return err
	}

	flushTask, err := pipeline.MilvusClient.Flush(ctx, milvusclient.NewFlushOption(pipeline.CollectionName))
	if err != nil {
		task.SafeSetError(fmt.Errorf("刷入失败: " + err.Error()))
		return fmt.Errorf("rag: fail to flush: %w", err)
	}
	if err = flushTask.Await(ctx); err != nil {
		task.SafeSetError(fmt.Errorf("刷入失败: " + err.Error()))
		return fmt.Errorf("rag: flush timeout: %w", err)
	}

	// 所有文件处理完毕
	task.SafeSetStatus("completed")
	task.mu.Lock()
	task.Progress = 1.0
	task.mu.Unlock()
	return nil
}
