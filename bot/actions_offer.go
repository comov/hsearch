package bot

import (
	"context"
	"log"

	"github.com/comov/hsearch/structs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	dislikeButton     = tgbotapi.NewInlineKeyboardButtonData("Точно нет!", "dislike")
	descriptionButton = tgbotapi.NewInlineKeyboardButtonData("Описание", "description")
	photoButton       = tgbotapi.NewInlineKeyboardButtonData("Фото", "photo")
)

func getKeyboard(offer *structs.Offer) tgbotapi.InlineKeyboardMarkup {
	row1 := tgbotapi.NewInlineKeyboardRow(dislikeButton)
	row2 := tgbotapi.NewInlineKeyboardRow()

	if len(offer.Body) != 0 {
		row2 = append(row2, descriptionButton)
	}

	if offer.Images != 0 {
		row2 = append(row2, photoButton)
	}

	if len(row2) == 0 {
		return tgbotapi.NewInlineKeyboardMarkup(row1)
	}

	return tgbotapi.NewInlineKeyboardMarkup(row1, row2)
}

// dislike - this button delete order from chat and no more show to user that order
func (b *Bot) dislike(ctx context.Context, query *tgbotapi.CallbackQuery) {
	messagesIds, err := b.storage.Dislike(
		ctx,
		query.Message.MessageID,
		query.Message.Chat.ID,
	)
	if err != nil {
		log.Println("[dislike.Dislike] error:", err)
		return
	}

	for _, id := range messagesIds {
		_, err := b.bot.DeleteMessage(
			tgbotapi.NewDeleteMessage(query.Message.Chat.ID, id),
		)
		if err != nil {
			log.Println("[dislike.DeleteMessage] error:", err)
		}
	}

	_, err = b.bot.AnswerCallbackQuery(tgbotapi.NewCallback(
		query.ID, "Больше никогда не покажу",
	))
	if err != nil {
		log.Println("[dislike.AnswerCallbackQuery] error:", err)
		return
	}
}

// description - return full description about order
func (b *Bot) description(ctx context.Context, query *tgbotapi.CallbackQuery) {
	offerId, body, err := b.storage.ReadOfferDescription(
		ctx,
		query.Message.MessageID,
		query.Message.Chat.ID,
	)
	if err != nil {
		log.Println("[description.ReadOfferDescription] error:", err)
		return
	}

	message := tgbotapi.NewMessage(query.Message.Chat.ID, body)
	message.ReplyToMessageID = query.Message.MessageID

	send, err := b.bot.Send(message)
	if err != nil {
		log.Println("[description.Send] error:", err)
	}

	err = b.storage.SaveMessage(
		ctx,
		send.MessageID,
		offerId,
		query.Message.Chat.ID,
		structs.KindDescription,
	)
	if err != nil {
		log.Println("[photo.SaveMessage] error:", err)
	}
}

// photo - this button return all orders photos from site
func (b *Bot) photo(ctx context.Context, query *tgbotapi.CallbackQuery) {
	offerId, images, err := b.storage.ReadOfferImages(ctx, query.Message.MessageID, query.Message.Chat.ID)
	if err != nil {
		log.Println("[photo.ReadOfferDescription] error:", err)
		return
	}

	waitMessage := tgbotapi.Message{}
	if len(images) != 0 {
		waitMessage, err = b.bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, WaitPhotoMessage(len(images))))
		if err != nil {
			log.Println("[photo.Send] error:", err)
		}
	}

	for _, album := range getSeparatedAlbums(images) {
		imgs := make([]interface{}, 0)
		for _, img := range album {
			imgs = append(imgs, tgbotapi.NewInputMediaPhoto(img))
		}

		message := tgbotapi.NewMediaGroup(query.Message.Chat.ID, imgs)
		message.ReplyToMessageID = query.Message.MessageID

		messages, err := b.SendGroupPhotos(message)
		if err != nil {
			log.Println("[photo.Send] sending album error:", err)
		}

		for _, msg := range messages {
			err = b.storage.SaveMessage(
				ctx,
				msg.MessageID,
				offerId,
				query.Message.Chat.ID,
				structs.KindPhoto,
			)
			if err != nil {
				log.Println("[photo.SaveMessage] error:", err)
			}
		}
	}

	if len(images) != 0 {
		_, err := b.bot.DeleteMessage(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, waitMessage.MessageID))

		if err != nil {
			log.Println("[photo.DeleteMessage] error:", err)
		}
	}
}

// getSeparatedAlbums - separate images array to 10-items albums. Telegram API
//  has limit: `max images in images album is 10`
func getSeparatedAlbums(images []string) [][]string {
	maxImages := 10
	albums := make([][]string, 0, (len(images)+maxImages-1)/maxImages)

	for maxImages < len(images) {
		images, albums = images[maxImages:], append(albums, images[0:maxImages:maxImages])
	}
	return append(albums, images)
}
