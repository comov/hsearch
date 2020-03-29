package bot

import (
	"fmt"
	"strings"

	"github.com/comov/hsearch/structs"
)

const startMessage = `
Это бот для поиска квартир. Основное его приемущество это фильтрация по просмотренным квартирам. Это не коммерческий проект, код в открытом доступе. Если есть идеи, оставляй фидбек :)

Доступные команды:
/start - запуск бота
/help - справка по командам
/stop - исключит Вас из списка пользователей для рассылки и остановит бота
/search - включит поиск квартир, бот будет отправлять Вам новые квартиры как найдет
/feedback <text> - отставить гневное сообщение автору 😐
`

const wrongAnswerText = `То ли я тупой, то ли лыжи. Посмотри пример и попробуй еще разок. Осталось попытор: %d`
const stopNotFound = `%s нет в базе. Это значит что я %s не отправлю`

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
