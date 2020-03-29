package settings

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// buttons for filters
var (
	filters = tgbotapi.NewInlineKeyboardButtonData("Фильтры", "filters")

	pricesRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Цена в KGS", "KGS"),
		tgbotapi.NewInlineKeyboardButtonData("Цена в USD", "USD"),
	)

	priceBack = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			backRow,
		},
	}
)

func MainFiltersHandler(msg *tgbotapi.Message) tgbotapi.Chattable {
	// todo: load from DB
	msgText := fmt.Sprintf(mainFiltersText,
		yesNo(MockStorage[msg.Chat.UserName]["withPhoto"].(bool)),
		price(MockStorage[msg.Chat.UserName]["KGS"].([2]int)),
		price(MockStorage[msg.Chat.UserName]["USD"].([2]int)),
	)

	message := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, msgText)
	message.ReplyMarkup = getFiltersKeyboard(
		MockStorage[msg.Chat.UserName]["withPhoto"].(bool),
	)
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}

func getFiltersKeyboard(photo bool) *tgbotapi.InlineKeyboardMarkup {
	text := "Только с фото"
	data := "withPhotoOn"
	if photo {
		text = "Можно и без фото"
		data = "withPhotoOff"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(text, data),
		),
		pricesRow,
		backRow,
	)
	return &keyboard
}

func FilterPriceHandler(msg *tgbotapi.Message, currency string) tgbotapi.Chattable {
	msgText := ""
	switch currency {
	case "USD":
		msgText = textUSD
	case "KGS":
		msgText = textKGS
	}

	message := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, msgText)
	message.ReplyMarkup = &priceBack
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}
