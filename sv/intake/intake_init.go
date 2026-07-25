package intake

import (
	"Main/utils"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
)

type IntakeTask struct {
	ID           int64
	RepoURL      string
	Status       string  // pending / running / completed / failed
	Progress     float64 // 0.0 ~ 1.0
	TempDir      string
	TotalFiles   int
	IndexedFiles int
	CreatedAt    time.Time
	Error        []string
}

func CreateIntakeTask(repoURL string) (IntakeTask, error) {
	snowFlake := utils.NewSnowflake(utils.GetMachineId())
	id := snowFlake.GenerateID()
	task := IntakeTask{
		ID:           id,
		RepoURL:      repoURL,
		Status:       "pending",
		Progress:     0.0,
		TotalFiles:   0,
		IndexedFiles: 0,
		CreatedAt:    time.Now(),
		Error:        nil,
	}

	//创建临时目录
	dirName := "intake_" + strconv.FormatInt(id, 10)
	tempDir, err := os.MkdirTemp("", dirName)
	if err != nil {
		task.Error = append(task.Error, "fail to create temp dir")
		task.Status = "failed"
		return task, err
	}
	task.TempDir = tempDir

	//克隆仓库
	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:   repoURL,
		Depth: 1,
	})
	if err != nil {
		err1 := os.RemoveAll(tempDir)
		task.Status = "failed"
		task.Error = append(task.Error, "fail to clone repository"+err.Error())
		if err1 != nil {
			err = fmt.Errorf("fail to clone repository" + err.Error() + ",fail to delete temp directory" + err1.Error())
		}
		return task, err
	}

	//统计文件数目
	fileCount := 0
	err = filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// 跳过不需要的目录
			dirName := d.Name()
			if dirName == ".git" || dirName == "vendor" || dirName == "node_modules" ||
				dirName == ".idea" || dirName == "bin" || dirName == "obj" {
				return filepath.SkipDir
			}
			return nil
		}
		// 检查文件是否应该被索引
		if isIndexableFile(path, d) {
			fileCount++
		}
		return nil
	})

	if err != nil {
		err1 := os.RemoveAll(tempDir)
		task.Status = "failed"
		task.Error = append(task.Error, "fail to range repository:"+err.Error())
		if err1 != nil {
			err = fmt.Errorf("fail to range repository!" + err.Error() + "fail to delete temp directory" + err1.Error())
		}
		return task, err
	}

	task.TotalFiles = fileCount
	task.Status = "running"
	return task, nil
}

func isIndexableFile(path string, d fs.DirEntry) bool {
	info, err := d.Info()
	if err != nil {
		return false
	}
	if info.Size() > 1*1024*1024 {
		return false
	}
	// 根据扩展名过滤二进制文件
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
		".ico": true, ".svg": true, ".webp": true,
		".mp3": true, ".wav": true, ".mp4": true, ".avi": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".zip": true, ".tar": true, ".gz": true, ".bz2": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".obj": true, ".lib": true,
		".ttf": true, ".otf": true, ".woff": true,
	}
	if binaryExts[ext] {
		return false
	}
	// 只索引文本/代码文件
	codeExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".java": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true, ".cs": true,
		".rb": true, ".php": true, ".swift": true, ".kt": true, ".scala": true,
		".rs": true, ".sh": true, ".bash": true, ".zsh": true,
		".yaml": true, ".yml": true, ".json": true, ".xml": true, ".toml": true,
		".md": true, ".txt": true, ".html": true, ".css": true, ".scss": true,
		".sql": true, ".r": true, ".m": true, ".mm": true,
	}
	if codeExts[ext] {
		return true
	}
	return false
}
