package bots

import (
	"log"

	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/structs"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type (
	Storage interface {
		StartSearch(chat int64, username string) error
		StopSearch(username string) error
		Dislike(msgId int, user *structs.User) ([]int, error)
		Skip(msgId int, user *structs.User) error
		Bookmarks(username string) ([]*structs.Offer, int64, error)
		Feedback(chat int64, username, body string) error

		SaveMessage(msgId int, offerId uint64, chat int64, kind string) error
		ReadOfferDescription(msgId int, user *structs.User) (uint64, string, error)
		ReadOfferImages(msgId int, user *structs.User) (uint64, []string, error)
		ReadNextOffer(user *structs.User) (*structs.Offer, error)
	}

	answer map[string]func(query *tgbotapi.CallbackQuery)

	Bot struct {
		bot     *tgbotapi.BotAPI
		st      Storage
		answers answer
	}
)

func NewTelegramBot(cnf *configs.Config, st Storage) *Bot {
	bot, err := tgbotapi.NewBotAPI(cnf.TelegramToken)
	if err != nil {
		log.Fatalln("[NewBot.NewBotAPI] error: ", err)
		return nil
	}

	bb := &Bot{
		bot:     bot,
		st:      st,
		answers: make(answer, 0),
	}

	bb.initAnswers()
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
			b.answers[update.CallbackQuery.Data](update.CallbackQuery)
		}

		if update.Message != nil {
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
				default:
					msg = "Нет среди доступных команд :("
				}

				message := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
				_, err := b.bot.Send(message)
				if err != nil {
					log.Println("[Start.Message.Send] error: ", err)
				}
			}
		}
	}
}

// TODO: It looks like shit. Please, rewrite this code :cry:
// SendOffer - отправляет offer пользователю, так же региструрует в бд под
// какие номером сообщения было отправленно сообщение и меняет клавиатуру в
// зависимости от offer
func (b *Bot) SendOffer(offer *structs.Offer, user *structs.User, query *tgbotapi.CallbackQuery, answer string) error {
	if query != nil {
		/* TODO: Найти offer по ID и сделать соответствующее действие */

		_, err := b.bot.AnswerCallbackQuery(tgbotapi.NewCallback(query.ID, answer))
		if err != nil {
			log.Println("[SendOffer.AnswerCallbackQuery] error: ", err)
		}
	}

	if offer == nil {
		return nil
	}

	message := tgbotapi.NewMessage(user.Chat, DefaultMessage(offer))
	message.DisableWebPagePreview = true
	message.ReplyMarkup = getKeyboard(offer)

	send, err := b.bot.Send(message)
	if err != nil {
		return err
	}

	err = b.st.SaveMessage(send.MessageID, offer.Id, user.Chat, structs.KindOffer)
	if err != nil {
		// Если и произошла ошибка, то пользователь уже получил сообщение в
		// телеграм. Мы просто оповещаем разработчика через лог и говорим, что
		// отправка сообщения была успешно
		log.Println("[SendOffer.SaveMessage] error:", err)
	}

	return nil
}

func (b *Bot) SendPreviewMessage(offer *structs.Offer, user *structs.User) error {
	return b.SendOffer(offer, user, nil, "")
}
