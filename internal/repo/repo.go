package repo

import (
	"fmt"
	"sync"

	"taskbot/internal/dto"
)

type tUserRec struct {
	id       uint64
	username string
	chatID   uint64
}

type tTaskRec struct {
	id       uint64
	content  string
	creator  *tUserRec
	executor *tUserRec
	resolved bool
}

func newUserResp(rec *tUserRec) *dto.User {
	if rec == nil {
		return nil
	}
	return &dto.User{
		ID:       rec.id,
		Username: rec.username,
		CharID:   rec.chatID,
	}
}

func newTaskResp(rec *tTaskRec) *dto.Task {
	if rec == nil {
		return nil
	}
	return &dto.Task{
		ID:       rec.id,
		Content:  rec.content,
		Owner:    newUserResp(rec.creator),
		Assignee: newUserResp(rec.executor),
	}
}

func New() *Repo {
	return &Repo{
		users: map[uint64]*tUserRec{},
		tasks: []*tTaskRec{},
	}
}

type Repo struct {
	mu     sync.RWMutex
	users  map[uint64]*tUserRec
	tasks  []*tTaskRec
	taskID uint64
}

func (r *Repo) ListAllTasks() ([]*dto.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*dto.Task, 0, len(r.tasks))
	for _, rec := range r.tasks {
		if !rec.resolved {
			list = append(list, newTaskResp(rec))
		}
	}

	return list, nil
}

func (r *Repo) ListOwnerTasks(userID uint64) ([]*dto.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*dto.Task
	for _, task := range r.tasks {
		if !task.resolved && task.creator != nil && task.creator.id == userID {
			list = append(list, newTaskResp(task))
		}
	}
	return list, nil
}

func (r *Repo) ListAssigneeTasks(userID uint64) ([]*dto.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*dto.Task
	for _, task := range r.tasks {
		if !task.resolved && task.executor != nil && task.executor.id == userID {
			list = append(list, newTaskResp(task))
		}
	}
	return list, nil
}

// creates user if not exists
func (r *Repo) createUser(req *dto.User) *tUserRec {
	user, ok := r.users[req.ID]
	if !ok {
		user = &tUserRec{
			id:       req.ID,
			username: req.Username,
			chatID:   req.CharID,
		}
		r.users[user.id] = user
	}
	return user
}

func (r *Repo) CreateTask(req *dto.CreateTask) (*dto.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	creator := r.createUser(req.User)

	r.taskID++
	task := &tTaskRec{
		id:      r.taskID,
		content: req.Content,
		creator: creator,
	}
	r.tasks = append(r.tasks, task)

	return newTaskResp(task), nil
}

func (r *Repo) AssignTask(req *dto.ChangeTaskState) (*dto.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task := r.tasks[req.TaskID-1]
	if task == nil || task.resolved {
		return nil, fmt.Errorf("task %w", dto.ErrNotFound)
	}

	user := r.createUser(req.User)
	if user == nil {
		return nil, fmt.Errorf("user %w", dto.ErrNotFound)
	}

	resp := newTaskResp(task)
	task.executor = user

	return resp, nil
}

func (r *Repo) UnassignTask(req *dto.ChangeTaskState) (*dto.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task := r.tasks[req.TaskID-1]
	if task == nil || task.resolved {
		return nil, fmt.Errorf("task %w", dto.ErrNotFound)
	}

	if task.executor == nil || task.executor.id != req.User.ID {
		return nil, dto.ErrForbidden
	}

	resp := newTaskResp(task)
	task.executor = nil

	return resp, nil
}

func (r *Repo) ResolveTask(req *dto.ChangeTaskState) (*dto.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task := r.tasks[req.TaskID-1]
	if task == nil || task.resolved {
		return nil, fmt.Errorf("task %w", dto.ErrNotFound)
	}

	if task.executor == nil || task.executor.id != req.User.ID {
		return nil, dto.ErrForbidden
	}

	resp := newTaskResp(task)
	task.resolved = true

	return resp, nil
}
