package bot

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/comov/hsearch/bot/settings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// answer - when we ask a question, we expect an answer. This structure stores
//  the wrong answers for deletion and the menu to which you want to go back.
type answer struct {
	deadline  time.Time
	callback  func(*tgbotapi.Message, answer)
	currency  string
	maxErrors int
	menuId    int
	messages  []int
}

const (
	waitSeconds = 20
	maxErrors   = 4
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

func (b *Bot) priceCallback(query *tgbotapi.CallbackQuery) {
	_, err := b.bot.Send(settings.FilterPriceHandler(query.Message, query.Data))
	if err != nil {
		log.Println("[priceCallback.Send] error:", err)
		return
	}

	b.addWaitCallback(query.Message.Chat.ID, answer{
		deadline:  time.Now().Add(time.Second * waitSeconds),
		callback:  b.priceWaiterCallback,
		currency:  query.Data,
		menuId:    query.Message.MessageID,
		maxErrors: maxErrors,
	})
}

// priceWaiterCallback - process a response from the user
func (b *Bot) priceWaiterCallback(message *tgbotapi.Message, a answer) {
	prices := strings.Split(message.Text, "-")
	if len(prices) < 2 {
		b.wrongAnswer(message, a)
		return
	}

	from, err := strconv.Atoi(strings.TrimSpace(prices[0]))
	if err != nil {
		b.wrongAnswer(message, a)
		log.Println("[priceWaiterCallback] error:", err)
		return
	}

	to, err := strconv.Atoi(strings.TrimSpace(prices[1]))
	if err != nil {
		b.wrongAnswer(message, a)
		log.Println("[priceWaiterCallback] error:", err)
		return
	}

	userSettings := settings.MockStorage[message.Chat.UserName]
	userSettings[a.currency] = [2]int{from, to}
	settings.MockStorage[message.Chat.UserName] = userSettings

	b.clearRetry(message.Chat, message.MessageID)
}
