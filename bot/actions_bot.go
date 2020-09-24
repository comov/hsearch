package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v4"
)

const feedBackWait = time.Minute * 5

func (b *Bot) help(_ *tgbotapi.Message) string {
	return helpMessage
}

func (b *Bot) start(ctx context.Context, m *tgbotapi.Message) string {
	_, err := b.storage.ReadChat(ctx, m.Chat.ID)
	if err == nil {
		return "Я уже работаю на тебя"
	}

	if err != pgx.ErrNoRows {
		log.Println("[start.ReadChat] error:", err)
		return "Что-то сломалось. Со мной такое впервые... 🤔"
	}

	title := m.Chat.Title
	if m.Chat.IsPrivate() {
		title = fmt.Sprintf("%s %s", m.Chat.FirstName, m.Chat.LastName)
	}
	err = b.storage.CreateChat(ctx, m.Chat.ID, m.Chat.UserName, title, m.Chat.Type)
	if err != nil {
		log.Println("[start.StopSearch] error:", err)
		return "Что-то сломалось. Со мной такое впервые... 🤔"
	}
	return "Теперь я буду искать для тебя квартиры"
}

func (b *Bot) stop(ctx context.Context, m *tgbotapi.Message) string {
	err := b.storage.DeleteChat(ctx, m.Chat.ID)
	if err != nil {
		return "Что-то сломалось. Со мной такое впервые... 🤔"
	}
	return "Я больше не буду искать для тебя квартиры"
}

func (b *Bot) feedback(_ context.Context, message *tgbotapi.Message) string {
	b.addWaitCallback(message.Chat.ID, answer{
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
	}

	_, err = b.Send(tgbotapi.NewMessage(message.Chat.ID, msgText))
	if err != nil {
		log.Println("[feedbackWaiterCallback.Send] error:", err)
	}

	if b.adminChatId != 0 {
		_, err = b.Send(tgbotapi.NewMessage(b.adminChatId, getFeedbackAdminText(message.Chat, message.Text)))
		if err != nil {
			log.Println("[feedbackWaiterCallback.Send2] error:", err)
		}
	}
}
