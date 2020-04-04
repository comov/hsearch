package settings

import (
	"fmt"

	"github.com/comov/hsearch/structs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// BackFlowMap - a list to register all menus with a "Back" button. If you
//  press the back button, the key will be searched in the text and the menu
//  will be called up by value.
var BackFlowMap = map[string]string{
	"Фильтры поиска":   "settings",
	"Настройки поиска": "settings",
	"Укажите суммы в":  "filters",
}

// todo: remove mocks
var MockStorage = map[string]map[string]interface{}{
	"aastashov": {
		"searchEnable": true,
		"withPhoto":    true,
		"KGS":     [2]int{10000, 22000},
		"USD":     [2]int{250, 350},
	},
	"stanislav_dev": {
		"searchEnable": true,
		"withPhoto":    true,
		"KGS":     [2]int{10000, 22000},
		"USD":     [2]int{250, 350},
	},
}

// buttons for configs
var (
	back    = tgbotapi.NewInlineKeyboardButtonData("<< назад", "back")
	backRow = tgbotapi.NewInlineKeyboardRow(back)

	mainSettings = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(search, filters),
	)
)

func MainSettingsHandler(msg *tgbotapi.Message, chat *structs.Chat) tgbotapi.Chattable {
	msgText := fmt.Sprintf(mainSettingsText,
		yesNo(chat.Enable),
		yesNo(chat.Photo),
		price(chat.KGS),
		price(chat.USD),
	)

	if msg.IsCommand() {
		message := tgbotapi.NewMessage(msg.Chat.ID, msgText)
		message.ReplyMarkup = &mainSettings
		message.ParseMode = tgbotapi.ModeMarkdown
		return message
	}

	message := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, msgText)
	message.ReplyMarkup = &mainSettings
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}