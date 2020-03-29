package settings

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	search = tgbotapi.NewInlineKeyboardButtonData("Поиск", "search")
)

func MainSearchHandler(msg *tgbotapi.Message) tgbotapi.Chattable {
	// todo: load from DB
	msgText := getSearchText(
		MockStorage[msg.Chat.UserName]["searchEnable"].(bool),
	)

	message := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, msgText)
	message.ReplyMarkup = getSearchKeyboard(
		MockStorage[msg.Chat.UserName]["searchEnable"].(bool),
	)
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}

func getSearchText(search bool) string {
	return fmt.Sprintf(mainSearchText, yesNo(search))
}

func getSearchKeyboard(search bool) *tgbotapi.InlineKeyboardMarkup {
	text := "Включить поиск"
	data := "searchOn"
	if search {
		text = "Выключить поиск"
		data = "searchOff"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(text, data),
		),
		backRow,
	)
	return &keyboard
}
