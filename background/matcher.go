package background

import (
	"log"
	"time"

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

			chats, err := m.st.ReadChatsForMatching(1)
			if err != nil {
				log.Printf("[matcher.ReadChatForOrder] Error: %s\n", err)
				return
			}

			for _, chat := range chats {
				go m.matching(chat)
			}
		}
	}
}

func (m *Manager) matching(chat *structs.Chat) {
	log.Printf("[matcher] Startmatcher matching for `%s`\n", chat.Title)

	offer, err := m.st.ReadNextOffer(chat)
	if err != nil {
		log.Printf("[matcher] Can't read offer for %s with an error: %s\n", chat.Title, err)
		return
	}

	if offer == nil {
		log.Printf("[matcher] For `%s` not new offers\n", chat.Title)
		return
	}

	err = m.bot.SendOffer(offer, chat.Id)
	if err != nil {
		log.Printf("[matcher] Can't send message for `%s` with an error: %s\n", chat.Title, err)
		return
	}

	log.Printf("[matcher] Successfully send offer %d for `%s`\n", offer.Id, chat.Title)
}
