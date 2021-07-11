package background

import (
	"context"
	"log"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/comov/hsearch/parser"
)

// todo: refactor this
// grabber - парсит удаленные ресурсы, находит предложения и пишет в хранилище,
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
			ctx := context.Background()

			for _, site := range m.sitesForParse {
				go m.grabbedApartments(ctx, site)
			}
		}
	}
}

func (m *Manager) grabbedApartments(ctx context.Context, site Site) {
	log.Printf("[grabber] StartGrabber parse `%s`\n", site.Name())
	apartmentsLinks, err := parser.FindApartmentsLinksOnSite(site)
	if err != nil {
		sentry.CaptureException(err)
		log.Printf("[grabber.FindApartmentsLinksOnSite] Error: %s\n", err)
		return
	}

	if len(apartmentsLinks) == 0 {
		log.Printf("[grabber] No apartments for site `%s`\n", site.Name())
		return
	}

	err = m.st.CleanFromExistApartments(ctx, apartmentsLinks, site.Name())
	if err != nil {
		sentry.CaptureException(err)
		log.Printf("[grabber.CleanFromExistOrders] Error: %s\n", err)
		return
	}

	log.Printf("[grabber] Find %d apartment for site `%s`\n", len(apartmentsLinks), site.Name())

	apartments := parser.LoadApartmentsDetail(apartmentsLinks, site)
	log.Printf("[grabber] Find %d new apartments for site `%s`\n", len(apartments), site.Name())

	_, err = m.st.WriteApartments(ctx, apartments)
	if err != nil {
		sentry.CaptureException(err)
		log.Printf("[grabber.WriteApartment] Error: %s\n", err)
	}
}
