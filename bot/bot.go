package bot

import (
	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/structs"
	"log"
	"sync"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type (
	Storage interface {
		StartSearch(id int64, username, title, chatType string) error
		StopSearch(id int64) error
		Dislike(msgId int, chatId int64) ([]int, error)
		Skip(msgId int, chatId int64) error
		Feedback(chat int64, username, body string) error

		SaveMessage(msgId int, offerId uint64, chat int64, kind string) error
		ReadOfferDescription(msgId int, chatId int64) (uint64, string, error)
		ReadOfferImages(msgId int, chatId int64) (uint64, []string, error)
		ReadNextOffer(chatId int64) (*structs.Offer, error)
	}

	callback func(query *tgbotapi.CallbackQuery)

	Bot struct {
		bot       *tgbotapi.BotAPI
		storage   Storage
		callbacks map[string]callback

		adminChatId int64

		// {time.Minutes * 3, b.callbackName(message *tgbotapi.Message)}
		waitAnswers map[int64]answer
		waitMutex   sync.Mutex
	}
)

func NewTelegramBot(cnf *configs.Config, st Storage) *Bot {
	bot, err := tgbotapi.NewBotAPI(cnf.TelegramToken)
	if err != nil {
		log.Fatalln("[NewBot.NewBotAPI] error: ", err)
		return nil
	}

	bb := &Bot{
		bot:         bot,
		storage:     st,
		adminChatId: cnf.AdminChatId,
		callbacks:   make(map[string]callback, 0),
		waitAnswers: make(map[int64]answer),
	}

	bb.registerCallbacks()
	return bb
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalln("[Start.GetUpdatesChan] error: ", err)
		return
	}

	log.Println("[bot] Start listen Telegram chanel")
	for update := range updates {
		if update.CallbackQuery != nil {
			go b.callbackHandler(update)
		}

		if update.Message != nil {
			go b.messageHandler(update)
		}
	}
}

// registerCallbacks - register all callbacks
func (b *Bot) registerCallbacks() {
	// order callbacks
	b.callbacks["skip"] = b.skip
	b.callbacks["dislike"] = b.dislike
	b.callbacks["description"] = b.description
	b.callbacks["photo"] = b.photo

	// settings callbacks
	b.callbacks["back"] = b.backCallback
	b.callbacks["settings"] = b.settingsCallback

	// settings search callbacks
	b.callbacks["search"] = b.searchCallback
	b.callbacks["searchOn"] = b.searchCallback
	b.callbacks["searchOff"] = b.searchCallback

	// settings filters callbacks
	b.callbacks["filters"] = b.filtersCallback
	b.callbacks["withPhotoOn"] = b.withPhotoCallback
	b.callbacks["withPhotoOff"] = b.withPhotoCallback
	b.callbacks["KGS"] = b.priceCallback
	b.callbacks["USD"] = b.priceCallback
}

// callbackHandler - handle all callback from user in go routines
func (b *Bot) callbackHandler(update tgbotapi.Update) {
	b.callbacks[update.CallbackQuery.Data](update.CallbackQuery)
}

// messageHandler - handle all user message from user in go routines
func (b *Bot) messageHandler(update tgbotapi.Update) {
	if update.Message.IsCommand() {
		msg := ""
		switch update.Message.Command() {
		case "start", "help":
			msg = b.start(update.Message)
		case "stop":
			msg = b.stop(update.Message)
		case "search":
			msg = b.search(update.Message)
		case "feedback":
			msg = b.feedback(update.Message)
		case "settings":
			b.callbacks["settings"](&tgbotapi.CallbackQuery{
				Message: update.Message,
			})
			return
		default:
			msg = "Нет среди доступных команд :("
		}

		message := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
		_, err := b.bot.Send(message)
		if err != nil {
			log.Println("[Start.Message.Send] error: ", err)
		}
		return
	}
	b.answerListener(update.Message)
}

// answerListener - if we need wait answer from some chat, we add waite command
//  to waitAnswers. This function listen all message and check need wait answer
//  or not. If need, we call callback and remove from wait map
func (b *Bot) answerListener(message *tgbotapi.Message) {
	b.waitMutex.Lock()
	defer b.waitMutex.Unlock()

	answer, ok := b.waitAnswers[message.Chat.ID]
	if ok {
		if answer.deadline.Unix() > time.Now().Unix() {
			answer.callback(message, answer)
			return
		}

		b.clearRetry(message.Chat, -1)
	}
}
