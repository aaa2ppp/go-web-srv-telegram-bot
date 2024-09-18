package dto

import "errors"

type User struct {
	ID       uint64
	Username string
	CharID   uint64
}

type Task struct {
	ID       uint64
	Content  string
	Owner    *User
	Assignee *User
}

type CreateTask struct {
	Content string
	User    *User
}

type ChangeTaskState struct {
	TaskID uint64
	User   *User
}

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
)
