package background

import (
	"log"
	"time"

	"github.com/comov/hsearch/parser"
)

// todo: refactor this
// grabber - парсит удаленнвые ресурсы, находит предложения и пишет в хранилище,
// после чего трегерит broker
func (m *Manager) grabber() {
	// при первом запуске менеджера, он начнет первый парсинг через 2 секунды,
	// а после изменится на время из настроек (sleep = m.cnf.ManagerDelay)
	sleep := time.Second * 2

	log.Printf("[grabber] StartGrabber Manager\n")
	for {
		select {
		case <-time.After(sleep):
			sleep = m.cnf.FrequencyTime

			for _, site := range m.sitesForParse {
				go m.grabbedOffers(site)
			}
		}
	}
}

func (m *Manager) grabbedOffers(site Site) {
	log.Printf("[grabber] StartGrabber parse `%s`\n", site.Name())
	offersLinks, err := parser.LoadNewOffers(site)
	if err != nil {
		log.Printf("[grabber.LoadNewOffers] Error: %s\n", err)
		return
	}

	if len(offersLinks) == 0 {
		log.Printf("[grabber] No offers for site `%s`\n", site.Name())
		return
	}

	err = m.st.CleanFromExistOrders(offersLinks, site.Name())
	if err != nil {
		log.Printf("[grabber.CleanFromExistOrders] Error: %s\n", err)
		return
	}

	log.Printf("[grabber] Find %d offer for site `%s`\n", len(offersLinks), site.Name())

	offers := parser.LoadOffersDetail(offersLinks, site)
	log.Printf("[grabber] Find %d new offers for site `%s`\n", len(offers), site.Name())

	_, err = m.st.WriteOffers(offers)
	if err != nil {
		log.Printf("[grabber.WriteOffer] Error: %s\n", err)
	}
}
