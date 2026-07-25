package intake

import (
	"context"
	"sync"
)

type TaskManager struct {
	TaskMap   map[int64]*IntakeTask
	CancelMap map[int64]context.CancelFunc
	rwMutex   sync.RWMutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		TaskMap:   make(map[int64]*IntakeTask),
		CancelMap: make(map[int64]context.CancelFunc),
		rwMutex:   sync.RWMutex{},
	}
}
func (tm *TaskManager) AddTask(intakeTask *IntakeTask) {
	tm.rwMutex.Lock()
	defer tm.rwMutex.Unlock()

	tm.TaskMap[intakeTask.ID] = intakeTask
}
func (tm *TaskManager) GetTask(id int64) *IntakeTask {
	tm.rwMutex.RLock()
	defer tm.rwMutex.RUnlock()

	return tm.TaskMap[id]
}
func (tm *TaskManager) GetTaskSnap(id int64) (IntakeTask, bool) {
	tm.rwMutex.RLock()
	task, ok := tm.TaskMap[id]
	tm.rwMutex.RUnlock()
	if !ok {
		return IntakeTask{}, false
	}
	task.mu.Lock()
	defer task.mu.Unlock()
	snapshot := *task
	if task.Error != nil {
		snapshot.Error = make([]string, len(task.Error))
		copy(snapshot.Error, task.Error)
	}
	//放置锁被继承
	snapshot.mu = sync.Mutex{}
	return snapshot, true
}
func (tm *TaskManager) RegisterCancel(id int64, cancelFunc context.CancelFunc) {
	tm.rwMutex.Lock()
	defer tm.rwMutex.Unlock()
	tm.CancelMap[id] = cancelFunc
}
func (tm *TaskManager) CancelTask(id int64) (context.CancelFunc, bool) {
	tm.rwMutex.Lock()
	defer tm.rwMutex.Unlock()
	cancel, ok := tm.CancelMap[id]
	if ok {
		delete(tm.CancelMap, id)
	}
	return cancel, ok
}
func (tm *TaskManager) DeleteTask(id int64) {
	tm.rwMutex.Lock()
	defer tm.rwMutex.Unlock()
	delete(tm.TaskMap, id)
}
func (tm *TaskManager) UpdateTask(id int64, updateFunc func(*IntakeTask)) {
	tm.rwMutex.Lock()
	defer tm.rwMutex.Unlock()

	updateFunc(tm.TaskMap[id])
}
