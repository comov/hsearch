package bots

import (
	"log"

	"github.com/aastashov/house_search_assistant/structs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// initAnswers - содержит список всех зарегистированных кнопок
func (b *Bot) initAnswers() {
	b.answers["skip"] = b.skip
	b.answers["dislike"] = b.dislike
	b.answers["description"] = b.description
	b.answers["photo"] = b.photo
}

// skip - обраьатывает нажатие на кнопку "Пропустить"
func (b *Bot) skip(query *tgbotapi.CallbackQuery) {
	user := &structs.User{
		Chat:     query.Message.Chat.ID,
		Username: query.Message.Chat.UserName,
	}
	err := b.st.Skip(query.Message.MessageID, user)
	if err != nil {
		log.Println("[skip.Skip] error:", err)
		return
	}

	offer, err := b.st.ReadNextOffer(user)
	if err != nil {
		log.Println("[skip.ReadNextOffer] error:", err)
		return
	}

	err = b.SendOffer(offer, user, query, "Покажу позже")
	if err != nil {
		log.Println("[skip.SendOffer] error:", err)
		return
	}
}

// dislike - обраьатывает нажатие на кнопку "Точно нет!"
func (b *Bot) dislike(query *tgbotapi.CallbackQuery) {
	user := &structs.User{
		Chat:     query.Message.Chat.ID,
		Username: query.Message.Chat.UserName,
	}

	err := b.st.Dislike(query.Message.MessageID, user)
	if err != nil {
		log.Println("[dislike.Dislike] error:", err)
		return
	}

	offer, err := b.st.ReadNextOffer(user)
	if err != nil {
		log.Println("[dislike.ReadNextOffer] error:", err)
		return
	}

	err = b.SendOffer(offer, user, query, "Больше никогда не покажу")
	if err != nil {
		log.Println("[dislike.SendOffer] error:", err)
		return
	}
}

func (b *Bot) description(query *tgbotapi.CallbackQuery) {
	user := &structs.User{
		Chat:     query.Message.Chat.ID,
		Username: query.Message.Chat.UserName,
	}

	body, err := b.st.ReadOfferDescription(query.Message.MessageID, user)
	if err != nil {
		log.Println("[description.ReadOfferDescription] error:", err)
		return
	}

	message := tgbotapi.NewMessage(user.Chat, body)
	message.ReplyToMessageID = query.Message.MessageID

	_, err = b.bot.Send(message)
	if err != nil {
		log.Println("[description.Send] error:", err)
	}
}

func (b *Bot) photo(query *tgbotapi.CallbackQuery) {
	user := &structs.User{
		Chat:     query.Message.Chat.ID,
		Username: query.Message.Chat.UserName,
	}

	images, err := b.st.ReadOfferImages(query.Message.MessageID, user)
	if err != nil {
		log.Println("[photo.ReadOfferDescription] error:", err)
		return
	}

	waitMessage := tgbotapi.Message{}
	if len(images) != 0 {
		waitMessage, err = b.bot.Send(tgbotapi.NewMessage(
			user.Chat,
			WaitPhotoMessage(len(images)),
		))
		if err != nil {
			log.Println("[photo.Send] error:", err)
		}
	}

	imgs := make([]interface{}, 0)
	for _, img := range images {
		imgs = append(imgs, tgbotapi.NewInputMediaPhoto(img))
	}

	message := tgbotapi.NewMediaGroup(user.Chat, imgs)
	message.ReplyToMessageID = query.Message.MessageID

	_, err = b.bot.Send(message)
	if err != nil {
		log.Println("[photo.Send] error:", err)
		_, err = b.bot.Send(tgbotapi.NewMessage(user.Chat, "Видимо много фото, не получается отправить. Потом починю"))
	}

	if len(images) != 0 {
		_, _ = b.bot.DeleteMessage(tgbotapi.NewDeleteMessage(
			query.Message.Chat.ID,
			waitMessage.MessageID,
		))
	}
}
