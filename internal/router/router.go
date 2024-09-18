package router

import (
	"strings"
	"unicode"

	"taskbot/internal/delivery"
	"taskbot/internal/dto"
)

func isAlphaNum(c byte) bool {
	return unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c))
}

type Router struct {
	Commands map[string]delivery.MessageHandler
	Usage    func() string
	Sender   delivery.MessageSender
}

func (r Router) HandleMessage(from *dto.User, text string) {

	if len(text) == 0 || text[0] != '/' {
		r.usage(from, "command must start with /")
		return
	}
	text = text[1:]

	p := 0
	for p < len(text) && isAlphaNum(text[p]) {
		p++
	}

	cmd := text[:p]
	if cmd == "" {
		r.usage(from, "command cannot be empty")
		return
	}

	if p < len(text) {
		text = strings.TrimSpace(text[p+1:])
	} else {
		text = ""
	}

	for k, v := range r.Commands {
		if k == cmd {
			v.HandleMessage(from, text)
			return
		}
	}

	r.usage(from, "unknown command: "+cmd)
}

func (r Router) usage(to *dto.User, msg string) {

	if r.Sender == nil {
		return
	}

	buf := &strings.Builder{}
	buf.WriteString(msg)

	if r.Usage != nil {
		buf.WriteByte('\n')
		buf.WriteString(r.Usage())
	}

	r.Sender.SendMessage(to, buf.String())
}
