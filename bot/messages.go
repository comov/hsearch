package bot

import (
	"fmt"
	"strings"

	"github.com/comov/hsearch/structs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const helpMessage = `
Это бот для поиска квартир. Основное его преимущество это фильтрация по просмотренным квартирам. Это не коммерческий проект, код в открытом доступе. Если есть идеи, оставляй фидбек :)

Доступные команды:
/start - зарегистрирует тебя для поиска квартиры
/help - справка по командам
/settings - настройки и фильтры бота
/feedback - отставить гневное сообщение автору 😐
`

const feedbackText = `Бот будет ждать от тебя сообщения примерно минут 5, после чего отправленный текст не будет считать фидбэком`
const wrongAnswerText = `То ли я тупой, то ли лыжи. Посмотри пример и попробуй еще разок. Осталось попытор: %d`

func DefaultMessage(offer *structs.Offer) string {
	var message strings.Builder
	message.WriteString(offer.Topic)
	message.WriteString("\n\n")

	if offer.FullPrice != "" {
		message.Grow(len("Цена: ") + len(offer.FullPrice) + len("\n"))
		message.WriteString("Цена: ")
		message.WriteString(offer.FullPrice)
		message.WriteString("\n")
	}

	if offer.Rooms != "" {
		message.Grow(len("Комнат: ") + len(offer.Rooms) + len("\n"))
		message.WriteString("Комнат: ")
		message.WriteString(offer.Rooms)
		message.WriteString("\n")
	}

	if offer.Floor != "" {
		message.Grow(len("Этаж: ") + len(offer.Floor) + len("\n"))
		message.WriteString("Этаж: ")
		message.WriteString(offer.Floor)
		message.WriteString("\n")
	}

	if offer.District != "" {
		message.Grow(len("Район: ") + len(offer.District) + len("\n"))
		message.WriteString("Район: ")
		message.WriteString(offer.District)
		message.WriteString("\n")
	}

	if offer.Area != "" {
		message.Grow(len("Площадь (кв.м.): ") + len(offer.Area) + len("\n"))
		message.WriteString("Площадь (кв.м.): ")
		message.WriteString(offer.Area)
		message.WriteString("\n")
	}

	if offer.Phone != "" {
		message.Grow(len("Номер: ") + len(offer.Phone) + len("\n"))
		message.WriteString("Номер: ")
		message.WriteString(offer.Phone)
		message.WriteString("\n")
	}

	message.Grow(len("Ссылка: ") + len(offer.Url) + len("\n"))
	message.WriteString("Ссылка: ")
	message.WriteString(offer.Url)
	message.WriteString("\n")
	return message.String()
}

func WaitPhotoMessage(count int) string {
	handler := func(end string) string {
		message := "Ща отправлю %d фот%s. Это долго, жди..."
		return fmt.Sprintf(message, count, end)
	}
	if count == 1 || count == 21 || count == 31 {
		return handler("ку")
	}
	if (count > 1 && count < 5) || (count > 21 && count < 25) {
		return handler("ки")
	}
	if (count >= 5 && count < 21) || (count >= 25 && count < 31) {
		return handler("ок")
	}

	return "Ща отправлю пару фоток. Это долго, жди..."
}

func getFeedbackAdminText(chat *tgbotapi.Chat, text string) string {
	msg := ""
	if chat.IsPrivate() {
		msg += fmt.Sprintf("Пользователь: %s %s\nС ником: %s\n\n",
			chat.FirstName,
			chat.LastName,
			chat.UserName,
		)
	} else {
		msg += fmt.Sprintf("В группе: %s\n\n", chat.Title)
	}

	msg += fmt.Sprintf("Оставил feedback:\n%s", text)
	return msg
}
