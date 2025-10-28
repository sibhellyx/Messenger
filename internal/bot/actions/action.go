package actions

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sibhellyx/Messenger/internal/bot"
)

func HandleStart() bot.ActionFunc {
	return func(ctx context.Context, bot *bot.Bot, update *tgbotapi.Update) error {
		slog.Debug("start action han")
		if update.Message == nil {
			return nil
		}

		commandArgs := update.Message.CommandArguments()

		if !strings.HasPrefix(commandArgs, "invite_") {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать! Для регистрации используйте ссылку-приглашение.")
			_, err := bot.Api.Send(msg)
			return err
		}

		token := strings.TrimPrefix(commandArgs, "invite_")

		tgName, err := bot.Service.GetTokenFromRedis(token)
		if err != nil {
			slog.Error("failed find user by token in redis repo", "error", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная или устаревшая ссылка приглашения. Запросите новую ссылку.")
			_, err := bot.Api.Send(msg)
			return err
		}

		id, err := bot.Service.Activate(tgName)
		if err != nil {
			slog.Error("failed to activate user", "error", err, "tgname", tgName)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка активации аккаунта. Обратитесь к администратору.")
			_, sendErr := bot.Api.Send(msg)
			if sendErr != nil {
				return fmt.Errorf("activation error: %v, send error: %v", err, sendErr)
			}
			return err
		}

		err = bot.Service.DeleteRegistrationTokenFromRedis(token)
		if err != nil {
			slog.Error("failed delete token from redis repo", "error", err)
			return err
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Ваш аккаунт успешно активирован! Теперь вы можете войти в систему.")
		_, err = bot.Api.Send(msg)
		if err != nil {
			return err
		}

		bot.Service.SaveUserRegistration(id, update.Message.Chat.ID)

		slog.Info("user activated successfully via bot", "tgname", tgName, "chat_id", update.Message.Chat.ID)
		return nil
	}
}
