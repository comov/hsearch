package bots

import (
	"fmt"
	"log"

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

const templateMessage = `
%s

Цена: %s
Комнат: %s
Номер: %s
Ссылка: %s
`

const stopNotFound = `%s нет в базе. Это значит что я %s не отправлю`
const noOffers = `Пока нет новых предложений`

func DefaultMessage(offer *structs.Offer) string {
	return fmt.Sprintf(templateMessage,
		offer.Topic,
		offer.FullPrice,
		offer.Rooms,
		offer.Phone,
		offer.Url,
	)
}

func (b *Bot) bookmarksMessages(offers []*structs.Offer, chat int64) {
	for _, offer := range offers {
		err := b.SendOffer(offer, chat, nil, "")
		if err != nil {
			log.Println("[bookmarksMessages.SendOffer] error:", err)
		}
	}
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
