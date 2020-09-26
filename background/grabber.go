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
func (m *Manager) grabber(ctx context.Context) {
	// при первом запуске менеджера, он начнет первый парсинг через 2 секунды,
	// а после изменится на время из настроек (sleep = m.cnf.ManagerDelay)
	sleep := time.Second * 2

	log.Printf("[grabber] StartGrabber Manager\n")
	for {
		select {
		case <-time.After(sleep):
			sleep = m.cnf.FrequencyTime

			for _, site := range m.sitesForParse {
				go m.grabbedOffers(ctx, site)
			}
		}
	}
}

func (m *Manager) grabbedOffers(ctx context.Context, site Site) {
	log.Printf("[grabber] StartGrabber parse `%s`\n", site.Name())
	offersLinks, err := parser.FindOffersLinksOnSite(site)
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "grabber",
			Data: map[string]interface{}{
				"method": "grabbedOffers.FindOffersLinksOnSite",
				"site": site.Name(),
			},
		})
		sentry.CaptureException(err)
		log.Printf("[grabber.FindOffersLinksOnSite] Error: %s\n", err)
		return
	}

	if len(offersLinks) == 0 {
		log.Printf("[grabber] No offers for site `%s`\n", site.Name())
		return
	}

	err = m.st.CleanFromExistOrders(ctx, offersLinks, site.Name())
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "grabber",
			Data: map[string]interface{}{
				"method": "grabbedOffers.CleanFromExistOrders",
				"site": site.Name(),
				"offersLinks": offersLinks,
			},
		})
		sentry.CaptureException(err)
		log.Printf("[grabber.CleanFromExistOrders] Error: %s\n", err)
		return
	}

	log.Printf("[grabber] Find %d offer for site `%s`\n", len(offersLinks), site.Name())

	offers := parser.LoadOffersDetail(offersLinks, site)
	log.Printf("[grabber] Find %d new offers for site `%s`\n", len(offers), site.Name())

	err = m.st.WriteOffers(ctx, offers)
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "grabber",
			Data: map[string]interface{}{
				"method": "grabbedOffers.WriteOffers",
				"site": site.Name(),
				"offersLinks": offersLinks,
				"offers": offers,
			},
		})
		sentry.CaptureException(err)
		log.Printf("[grabber.WriteOffer] Site: %s; Error: %s\n", site.Name(), err)
	}
}
