package bot

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const feedBackWait = time.Minute * 5

func (b *Bot) start(_ *tgbotapi.Message) string {
	return startMessage
}

func (b *Bot) stop(message *tgbotapi.Message) string {
	err := b.storage.StopSearch(message.Chat.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			first := "Этой группы"
			second := "и так ничего сюда"
			if message.Chat.IsPrivate() {
				first = "Тебя"
				second = "тебе и так ничего"
			}
			return fmt.Sprintf(stopNotFound, first, second)
		}

		log.Println("[stop.StopSearch] error:", err)
		return "Прости, говнокод сломался"
	}

	return "Ok! Я больше не буду отправлять тебе квартиры"
}

func (b *Bot) search(m *tgbotapi.Message) string {
	title := m.Chat.Title
	if m.Chat.IsPrivate() {
		title = fmt.Sprintf("%s %s", m.Chat.FirstName, m.Chat.LastName)
	}

	err := b.storage.StartSearch(m.Chat.ID, m.Chat.UserName, title, m.Chat.Type)
	if err != nil {
		log.Println("[search.StartSearchForChat] error:", err)
		return "Прости, говнокод сломался"
	}

	return "Теперь я буду искать для тебя квартиры"
}

func (b *Bot) feedback(message *tgbotapi.Message) string {
	b.addWaitCallback(message.Chat.ID, answer{
		deadline: time.Now().Add(feedBackWait),
		callback: b.feedbackWaiterCallback,
	})
	return feedbackText
}

func (b *Bot) feedbackWaiterCallback(message *tgbotapi.Message, _ answer) {
	msgText := "Понял, предам!"
	err := b.storage.Feedback(message.Chat.ID, message.Chat.UserName, message.Text)
	if err != nil {
		log.Println("[feedbackWaiterCallback.Feedback] error:", err)
		msgText = "Прости, даже фидбек может быть сломан"
	}

	_, err = b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, msgText))
	if err != nil {
		log.Println("[feedbackWaiterCallback.Send] error:", err)
	}

	if b.adminChatId != 0 {
		_, err = b.bot.Send(tgbotapi.NewMessage(
			b.adminChatId,
			getFeedbackAdminText(message.Chat, message.Text),
		))
		if err != nil {
			log.Println("[feedbackWaiterCallback.Send2] error:", err)
		}
	}
}
