package tg

import (
	"strings"

	"taskbot/internal/dto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Sender struct {
	Bot *tgbotapi.BotAPI
}

func (s Sender) SendMessage(to *dto.User, text string) error {
	msg := tgbotapi.NewMessage(int64(to.CharID), strings.TrimSpace(text))
	_, err := s.Bot.Send(msg)
	return err
}
