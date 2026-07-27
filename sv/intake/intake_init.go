package intake

import (
	"Main/model"
	"Main/utils"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	FilterPath   []string
	Include      []string
	Exclude      []string
	CreatedAt    time.Time
	mu           sync.Mutex
	Error        []string
}

func (t *IntakeTask) SafeSetStatus(status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status
}

func (t *IntakeTask) SafeIncrementProgress() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.IndexedFiles++
	if t.TotalFiles > 0 {
		t.Progress = float64(t.IndexedFiles) / float64(t.TotalFiles)
	}
}

func (t *IntakeTask) SafeSetError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Error = append(t.Error, err.Error())
}

func CreateIntakeTask(newIntakeModel model.IngestNewModel) (IntakeTask, error) {
	snowFlake := utils.NewSnowflake(utils.GetMachineId())
	id := snowFlake.GenerateID()
	task := IntakeTask{
		ID:           id,
		RepoURL:      newIntakeModel.RepoUrl,
		Status:       "pending",
		Progress:     0.0,
		TotalFiles:   0,
		IndexedFiles: 0,
		FilterPath:   newIntakeModel.FilterPath,
		Include:      newIntakeModel.IncludePatterns,
		Exclude:      newIntakeModel.ExcludePatterns,
		CreatedAt:    time.Now(),
		Error:        nil,
	}

	//创建临时目录
	dirName := "intake_" + strconv.FormatInt(id, 10)
	tempDir, err := os.MkdirTemp("", dirName)
	if err != nil {
		task.SafeSetError(fmt.Errorf("fail to create temp dir"))
		return task, err
	}
	task.TempDir = tempDir

	//克隆仓库
	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:   newIntakeModel.RepoUrl,
		Depth: 1,
	})
	if err != nil {
		err1 := os.RemoveAll(tempDir)
		task.SafeSetError(fmt.Errorf("fail to clone repository" + err.Error()))
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
		if isIndexableFile(path, task.Include, task.Exclude, d) {
			fileCount++
		}
		return nil
	})

	if err != nil {
		err1 := os.RemoveAll(tempDir)
		task.SafeSetError(fmt.Errorf("fail to range repository:" + err.Error()))
		if err1 != nil {
			err = fmt.Errorf("fail to range repository!" + err.Error() + "fail to delete temp directory" + err1.Error())
		}
		return task, err
	}

	task.TotalFiles = fileCount

	return task, nil
}

func isIndexableFile(path string, include []string, exclude []string, d fs.DirEntry) bool {
	info, err := d.Info()
	if err != nil {
		return false
	}
	if info.Size() > 1*1024*1024 { // 1MB
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))

	//过滤不包含文件
	excludeMap := make(map[string]bool, len(exclude))
	for _, e := range exclude {
		excludeMap[strings.ToLower(e)] = true
	}
	if excludeMap[ext] {
		return false
	}

	//保留包含文件
	includeMap := make(map[string]bool, len(include))
	for _, i := range include {
		includeMap[strings.ToLower(i)] = true
	}
	if includeMap[ext] {
		return true
	}
	return false
}
