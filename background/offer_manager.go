// Запустив впервые этого менеджера, я получил высер из сообщений. Подумав как
// обойти не только этот высер, но и оповещать пользователя о новых
// предложениях, я решил что следует сделать 2 инструмента. 1-й будет просто
// искать и писать в бд новые offers, после чего будет трегерить 2-й
// инструмент который будет проходиться по всем пользователям и искать любое
// первое предложение за последнее N времени и высылать в чат
package background

/* TODO: для всего пакета
    1. Для менеджера лучше сделать отдельный логер в файл и более подробно
      логировать его работу, так как сейчас это немного черный ящик
    2. Парсер нужно переделать на горутины, так как мы тратим много времени на
      ожидание по сети, а могли бы в этот момент парсить данные и писать их в
      бд. Нужно рассмотреть reader, writer для реализации pipeline
    3. ...
*/

import (
	"fmt"
	"log"
	"time"

	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/parser"
	"github.com/comov/hsearch/structs"
)

type (
	Storage interface {
		WriteOffers(offer []*structs.Offer) (int, error)
		ReadChatsForMatching() ([]*structs.Chat, error)
		ReadNextOffer(chatId int64) (*structs.Offer, error)
		CleanFromExistOrders(offers map[uint64]string) error
	}

	Bot interface {
		SendOffer(offer *structs.Offer, chatId int64) error
	}

	OfferManager struct {
		st  Storage
		bot Bot
		url string
		cnf *configs.Config
	}
)


// кодичество предложений на одной странице в diesel
const offersOnPage = 39

// StartOfferManager - инициализирует менеджера
func StartOfferManager(url string, cnf *configs.Config, st Storage, bot Bot) *OfferManager {
	return &OfferManager{
		st:  st,
		bot: bot,
		url: url,
		cnf: cnf,
	}
}

// Start - запускает работу менеджера
func (m *OfferManager) Start() {
	m.parser()
}

// parser - парсит удаленнвые ресурсы, находит предложения и пишет в хранилище,
// после чего трегерит broker
func (m *OfferManager) parser() {
	// при первом запуске менеджера, он начнет первый парсинг через 2 секунды,
	// а после изменится на время из настроек (sleep = m.cnf.ManagerDelay)
	sleep := time.Second * 2

	log.Println("[offer_manager] Start Offer Manager")
	for {
		select {
		case <-time.After(sleep):
			sleep = m.cnf.ManagerDelay

			for i := 1; i <= m.cnf.MaxPage; i++ {
				target := fmt.Sprintf(m.url, i)
				log.Printf("[offer_manager] start parse %s\n", target)
				offersLinks, err := parser.LoadNewOffers(target)
				if err != nil {
					log.Println("[offer_manager.LoadNewOffers] error: ", err)
					continue
				}

				if len(offersLinks) == 0 {
					log.Println("[offer_manager] no offers for target", target)
					continue
				}

				err = m.st.CleanFromExistOrders(offersLinks)
				if err != nil {
					log.Println("[offer_manager.CleanFromExistOrders] error", err)
					continue
				}

				offers := parser.LoadOffersDetail(offersLinks)
				log.Printf("[offer_manager] Find %d new offers for targer %s", len(offers), target)

				newOffers, err := m.st.WriteOffers(offers)
				if err != nil {
					log.Println("[offer_manager.WriteOffer] error", err)
					continue
				}

				if newOffers < offersOnPage {
					m.cnf.MaxPage = 1
				}
			}

			// после того как мы нашли новые offer, мы начинаем рассылку
			m.broker()
		}
	}
}

// broker - вытягивает вск чаты из бд и начинает для них рассылку
func (m *OfferManager) broker() {
	chats, err := m.st.ReadChatsForMatching()
	if err != nil {
		log.Println("[offer_manager.ReadChatForOrder] error:", err)
		return
	}

	if len(chats) <= 0 {
		log.Println("[offer_manager.chats.len] no chats")
		return
	}

	for _, chat := range chats {
		// TODO: каждый раз создавать горутину на чат, это жирно. Нужно
		//  сделать воркеров которые будут создаваться при старте, затем
		//  передавать им чаты и они будут заниматься matching
		go m.matching(chat)
	}
}

func (m *OfferManager) matching(chat *structs.Chat) {
	log.Printf("[offer_manager] Start matching for %s", chat.Title)

	offer, err := m.st.ReadNextOffer(chat.Id)
	if err != nil {
		log.Printf("[offer_manager] Can't read offer for %s with an error %s\n", chat.Title, err)
		return
	}

	if offer == nil {
		log.Printf("[offer_manager] For %s not new offers", chat.Title)
		return
	}

	err = m.bot.SendOffer(offer, chat.Id)
	if err != nil {
		log.Printf("[offer_manager] Can't send message for `%s` with an error %s\n", chat.Title, err)
		return
	}

	log.Printf("[offer_manager] Successfully send offer %d for %s\n", offer.Id, chat.Title)
}
