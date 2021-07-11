package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/comov/hsearch/structs"
)

var (
	dislikeButton     = tgbotapi.NewInlineKeyboardButtonData("Точно нет!", "dislike")
	descriptionButton = tgbotapi.NewInlineKeyboardButtonData("Описание", "description")
)

func getKeyboard(apartment *structs.Apartment) tgbotapi.InlineKeyboardMarkup {
	row1 := tgbotapi.NewInlineKeyboardRow(dislikeButton)
	row2 := tgbotapi.NewInlineKeyboardRow()

	if len(apartment.Body) != 0 {
		row2 = append(row2, descriptionButton)
	}

	if apartment.ImagesCount != 0 {
		row2 = append(row2, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Фото (%d)", apartment.ImagesCount), "photo"))
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
		sentry.CaptureException(err)
		log.Println("[dislike.Dislike] error:", err)
		return
	}

	for _, id := range messagesIds {
		_, err := b.bot.DeleteMessage(
			tgbotapi.NewDeleteMessage(query.Message.Chat.ID, id),
		)
		if err != nil {
			sentry.CaptureException(err)
			log.Println("[dislike.DeleteMessage] error:", err)
		}
	}

	_, err = b.bot.AnswerCallbackQuery(tgbotapi.NewCallback(
		query.ID, "Больше никогда не покажу",
	))
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[dislike.AnswerCallbackQuery] error:", err)
	}
}

// description - return full description about order
func (b *Bot) description(ctx context.Context, query *tgbotapi.CallbackQuery) {
	apartmentId, body, err := b.storage.ReadApartmentDescription(
		ctx,
		query.Message.MessageID,
		query.Message.Chat.ID,
	)
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[description.ReadApartmentDescription] error:", err)
		return
	}

	message := tgbotapi.NewMessage(query.Message.Chat.ID, body)
	message.ReplyToMessageID = query.Message.MessageID

	send, err := b.Send(message)
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[description.Send] error:", err)
	}

	err = b.storage.SaveMessage(
		ctx,
		send.MessageID,
		apartmentId,
		query.Message.Chat.ID,
		structs.KindDescription,
	)
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[photo.SaveMessage] error:", err)
	}
}

// photo - this button return all orders photos from site
func (b *Bot) photo(ctx context.Context, query *tgbotapi.CallbackQuery) {
	apartmentId, images, err := b.storage.ReadApartmentImages(ctx, query.Message.MessageID, query.Message.Chat.ID)
	if err != nil {
		sentry.CaptureException(err)
		log.Println("[photo.ReadApartmentDescription] error:", err)
		return
	}

	waitMessage := tgbotapi.Message{}
	if len(images) != 0 && query.Message.Chat.Type != "channel" {
		waitMessage, err = b.Send(tgbotapi.NewMessage(query.Message.Chat.ID, WaitPhotoMessage(len(images))))
		if err != nil {
			sentry.CaptureException(err)
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
			sentry.CaptureException(err)
			log.Println("[photo.Send] sending album error:", err)
		}

		for _, msg := range messages {
			err = b.storage.SaveMessage(
				ctx,
				msg.MessageID,
				apartmentId,
				query.Message.Chat.ID,
				structs.KindPhoto,
			)
			if err != nil {
				sentry.CaptureException(err)
				log.Println("[photo.SaveMessage] error:", err)
			}
		}
	}

	if len(images) != 0 && query.Message.Chat.Type != "channel" {
		_, err := b.bot.DeleteMessage(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, waitMessage.MessageID))

		if err != nil {
			sentry.CaptureException(err)
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
