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
	msgText := fmt.Sprintf(mainSearchText,
		yesNo(chat.Enable),
		yesNo(chat.Diesel),
		yesNo(chat.Lalafo),
	)

	message := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, msgText)
	message.ReplyMarkup = getSearchKeyboard(
		chat.Enable,
		chat.Diesel,
		chat.Lalafo,
	)
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}

func getSearchKeyboard(search, diesel, lalafo bool) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText(
				"Включить поиск", "searchOn",
				search,
				"Выключить поиск", "searchOff",
			)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getButtonText(
				"Искать на diesel", "dieselOn",
				diesel,
				"Не искать на diesel", "dieselOff",
			)),
			tgbotapi.NewInlineKeyboardButtonData(getButtonText(
				"Искать на lalafo", "lalafoOn",
				lalafo,
				"Не искать на lalafo", "lalafoOff",
			)),
		),
		backRow,
	)
	return &keyboard
}

func getButtonText(t1, d1 string, operator bool, t2, d2 string) (string, string) {
	if operator {
		return t2, d2
	}
	return t1, d1
}
