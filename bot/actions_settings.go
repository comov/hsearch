package bot

import (
	"log"
	"strings"

	"github.com/comov/hsearch/bot/settings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//// buttons for configs
// settingsCallback - show all settings for user
func (b *Bot) settingsCallback(query *tgbotapi.CallbackQuery) {
	_, err := b.bot.Send(settings.MainSettingsHandler(query.Message))
	if err != nil {
		log.Println("[settingsCallback.Send] error:", err)
	}
}

// backCallback - navigation button in settings menu
func (b *Bot) backCallback(query *tgbotapi.CallbackQuery) {
	for text, key := range settings.BackFlowMap {
		if strings.Contains(query.Message.Text, text) {
			b.callbacks[key](query)
			return
		}
	}
	log.Println("[backCallback.Send] not found key for text:", query.Message.Text)
}

// searchCallback - change search parameters for bot
func (b *Bot) searchCallback(query *tgbotapi.CallbackQuery) {
	userSettings := settings.MockStorage[query.Message.Chat.UserName]
	switch query.Data {
	case "searchOn":
		userSettings["searchEnable"] = true
	case "searchOff":
		userSettings["searchEnable"] = false
	}
	settings.MockStorage[query.Message.Chat.UserName] = userSettings

	_, err := b.bot.Send(settings.MainSearchHandler(query.Message))
	if err != nil {
		log.Println("[searchCallback.Send] error:", err)
	}
}

//// buttons for filters
func (b *Bot) filtersCallback(query *tgbotapi.CallbackQuery) {
	_, err := b.bot.Send(settings.MainFiltersHandler(query.Message))
	if err != nil {
		log.Println("[filtersCallback.Send] error:", err)
	}
}

func (b *Bot) withPhotoCallback(query *tgbotapi.CallbackQuery) {
	userSettings := settings.MockStorage[query.Message.Chat.UserName]
	switch query.Data {
	case "withPhotoOn":
		userSettings["withPhoto"] = true
	case "withPhotoOff":
		userSettings["withPhoto"] = false
	}
	settings.MockStorage[query.Message.Chat.UserName] = userSettings

	_, err := b.bot.Send(settings.MainFiltersHandler(query.Message))
	if err != nil {
		log.Println("[withPhotoCallback.Send] error:", err)
	}
}

func (b *Bot) priceKGSCallback(query *tgbotapi.CallbackQuery) {
	_, err := b.bot.Send(settings.FilterPriceKGSHandler(query.Message))
	if err != nil {
		log.Println("[priceKGSCallback.Send] error:", err)
	}

	// todo: add listener on one minute for receive new messages. If user send
	//  not valid message, we send "Read example, and send valid price range!".
	//  If user send 3 not valid message (without price or bad format), we send
	//  message we send mail settings menu and remove listener.
	//  If user send success price range, we save in settings and remove
	//  listener.
}

func (b *Bot) priceUSDCallback(query *tgbotapi.CallbackQuery) {
	_, err := b.bot.Send(settings.FilterPriceUSDHandler(query.Message))
	if err != nil {
		log.Println("[priceUSDCallback.Send] error:", err)
	}

	// todo: add listener on one minute for receive new messages. If user send
	//  not valid message, we send "Read example, and send valid price range!".
	//  If user send 3 not valid message (without price or bad format), we send
	//  message we send mail settings menu and remove listener.
	//  If user send success price range, we save in settings and remove
	//  listener.
}
