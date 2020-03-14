package bots

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) start(_ *tgbotapi.Message) string {
	return startMessage
}

func (b *Bot) stop(message *tgbotapi.Message) string {
	err := b.st.StopSearch(message.Chat.ID)
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

	err := b.st.StartSearch(m.Chat.ID, m.Chat.UserName, title, m.Chat.Type)
	if err != nil {
		log.Println("[search.StartSearchForChat] error:", err)
		return "Прости, говнокод сломался"
	}

	return "Теперь я буду искать для тебя квартиры"
}

func (b *Bot) feedback(message *tgbotapi.Message) string {
	if message.CommandArguments() == "" {
		return "Нужно оставить комментарий в виде:\n /feedback Я тебя найду...."
	}

	err := b.st.Feedback(message.Chat.ID, message.Chat.UserName, message.CommandArguments())
	if err != nil {
		log.Println("[feedback.StartSearchForChat] error:", err)
		return "Прости, даже фидбек может быть сломан"
	}
	return "Понял, предам!"
}
