package settings

import (
	"fmt"

	"github.com/comov/hsearch/structs"

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

func MainFiltersHandler(msg *tgbotapi.Message, chat *structs.Chat) tgbotapi.Chattable {
	msgText := fmt.Sprintf(mainFiltersText,
		yesNo(chat.Photo),
		price(chat.KGS),
		price(chat.USD),
	)

	message := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, msgText)
	message.ReplyMarkup = getFiltersKeyboard(
		chat.Photo,
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
