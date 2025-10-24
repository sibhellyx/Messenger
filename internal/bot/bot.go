package bot

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"runtime/debug"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type AuthServiceInterface interface {
	Activate(tgName string) error
}

type Bot struct {
	Api               *tgbotapi.BotAPI
	actions           map[string]ActionFunc // for handle start action
	RegistratingUsers map[string]string
	Service           AuthServiceInterface
}

type ActionFunc func(ctx context.Context, bot *Bot, update *tgbotapi.Update) error

func NewBot(api *tgbotapi.BotAPI, service AuthServiceInterface) *Bot {
	return &Bot{
		Api:               api,
		Service:           service,
		RegistratingUsers: make(map[string]string),
	}
}

func (b *Bot) RegisterAction(nameAction string, action ActionFunc) {
	if b.actions == nil {
		b.actions = make(map[string]ActionFunc)
	}
	b.actions[nameAction] = action
}

func SendUserToRegister() {

}

func (b *Bot) GetLinkForFinishRegister(tgName string) string {
	slog.Debug("sending link for user", "tgName", tgName)
	token := uuid.New().String()
	b.RegistratingUsers[token] = tgName
	return fmt.Sprintf("https://t.me/%s?start=invite_%s", b.Api.Self.UserName, token)
}

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.Api.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			updateCtx, updateCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			b.handleUpdate(updateCtx, update)
			updateCancel()
		case <-ctx.Done():
			return ctx.Err()
			// add case for registration
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("[ERROR] panic recovered: %v\n%s", p, string(debug.Stack()))
		}
	}()

	var action ActionFunc

	cmd := update.Message.Command()

	actionView, ok := b.actions[cmd]

	if !ok {
		return
	}

	action = actionView
	if err := action(ctx, b, &update); err != nil {
		log.Printf("[ERROR] failed to execute action: %v", err)

		if _, err := b.Api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Internal error")); err != nil {
			log.Printf("[ERROR] failed to send error message: %v", err)
		}
	}

}
