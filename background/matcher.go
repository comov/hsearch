package background

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/comov/hsearch/structs"
)

// todo: refactor this
// matcher - an intermediary to receive all users and start the mailing list
//  for them
func (m *Manager) matcher() {
	sleep := time.Second * 2

	log.Printf("[matcher] StartMatcher Manager\n")
	for {
		select {
		case <-time.After(sleep):
			sleep = m.cnf.FrequencyTime
			ctx := context.Background()

			chats, err := m.st.ReadChatsForMatching(ctx, 1)
			if err != nil {
				sentry.AddBreadcrumb(&sentry.Breadcrumb{
					Category: "matcher",
					Data: map[string]interface{}{
						"method": "ReadChatsForMatching",
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

	apartment, err := m.st.ReadNextApartment(ctx, chat)
	if err != nil {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "matcher",
			Data: map[string]interface{}{
				"method": "ReadNextApartment",
				"chat.id": chat.ChatId,
				"chat.title": chat.Title,
			},
		})
		log.Printf("[matcher] Can't read apartment for %s with an error: %s\n", chat.Title, err)
		return
	}

	if apartment == nil {
		log.Printf("[matcher] For `%s` not new apartments\n", chat.Title)
		return
	}

	err = m.bot.SendApartment(ctx, apartment, chat)
	if err != nil {
		if strings.Contains(err.Error(), "blocked by the user") {
			chat.Enable = false
			if err = m.st.UpdateSettings(ctx, chat); err != nil {
				sentry.CaptureException(err)
			}
			return
		}

		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Category: "matcher",
			Data: map[string]interface{}{
				"method": "SendApartment",
				"apartment.id": apartment.Id,
				"apartment.url": apartment.Url,
				"apartment.topic": apartment.Topic,
				"chat.id": chat.ChatId,
				"chat.title": chat.Title,
			},
		})
		sentry.CaptureException(err)
		log.Printf("[matcher] Can't send message for `%s` with an error: %s\n", chat.Title, err)
		return
	}

	log.Printf("[matcher] Successfully send apartment %d for `%s`\n", apartment.Id, chat.Title)
}
