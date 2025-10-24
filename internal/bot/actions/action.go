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
		slog.Debug("start action handling")
		if update.Message == nil {
			return nil
		}

		// Получаем аргументы команды /start
		commandArgs := update.Message.CommandArguments()

		if !strings.HasPrefix(commandArgs, "invite_") {
			// Если это обычная команда /start без инвайта
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать! Для регистрации используйте ссылку-приглашение.")
			_, err := bot.Api.Send(msg)
			return err
		}

		// Извлекаем токен из аргументов
		token := strings.TrimPrefix(commandArgs, "invite_")

		// Ищем tgName по токену в карте регистрирующихся пользователей
		tgName, exists := bot.RegistratingUsers[token]
		if !exists {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная или устаревшая ссылка приглашения. Запросите новую ссылку.")
			_, err := bot.Api.Send(msg)
			return err
		}

		// Активируем пользователя через сервис
		err := bot.Service.Activate(tgName)
		if err != nil {
			slog.Error("failed to activate user", "error", err, "tgname", tgName)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка активации аккаунта. Обратитесь к администратору.")
			_, sendErr := bot.Api.Send(msg)
			if sendErr != nil {
				return fmt.Errorf("activation error: %v, send error: %v", err, sendErr)
			}
			return err
		}

		// Удаляем использованный токен из карты
		delete(bot.RegistratingUsers, token)

		// Отправляем сообщение об успешной активации
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Ваш аккаунт успешно активирован! Теперь вы можете войти в систему.")
		_, err = bot.Api.Send(msg)
		if err != nil {
			return err
		}

		slog.Info("user activated successfully via bot", "tgname", tgName, "chat_id", update.Message.Chat.ID)
		return nil
	}
}
