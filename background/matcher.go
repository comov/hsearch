package background

import (
	"context"
	"log"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/comov/hsearch/structs"
)

// todo: refactor this
// matcher - an intermediary to receive all users and start the mailing list
//  for them
func (m *Manager) matcher(ctx context.Context) {
	sleep := time.Second * 2

	log.Printf("[matcher] StartMatcher Manager\n")
	for {
		select {
		case <-time.After(sleep):
			sleep = m.cnf.FrequencyTime

			chats, err := m.st.ReadChatsForMatching(ctx, 1)
			if err != nil {
				sentry.AddBreadcrumb(&sentry.Breadcrumb{
					Category: "matcher.ReadChatsForMatching",
					Data: map[string]interface{}{
						"sleep": sleep,
						"chats": chats,
					},
				})
				sentry.CaptureException(err)
				log.Printf("[matcher.ReadChatForOrder] Error: %s\n", err)
				return
			}

			for _, chat := range chats {
				go m.matching(ctx, chat)
			}
		}
	}
}

func (m *Manager) matching(ctx context.Context, chat *structs.Chat) {
	log.Printf("[matcher] Startmatcher matching for `%s`\n", chat.Title)

	offer, err := m.st.ReadNextOffer(ctx, chat)
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "matcher.ReadNextOffer",
			Data: map[string]interface{}{
				"chat.id": chat.Id,
				"chat.title": chat.Title,
			},
		})
		log.Printf("[matcher] Can't read offer for %s with an error: %s\n", chat.Title, err)
		return
	}

	if offer == nil {
		log.Printf("[matcher] For `%s` not new offers\n", chat.Title)
		return
	}

	err = m.bot.SendOffer(ctx, offer, chat.Id)
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "matcher",
			Data: map[string]interface{}{
				"method": "SendOffer",
				"offer.id": offer.Id,
				"offer.url": offer.Url,
				"offer.topic": offer.Topic,
				"chat.id": chat.Id,
				"chat.title": chat.Title,
			},
		})
		sentry.CaptureException(err)
		log.Printf("[matcher] Can't send message for `%s` with an error: %s\n", chat.Title, err)
		return
	}

	log.Printf("[matcher] Successfully send offer %d for `%s`\n", offer.Id, chat.Title)
}
