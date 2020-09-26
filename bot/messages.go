package bot

import (
	"fmt"
	"strings"

	"github.com/comov/hsearch/structs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const helpMessage = `
Поиска квартир для долгосрочной аренды по Кыргызстану. Тут есть фильтры и нет дубликатов при просмотре объявлений

Доступные команды:
/help - справка по командам
/settings - настройки и фильтры бота
/feedback - отставить гневное сообщение автору 😐
`

const feedbackText = `Бот будет ждать от тебя сообщения примерно минут 5, после чего отправленный текст не будет считать фидбэком`
const wrongAnswerText = `Ты что-то не так ввел. Посмотри пример и попробуй еще раз. Осталось попыток: %d`
const somethingWrong = "Что-то пошло не так..."

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
		message.Grow(len("Площадь: ") + len(offer.Area) + len("\n"))
		message.WriteString("Площадь: ")
		message.WriteString(offer.Area)
		message.WriteString("\n")
	}

	if offer.Phone != "" {
		message.Grow(len("Номер: ") + len(offer.Phone) + len("\n"))
		message.WriteString("Номер: ")
		message.WriteString(offer.Phone)
		message.WriteString("\n")
	}

	message.Grow(len("\n") + len(offer.Url) + len("\n"))
	message.WriteString("\n")
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
