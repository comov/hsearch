package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/comov/hsearch/structs"
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

func DefaultMessage(apartment *structs.Apartment) string {
	var message strings.Builder
	message.WriteString(apartment.Topic)
	message.WriteString("\n\n")

	fullPrice := apartment.GetFullPrice()
	if fullPrice != "" {
		message.Grow(len("Цена: ") + len(fullPrice) + len("\n"))
		message.WriteString("Цена: ")
		message.WriteString(fullPrice)
		message.WriteString("\n")
	}

	if apartment.Rooms != 0 {
		roomsStr := fmt.Sprint(apartment.Rooms)
		message.Grow(len("Комнат: ") + len(roomsStr) + len("\n"))
		message.WriteString("Комнат: ")
		message.WriteString(roomsStr)
		message.WriteString("\n")
	}

	if apartment.Floor != 0 {
		floorStr := fmt.Sprint(apartment.Floor)
		message.Grow(len("Этаж: ") + len(floorStr) + len("\n"))
		message.WriteString("Этаж: ")
		message.WriteString(floorStr)
		message.WriteString("\n")
	}

	if apartment.District != "" {
		message.Grow(len("Район: ") + len(apartment.District) + len("\n"))
		message.WriteString("Район: ")
		message.WriteString(apartment.District)
		message.WriteString("\n")
	}

	if apartment.Area != 0 {
		areaStr := fmt.Sprint(apartment.Area)
		message.Grow(len("Площадь: ") + len(areaStr) + len("\n"))
		message.WriteString("Площадь: ")
		message.WriteString(areaStr)
		message.WriteString("\n")
	}

	if apartment.Phone != "" {
		message.Grow(len("Номер: ") + len(apartment.Phone) + len("\n"))
		message.WriteString("Номер: ")
		message.WriteString(apartment.Phone)
		message.WriteString("\n")
	}

	message.Grow(len("\n") + len(apartment.Url) + len("\n"))
	message.WriteString("\n")
	message.WriteString(apartment.Url)
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
