package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/comov/hsearch/structs"
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

// SendApartment - send the apartment to a chat and save the delivery report to a chat
//  room
func (b *Bot) SendApartment(ctx context.Context, apartment *structs.Apartment, chat *structs.Chat) error {
	message := tgbotapi.NewMessage(chat.ChatId, DefaultMessage(apartment))
	message.DisableWebPagePreview = true
	message.ParseMode = tgbotapi.ModeMarkdown

	if !chat.IsChannel() {
		message.ReplyMarkup = getKeyboard(apartment)
	}

	msg, err := b.Send(message)
	if err != nil {
		return err
	}

	err = b.storage.SaveMessage(ctx, msg.MessageID, apartment.Id, chat.ChatId, structs.KindApartment)
	if err != nil {
		return err
	}

	if !chat.IsChannel() {
		return nil
	}

	query := &tgbotapi.CallbackQuery{
		Message: &tgbotapi.Message{
			MessageID: msg.MessageID,
			Chat: &tgbotapi.Chat{
				ID:   chat.ChatId,
				Type: "channel",
			},
		},
	}
	b.photo(ctx, query)
	return nil
}

func (b *Bot) addWaitCallback(c int64, answer answer) {
	b.waitMutex.Lock()
	defer b.waitMutex.Unlock()
	b.waitAnswers[c] = answer
}

func (b *Bot) wrongAnswer(ctx context.Context, message *tgbotapi.Message, a answer) {
	a.maxErrors -= 1
	a.deadline = time.Now().Add(time.Second * 20)
	if a.maxErrors <= 0 {
		b.clearRetry(ctx, message.Chat, message.MessageID)
		return
	}

	newMessage := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf(wrongAnswerText, a.maxErrors))
	m, err := b.Send(newMessage)
	if err != nil {
		b.waitAnswers[message.Chat.ID] = a
		log.Println("[wrongAnswer.Send] error:", err)
		return
	}

	a.messages = append(a.messages, message.MessageID, m.MessageID)
	b.waitAnswers[message.Chat.ID] = a
}

func (b *Bot) clearRetry(ctx context.Context, chat *tgbotapi.Chat, lastMsgId int) {
	a := b.waitAnswers[chat.ID]
	if lastMsgId != -1 {
		a.messages = append(a.messages, lastMsgId)
	}
	for _, id := range a.messages {
		deleteMessage := tgbotapi.NewDeleteMessage(chat.ID, id)
		_, err := b.Send(deleteMessage)
		if err != nil {
			log.Println("[clearRetry.Send] error:", err)
		}
	}

	if a.menuId != 0 {
		b.callbacks["filters"](ctx, &tgbotapi.CallbackQuery{Message: &tgbotapi.Message{
			Chat:      chat,
			MessageID: a.menuId,
		}})
	}

	delete(b.waitAnswers, chat.ID)
}
