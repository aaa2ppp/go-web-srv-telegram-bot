package delivery

import "taskbot/internal/dto"

type MessageHandler interface {
	HandleMessage(from *dto.User, body string)
}

type MessageHandlerFunc func(from *dto.User, body string)

func (f MessageHandlerFunc) HandleMessage(from *dto.User, body string) {
	f(from, body)
}

type MessageSender interface {
	SendMessage(to *dto.User, text string) error
}

type MessageSenderFunc func(from *dto.User, body string) error

func (f MessageSenderFunc) SendMessage(to *dto.User, body string) error {
	return f(to, body)
}
