package settings

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// buttons for filters
var (
	filters = tgbotapi.NewInlineKeyboardButtonData("Фильтры", "filters")

	pricesRow = tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Цена в KGS", "priceKGS"),
		tgbotapi.NewInlineKeyboardButtonData("Цена в USD", "priceUSD"),
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
		MockStorage[msg.Chat.UserName]["priceKGS"],
		MockStorage[msg.Chat.UserName]["priceUSD"],
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

func FilterPriceKGSHandler(msg *tgbotapi.Message) tgbotapi.Chattable {
	message := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, textKGS)
	message.ReplyMarkup = &priceBack
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}

func FilterPriceUSDHandler(msg *tgbotapi.Message) tgbotapi.Chattable {
	message := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, textUSD)
	message.ReplyMarkup = &priceBack
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}
