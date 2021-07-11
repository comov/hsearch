package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v4"
)

const feedBackWait = time.Minute * 5

func (b *Bot) help(_ *tgbotapi.Message) string {
	return helpMessage
}

func (b *Bot) start(ctx context.Context, chat *tgbotapi.Chat) string {
	_, err := b.storage.ReadChat(ctx, chat.ID)
	if err == nil {
		return "Я уже работаю на тебя"
	}

	if err != pgx.ErrNoRows {
		sentry.CaptureException(err)
		log.Println("[start.ReadChat] error:", err)
		return "Что-то сломалось. Со мной такое впервые... 🤔"
	}

	title := chat.Title
	if chat.IsPrivate() {
		title = fmt.Sprintf("%s %s", chat.FirstName, chat.LastName)
	}

	err = b.storage.CreateChat(ctx, chat.ID, chat.UserName, title, chat.Type, true)
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[start.CreateChat] error:", err)
		return "Что-то сломалось. Со мной такое впервые... 🤔"
	}
	return "Теперь я буду искать для тебя квартиры"
}

func (b *Bot) stop(ctx context.Context, chat *tgbotapi.Chat) string {
	err := b.storage.DeleteChat(ctx, chat.ID)
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[stop.DeleteChat] error:", err)
		return "Что-то сломалось. Со мной такое впервые... 🤔"
	}
	return "Я больше не буду искать для тебя квартиры"
}

func (b *Bot) feedback(_ context.Context, chat *tgbotapi.Chat) string {
	b.addWaitCallback(chat.ID, answer{
		deadline: time.Now().Add(feedBackWait),
		callback: b.feedbackWaiterCallback,
	})
	return feedbackText
}

func (b *Bot) feedbackWaiterCallback(ctx context.Context, message *tgbotapi.Message, _ answer) {
	msgText := "Понял, передам!"
	err := b.storage.Feedback(ctx, message.Chat.ID, message.Chat.UserName, message.Text)
	if err != nil {
		log.Println("[feedbackWaiterCallback.Feedback] error:", err)
		msgText = "Прости, даже фидбек может быть сломан"
		sentry.CaptureException(err)
	}

	_, err = b.Send(tgbotapi.NewMessage(message.Chat.ID, msgText))
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[feedbackWaiterCallback.Send] error:", err)
	}

	if b.adminChatId != 0 {
		_, err = b.Send(tgbotapi.NewMessage(b.adminChatId, getFeedbackAdminText(message.Chat, message.Text)))
		if err != nil {
			sentry.CaptureException(err)
			log.Println("[feedbackWaiterCallback.Send2] error:", err)
		}
	}
}
