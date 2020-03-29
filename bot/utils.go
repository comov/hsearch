package bot

import (
	"encoding/json"
	"fmt"
	"github.com/comov/hsearch/structs"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/url"
	"strconv"
	"time"
)

// SendGroupPhotos - sends a group of photos to a chat room, but unlike Send,
// it returns a list of messages because sending a group of photos is sending
// multiple messages.
func (b *Bot) SendGroupPhotos(config tgbotapi.MediaGroupConfig) ([]tgbotapi.Message, error) {
	params, err := buildParams(config)
	if err != nil {
		return []tgbotapi.Message{}, err
	}

	resp, err := b.bot.MakeRequest("sendMediaGroup", params)
	if err != nil {
		return []tgbotapi.Message{}, err
	}

	var messages []tgbotapi.Message
	err = json.Unmarshal(resp.Result, &messages)
	return messages, err
}

func buildParams(config tgbotapi.MediaGroupConfig) (url.Values, error) {
	chat := config.BaseChat
	v := url.Values{}
	if chat.ChannelUsername != "" {
		v.Add("chat_id", chat.ChannelUsername)
	} else {
		v.Add("chat_id", strconv.FormatInt(chat.ChatID, 10))
	}

	if chat.ReplyToMessageID != 0 {
		v.Add("reply_to_message_id", strconv.Itoa(chat.ReplyToMessageID))
	}

	if chat.ReplyMarkup != nil {
		data, err := json.Marshal(chat.ReplyMarkup)
		if err != nil {
			return v, err
		}

		v.Add("reply_markup", string(data))
	}

	v.Add("disable_notification", strconv.FormatBool(chat.DisableNotification))

	data, err := json.Marshal(config.InputMedia)
	if err != nil {
		return v, err
	}

	v.Add("media", string(data))

	return v, nil
}


// SendOffer - send the offer to a chat and save the delivery report to a chat
//  room
func (b *Bot) SendOffer(offer *structs.Offer, chatId int64) error {
	message := tgbotapi.NewMessage(chatId, DefaultMessage(offer))
	message.DisableWebPagePreview = true
	message.ReplyMarkup = getKeyboard(offer)

	send, err := b.bot.Send(message)
	if err != nil {
		return err
	}
	return b.storage.SaveMessage(send.MessageID, offer.Id, chatId, structs.KindOffer)
}

func (b *Bot) addWaitCallback(c int64, answer answer) {
	b.waitMutex.Lock()
	defer b.waitMutex.Unlock()
	b.waitAnswers[c] = answer
}

func (b *Bot) wrongAnswer(message *tgbotapi.Message, a answer) {
	a.maxErrors -= 1
	a.deadline = time.Now().Add(time.Second * 20)
	if a.maxErrors <= 0 {
		b.clearRetry(message.Chat, message.MessageID)
		return
	}

	newMessage := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(wrongAnswerText, a.maxErrors))
	m, err := b.bot.Send(newMessage)
	if err != nil {
		b.waitAnswers[message.Chat.ID] = a
		log.Println("[wrongAnswer.Send] error:", err)
		return
	}

	a.messages = append(a.messages, message.MessageID, m.MessageID)
	b.waitAnswers[message.Chat.ID] = a
}

func (b *Bot) clearRetry(chat *tgbotapi.Chat, lastMsgId int) {
	a := b.waitAnswers[chat.ID]
	if lastMsgId != -1 {
		a.messages = append(a.messages, lastMsgId)
	}
	for _, id := range a.messages {
		deleteMessage := tgbotapi.NewDeleteMessage(chat.ID, id)
		_, err := b.bot.Send(deleteMessage)
		if err != nil {
			log.Println("[clearRetry.Send] error:", err)
		}
	}

	b.callbacks["filters"](&tgbotapi.CallbackQuery{Message: &tgbotapi.Message{
		Chat:      chat,
		MessageID: a.menuId,
	}})

	delete(b.waitAnswers, chat.ID)
}
