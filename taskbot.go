package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"taskbot/internal/delivery"
	"taskbot/internal/delivery/tg"
	"taskbot/internal/dto"
	"taskbot/internal/repo"
	"taskbot/internal/router"
	"taskbot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	WebhookURL string
	BotToken   string
)

func debugHandler(next delivery.MessageHandler) delivery.MessageHandlerFunc {
	return func(user *dto.User, body string) {
		log.Printf("<= [@%s] %s", user.Username, body)
		next.HandleMessage(user, body)
	}
}

func debugSender(next delivery.MessageSender) delivery.MessageSenderFunc {
	return func(user *dto.User, body string) error {
		log.Printf("=> [@%s] %s", user.Username, strings.TrimSpace(body))
		err := next.SendMessage(user, body)
		return err
	}
}

func startTaskBot(ctx context.Context, httpListenAddr string) error {
	const op = "startTaskBot"

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return fmt.Errorf("%s: can't connect: %w", op, err)
	}

	// bot.Debug = true
	log.Printf("%s: authorized on account @%s", op, bot.Self.UserName)

	// sender := tg.Sender{Bot: bot}
	sender := debugSender(tg.Sender{Bot: bot})

	s := &service.Service{
		Repo:   repo.New(),
		Sender: sender,
	}

	ro := &router.Router{
		Sender: sender,

		Usage: func() string {
			return `Usage:
	/tasks - выводит список всех задач
	/new XXX YYY ZZZ - создаёт новую задачу
	/assign_$ID - делаеть пользователя исполнителем задачи
	/unassign_$ID - снимает задачу с текущего исполнителя
	/resolve_$ID - выполняет задачу, удаляет её из списка
	/my - показывает задачи, которые назначены на меня
	/owner - показывает задачи которые были созданы мной`
		},

		Commands: map[string]delivery.MessageHandler{
			"tasks":    delivery.MessageHandlerFunc(s.ListAllTasks),
			"new":      delivery.MessageHandlerFunc(s.CreateTask),
			"assign":   delivery.MessageHandlerFunc(s.AssignTask),
			"unassign": delivery.MessageHandlerFunc(s.UnassignTask),
			"resolve":  delivery.MessageHandlerFunc(s.ResolveTask),
			"my":       delivery.MessageHandlerFunc(s.ListAssigneeTasks),
			"owner":    delivery.MessageHandlerFunc(s.ListOwnerTasks),
		},
	}

	svr := tg.Server{
		Bot: bot,
		// Handler: ro,
		Handler:    debugHandler(ro),
		Concurency: 2,
	}

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// Block until a signal is received.
		s := <-c
		fmt.Printf("%s: Got signal: %v", op, s)
		cancel()
	}()

	return svr.ListenAndServe(ctx, httpListenAddr, WebhookURL)
}

// это заглушка чтобы импорт сохранился
func __dummy() {
	tgbotapi.APIEndpoint = "_dummy"
}
