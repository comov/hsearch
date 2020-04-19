package bot

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/comov/hsearch/bot/settings"
	"github.com/comov/hsearch/structs"

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
	chat, err := b.storage.ReadChat(query.Message.Chat.ID)
	if err != nil {
		log.Println("[settingsCallback.ReadChat] error:", err)
		_, err = b.bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Прости, говнокод сломался"))
		if err != nil {
			log.Println("[settingsCallback.Send.error] error:", err)
		}
		return
	}

	_, err = b.bot.Send(settings.MainSettingsHandler(query.Message, chat))
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
	chat, err := b.storage.ReadChat(query.Message.Chat.ID)
	if err != nil {
		log.Println("[settingsCallback.ReadChat] error:", err)
		_, err = b.bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Прости, говнокод сломался"))
		if err != nil {
			log.Println("[settingsCallback.Send.error] error:", err)
		}
		return
	}

	switch query.Data {
	case "searchOn":
		chat.Enable = true
	case "searchOff":
		chat.Enable = false
	case "dieselOn":
		chat.Diesel = true
	case "dieselOff":
		chat.Diesel = false
	case "lalafoOn":
		chat.Lalafo = true
	case "lalafoOff":
		chat.Lalafo = false
	}

	err = b.storage.UpdateSettings(chat)
	if err != nil {
		log.Println("[settingsCallback.ReadChat] error:", err)
		_, err = b.bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Прости, говнокод сломался"))
		if err != nil {
			log.Println("[settingsCallback.Send.error] error:", err)
		}
		return
	}

	_, err = b.bot.Send(settings.MainSearchHandler(query.Message, chat))
	if err != nil {
		log.Println("[searchCallback.Send] error:", err)
	}
}

//// buttons for filters
func (b *Bot) filtersCallback(query *tgbotapi.CallbackQuery) {
	chat, err := b.storage.ReadChat(query.Message.Chat.ID)
	if err != nil {
		log.Println("[settingsCallback.ReadChat] error:", err)
		_, err = b.bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Прости, говнокод сломался"))
		if err != nil {
			log.Println("[settingsCallback.Send.error] error:", err)
		}
		return
	}

	_, err = b.bot.Send(settings.MainFiltersHandler(query.Message, chat))
	if err != nil {
		log.Println("[filtersCallback.Send] error:", err)
	}
}

func (b *Bot) withPhotoCallback(query *tgbotapi.CallbackQuery) {
	chat, err := b.storage.ReadChat(query.Message.Chat.ID)
	if err != nil {
		log.Println("[settingsCallback.ReadChat] error:", err)
		_, err = b.bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Прости, говнокод сломался"))
		if err != nil {
			log.Println("[settingsCallback.Send.error] error:", err)
		}
		return
	}

	switch query.Data {
	case "withPhotoOn":
		chat.Photo = true
	case "withPhotoOff":
		chat.Photo = false
	}

	err = b.storage.UpdateSettings(chat)
	if err != nil {
		log.Println("[settingsCallback.ReadChat] error:", err)
		_, err = b.bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Прости, говнокод сломался"))
		if err != nil {
			log.Println("[settingsCallback.Send.error] error:", err)
		}
		return
	}

	_, err = b.bot.Send(settings.MainFiltersHandler(query.Message, chat))
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

	chat, err := b.storage.ReadChat(message.Chat.ID)
	if err != nil {
		log.Println("[settingsCallback.ReadChat] error:", err)
		_, err = b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Прости, говнокод сломался"))
		if err != nil {
			log.Println("[settingsCallback.Send.error] error:", err)
		}
		return
	}

	switch a.currency {
	case "USD":
		chat.USD = structs.Price{from, to}
	case "KGS":
		chat.KGS = structs.Price{from, to}
	}

	err = b.storage.UpdateSettings(chat)
	if err != nil {
		log.Println("[settingsCallback.ReadChat] error:", err)
		_, err = b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Прости, говнокод сломался"))
		if err != nil {
			log.Println("[settingsCallback.Send.error] error:", err)
		}
		return
	}

	b.clearRetry(message.Chat, message.MessageID)
}
