package intake

import "sync"

type TaskManager struct {
	TaskMap map[int64]*IntakeTask
	rwMutex sync.RWMutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		TaskMap: make(map[int64]*IntakeTask),
		rwMutex: sync.RWMutex{},
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
