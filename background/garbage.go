package background

import (
	"context"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

func (m *Manager) garbage(ctx context.Context) {
	expireDate := time.Now().AddDate(0, 0, m.cnf.ExpireDays*-1).Unix()

	log.Printf("[garbage] StartGarbageCollector Manager\n")
	err := m.st.CleanExpiredOffers(ctx, expireDate)
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "garbage.CleanExpiredOffers",
			Data: map[string]interface{}{
				"expireDate": expireDate,
			},
		})
		sentry.CaptureException(err)
		log.Printf("[garbage.CleanExpiredOffers] Error: %s\n", err)
	}

	err = m.st.CleanExpiredImages(ctx, expireDate)
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "garbage.CleanExpiredImages",
			Data: map[string]interface{}{
				"expireDate": expireDate,
			},
		})
		sentry.CaptureException(err)
		log.Printf("[garbage.CleanExpiredImages] Error: %s\n", err)
	}

	err = m.st.CleanExpiredAnswers(ctx, expireDate)
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "garbage.CleanExpiredAnswers",
			Data: map[string]interface{}{
				"expireDate": expireDate,
			},
		})
		sentry.CaptureException(err)
		log.Printf("[garbage.CleanExpiredAnswers] Error: %s\n", err)
	}

	err = m.st.CleanExpiredTGMessages(ctx, expireDate)
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "garbage.CleanExpiredTGMessages",
			Data: map[string]interface{}{
				"expireDate": expireDate,
			},
		})
		sentry.CaptureException(err)
		log.Printf("[garbage.CleanExpiredTGMessages] Error: %s\n", err)
	}
}
