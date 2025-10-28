package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"runtime/debug"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type AuthServiceInterface interface {
	Activate(tgName string) (uint, error)
	GetTokenFromRedis(token string) (string, error)
	DeleteRegistrationTokenFromRedis(token string) error
	SaveUserRegistration(userId uint, tgChatId int64) error
	GetUserRegistration(userID uint) (int64, error)
}

type Bot struct {
	Api     *tgbotapi.BotAPI
	actions map[string]ActionFunc // for handle start action
	Service AuthServiceInterface
}

type ActionFunc func(ctx context.Context, bot *Bot, update *tgbotapi.Update) error

func NewBot(api *tgbotapi.BotAPI, service AuthServiceInterface) *Bot {
	return &Bot{
		Api:     api,
		Service: service,
	}
}

func (b *Bot) RegisterAction(nameAction string, action ActionFunc) {
	if b.actions == nil {
		b.actions = make(map[string]ActionFunc)
	}
	b.actions[nameAction] = action
}

func (b *Bot) GetLinkForFinishRegister(tgName string) (string, string) {
	slog.Debug("sending link for user", "tgName", tgName)
	token := uuid.New().String()
	return token, fmt.Sprintf("https://t.me/%s?start=invite_%s", b.Api.Self.UserName, token)
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

func (b *Bot) SendCode(code string, userId uint) error {
	// Получаем chat ID пользователя из Redis
	chatID, err := b.Service.GetUserRegistration(userId)
	if err != nil {
		slog.Error("failed to get user chat ID", "user_id", userId, "error", err)
		return errors.New("failed find user, not active profile")
	}

	message := fmt.Sprintf(
		"🔐 *Код подтверждения*\n\n"+
			"Ваш код для входа: `%s`\n\n"+
			"⚠️ *Никому не сообщайте этот код!*\n"+
			"⏳ Код действителен в течение 10 минут",
		code,
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"

	_, err = b.Api.Send(msg)
	if err != nil {
		slog.Error("failed to send verification code",
			"error", err, "user_id", userId, "chat_id", chatID, "code", code)
		return fmt.Errorf("failed to send verification code to user %d (chat %d): %w", userId, chatID, err)
	}

	slog.Info("verification code sent successfully",
		"user_id", userId, "chat_id", chatID, "code", code)
	return nil
}
