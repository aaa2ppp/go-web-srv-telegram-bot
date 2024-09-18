package tg

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"taskbot/internal/delivery"
	"taskbot/internal/dto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Server struct {
	Bot        *tgbotapi.BotAPI
	Handler    delivery.MessageHandler
	Concurency int
}

func (s Server) ListenAndServe(ctx context.Context, httpListenAddr, WebhookURL string) error {
	const op = "Server.ListenAndServe"

	var updates tgbotapi.UpdatesChannel
	if httpListenAddr != "" {

		// for tests only!
		s.Bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
		updates = s.Bot.ListenForWebhook("/")
		go http.ListenAndServe(httpListenAddr, nil)

	} else {

		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		var err error
		updates, err = s.Bot.GetUpdatesChan(u)
		if err != nil {
			return fmt.Errorf("%s: can't get update chennel: %w", op, err)
		}

		// // Optional: wait for updates and clear them if you don't want to handle
		// // a large backlog of old messages
		// time.Sleep(time.Millisecond * 500)
		// updates.Clear()
	}

	var wg sync.WaitGroup
	for i := 1; i < s.Concurency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.updateLoop(ctx, updates)
		}()
	}

	s.updateLoop(ctx, updates)
	wg.Wait()

	return nil
}

func (s Server) updateLoop(ctx context.Context, updates tgbotapi.UpdatesChannel) {
loop:
	for {
		select {
		case <-ctx.Done():
			return

		case update, ok := <-updates:
			if !ok {
				return
			}

			msg := update.Message
			if msg == nil {
				continue loop
			}

			if s.Handler == nil {
				resp := tgbotapi.NewMessage(msg.Chat.ID, "You say: "+msg.Text)
				s.Bot.Send(resp)
				continue loop
			}

			from := &dto.User{
				ID:       uint64(msg.From.ID),
				Username: msg.From.UserName,
				CharID:   uint64(msg.Chat.ID),
			}

			s.Handler.HandleMessage(from, msg.Text)
		}
	}
}
