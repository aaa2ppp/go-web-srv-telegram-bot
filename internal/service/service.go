package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"taskbot/internal/delivery"
	"taskbot/internal/dto"
)

type Repo interface {
	CreateTask(*dto.CreateTask) (*dto.Task, error)
	ListAllTasks() ([]*dto.Task, error)
	ListAssigneeTasks(userID uint64) ([]*dto.Task, error)
	ListOwnerTasks(userID uint64) ([]*dto.Task, error)
	AssignTask(*dto.ChangeTaskState) (*dto.Task, error)
	UnassignTask(*dto.ChangeTaskState) (*dto.Task, error)
	ResolveTask(*dto.ChangeTaskState) (*dto.Task, error)
}

type Service struct {
	Repo   Repo
	Sender delivery.MessageSender
}

func (r Service) ListAllTasks(user *dto.User, _ string) {
	tasks, err := r.Repo.ListAllTasks()
	if err != nil {
		r.sendUnexpecteError(user)
		log.Printf("Service.ListAllTasks: %v", err)
		return
	}

	r.sendTaskList(user, tasks, false)
}

func (r Service) ListOwnerTasks(user *dto.User, _ string) {
	tasks, err := r.Repo.ListOwnerTasks(user.ID)
	if err != nil {
		r.sendUnexpecteError(user)
		log.Printf("Service.ListOwnerTasks: %v", err)
		return
	}

	r.sendTaskList(user, tasks, false)
}

func (r Service) ListAssigneeTasks(user *dto.User, _ string) {
	tasks, err := r.Repo.ListAssigneeTasks(user.ID)
	if err != nil {
		r.sendUnexpecteError(user)
		log.Printf("Service.ListAssigneeTasks: %v", err)
		return
	}

	r.sendTaskList(user, tasks, true)
}

func (r Service) sendTaskList(user *dto.User, tasks []*dto.Task, hideAssignee bool) {

	if len(tasks) == 0 {
		r.Sender.SendMessage(user, "Нет задач")
		return
	}

	buf := &strings.Builder{}

	for _, task := range tasks {
		fmt.Fprintf(buf, "\n%d. %s by @%s\n", task.ID, task.Content, task.Owner.Username)

		if task.Assignee == nil {
			fmt.Fprintf(buf, "/assign_%d\n", task.ID)

		} else if task.Assignee.ID != user.ID {
			fmt.Fprintf(buf, "assignee: @%s\n", task.Assignee.Username)

		} else {
			if !hideAssignee {
				fmt.Fprintf(buf, "assignee: я\n")
			}
			fmt.Fprintf(buf, "/unassign_%d /resolve_%d\n", task.ID, task.ID)
		}
	}

	r.Sender.SendMessage(user, buf.String())
}

func (r Service) sendUnexpecteError(user *dto.User) {
	r.Sender.SendMessage(user, "oops!..")
}

func (r Service) CreateTask(user *dto.User, text string) {

	task, err := r.Repo.CreateTask(&dto.CreateTask{User: user, Content: text})
	if err != nil {
		r.sendUnexpecteError(user)
		log.Printf("Service.CreateTask: %v", err)
		return
	}

	r.Sender.SendMessage(user, fmt.Sprintf("Задача %q создана, id=%d", task.Content, task.ID))
}

func (r Service) AssignTask(user *dto.User, text string) {

	taskID, err := strconv.ParseUint(text, 0, 64)
	if err != nil || taskID < 1 {
		r.Sender.SendMessage(user, "task id must be int >= 1")
		return
	}

	task, err := r.Repo.AssignTask(&dto.ChangeTaskState{User: user, TaskID: taskID})
	if err != nil {
		if !r.handleChangeTaskStateError(user, err) {
			log.Printf("Service.ResolveTask: %v", err)
		}
		return
	}

	r.Sender.SendMessage(user, fmt.Sprintf("Задача %q назначена на вас\n", task.Content))

	if task.Assignee != nil {
		r.Sender.SendMessage(task.Assignee, fmt.Sprintf("Задача %q назначена на @%s\n", task.Content, user.Username))

	} else if task.Owner.ID != user.ID {
		r.Sender.SendMessage(task.Owner, fmt.Sprintf("Задача %q назначена на @%s\n", task.Content, user.Username))
	}
}

func (r Service) UnassignTask(user *dto.User, text string) {

	taskID, err := strconv.ParseUint(text, 0, 64)
	if err != nil || taskID < 1 {
		r.Sender.SendMessage(user, "task id must be int >= 1")
		return
	}

	task, err := r.Repo.UnassignTask(&dto.ChangeTaskState{User: user, TaskID: taskID})
	if err != nil {
		if !r.handleChangeTaskStateError(user, err) {
			log.Printf("Service.UnassignTask: %v", err)
		}
		return
	}

	r.Sender.SendMessage(user, "Принято\n")
	r.Sender.SendMessage(task.Owner, fmt.Sprintf("Задача %q осталась без исполнителя\n", task.Content))
}

func (r Service) ResolveTask(user *dto.User, text string) {

	taskID, err := strconv.ParseUint(text, 0, 64)
	if err != nil || taskID < 1 {
		r.Sender.SendMessage(user, "task id must be int >= 1")
		return
	}

	task, err := r.Repo.ResolveTask(&dto.ChangeTaskState{User: user, TaskID: taskID})
	if err != nil {
		if !r.handleChangeTaskStateError(user, err) {
			log.Printf("Service.ResolveTask: %v", err)
		}
		return
	}

	r.Sender.SendMessage(user, fmt.Sprintf("Задача %q выполнена\n", task.Content))

	if task.Owner.ID != user.ID {
		r.Sender.SendMessage(task.Owner, fmt.Sprintf("Задача %q выполнена @%s\n", task.Content, task.Assignee.Username))
	}
}

func (r Service) handleChangeTaskStateError(user *dto.User, err error) bool {

	switch {
	case errors.Is(err, dto.ErrForbidden):
		r.Sender.SendMessage(user, "Задача не на вас")
		return true
	case errors.Is(err, dto.ErrNotFound):
		r.Sender.SendMessage(user, "Задача не существует или выполнена")
		return true
	}

	r.sendUnexpecteError(user)
	return false
}
