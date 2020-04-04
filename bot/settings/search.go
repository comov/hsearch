package settings

import (
	"fmt"

	"github.com/comov/hsearch/structs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	search = tgbotapi.NewInlineKeyboardButtonData("Поиск", "search")
)

func MainSearchHandler(msg *tgbotapi.Message, chat *structs.Chat) tgbotapi.Chattable {
	msgText := getSearchText(
		chat.Enable,
	)

	message := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, msgText)
	message.ReplyMarkup = getSearchKeyboard(
		chat.Enable,
	)
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}

func getSearchText(v bool) string {
	return fmt.Sprintf(mainSearchText, yesNo(v))
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
