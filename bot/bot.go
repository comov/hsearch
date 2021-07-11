package bot

import (
	"context"
	"github.com/comov/hsearch/bot/settings"
	"log"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/structs"
)

type (
	Storage interface {
		Dislike(ctx context.Context, msgId int, chatId int64) ([]int, error)
		Feedback(ctx context.Context, chat int64, username, body string) error

		SaveMessage(ctx context.Context, msgId int, apartmentId uint64, chat int64, kind string) error
		ReadApartmentDescription(ctx context.Context, msgId int, chatId int64) (uint64, string, error)
		ReadApartmentImages(ctx context.Context, msgId int, chatId int64) (uint64, []string, error)

		ReadChat(ctx context.Context, id int64) (*structs.Chat, error)
		CreateChat(ctx context.Context, id int64, username, title, cType string, enabled bool) error
		DeleteChat(ctx context.Context, id int64) error
		UpdateSettings(ctx context.Context, chat *structs.Chat) error
	}

	callback func(ctx context.Context, query *tgbotapi.CallbackQuery)

	Bot struct {
		bot       *tgbotapi.BotAPI
		storage   Storage
		callbacks map[string]callback

		adminChatId int64
		release     string

		// {time.Minutes * 3, b.callbackName(message *tgbotapi.Message)}
		waitAnswers map[int64]answer
		waitMutex   sync.Mutex
	}
)

func NewTelegramBot(cnf *configs.Config, st Storage) *Bot {
	bot, err := tgbotapi.NewBotAPI(cnf.TelegramToken)
	if err != nil {
		log.Fatalln("[bot.NewBotAPI] error: ", err)
		return nil
	}

	bb := &Bot{
		bot:         bot,
		storage:     st,
		adminChatId: cnf.TelegramChatId,
		release:     cnf.Release,
		callbacks:   make(map[string]callback, 0),
		waitAnswers: make(map[int64]answer),
	}

	bb.registerCallbacks()
	return bb
}

func (b *Bot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return b.bot.Send(c)
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalln("[bot.GetUpdatesChan] error: ", err)
		return
	}

	log.Printf("[bot] Start listen Telegram chanel. Version %s\n", b.release)
	for update := range updates {
		ctx := context.Background()
		if update.CallbackQuery != nil {
			go b.callbackHandler(ctx, update)
		}

		if update.Message != nil {
			go b.messageHandler(ctx, update)
		}

		if update.ChannelPost != nil {
			go b.channelHandler(ctx, update)
		}
	}
}

// registerCallbacks - register all callbacks
func (b *Bot) registerCallbacks() {
	// order callbacks
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
	b.callbacks["dieselOn"] = b.searchCallback
	b.callbacks["dieselOff"] = b.searchCallback
	b.callbacks["houseOn"] = b.searchCallback
	b.callbacks["houseOff"] = b.searchCallback
	b.callbacks["lalafoOn"] = b.searchCallback
	b.callbacks["lalafoOff"] = b.searchCallback

	// settings filters callbacks
	b.callbacks["filters"] = b.filtersCallback
	b.callbacks["withPhotoOn"] = b.withPhotoCallback
	b.callbacks["withPhotoOff"] = b.withPhotoCallback
	b.callbacks["KGS"] = b.priceCallback
	b.callbacks["USD"] = b.priceCallback
}

// callbackHandler - handle all callback from user in go routines
func (b *Bot) callbackHandler(ctx context.Context, update tgbotapi.Update) {
	b.callbacks[update.CallbackQuery.Data](ctx, update.CallbackQuery)
}

// messageHandler - handle all user message from user in go routines
func (b *Bot) messageHandler(ctx context.Context, update tgbotapi.Update) {
	if update.Message.IsCommand() {
		msg := ""
		switch update.Message.Command() {
		case "start":
			msg = b.start(ctx, update.Message.Chat)
		case "stop":
			msg = b.stop(ctx, update.Message.Chat)
		case "help":
			msg = b.help(update.Message)
		case "settings":
			b.callbacks["settings"](ctx, &tgbotapi.CallbackQuery{Message: update.Message})
			return
		case "feedback":
			msg = b.feedback(ctx, update.Message.Chat)
		default:
			msg = "Нет среди доступных команд :("
		}

		message := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
		_, err := b.Send(message)
		if err != nil {
			sentry.CaptureException(err)
			log.Println("[messageHandler.bot.Send] error: ", err)
		}
		return
	}
	b.answerListener(ctx, update.Message)
}

func (b *Bot) channelHandler(ctx context.Context, update tgbotapi.Update) {
	if !update.ChannelPost.IsCommand() {
		return
	}

	msg := ""
	switch update.ChannelPost.Command() {
	case "start":
		msg = b.start(ctx, update.ChannelPost.Chat)
	case "stop":
		msg = b.stop(ctx, update.ChannelPost.Chat)
	case "help":
		msg = b.help(update.ChannelPost)
	case "settings":
		chat, err := b.storage.ReadChat(ctx, update.ChannelPost.Chat.ID)
		if err != nil {
			sentry.CaptureException(err)
			log.Println("[channelHandler.settings.ReadChat] error:", err)
			return
		}

		_, err = b.Send(settings.MainSettingsHandler(update.ChannelPost, chat))
		if err != nil {
			sentry.CaptureException(err)
			log.Println("[channelHandler.settings.Send] error:", err)
		}
		return
	case "feedback":
		msg = b.feedback(ctx, update.ChannelPost.Chat)
	default:
		msg = "Нет среди доступных команд"
	}

	message := tgbotapi.NewMessage(update.ChannelPost.Chat.ID, msg)
	_, err := b.Send(message)
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[channelHandler.bot.Send] error: ", err)
	}
}

// answerListener - if we need wait answer from some chat, we add waite command
//  to waitAnswers. This function listen all message and check need wait answer
//  or not. If need, we call callback and remove from wait map
func (b *Bot) answerListener(ctx context.Context, message *tgbotapi.Message) {
	b.waitMutex.Lock()
	defer b.waitMutex.Unlock()

	answer, ok := b.waitAnswers[message.Chat.ID]
	if ok {
		if answer.deadline.Unix() > time.Now().Unix() {
			answer.callback(ctx, message, answer)
			return
		}

		b.clearRetry(ctx, message.Chat, -1)
	}
}

func (b *Bot) SendError(where string, err error, chatId int64) {
	log.Println("[", where, "] error:", err)
	_, err = b.Send(tgbotapi.NewMessage(chatId, somethingWrong))
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[", where, ".Send.error] error:", err)
	}
}
