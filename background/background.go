package background

import (
	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/parser"
	"github.com/comov/hsearch/structs"

	"github.com/PuerkitoBio/goquery"
)

type (
	Storage interface {
		WriteOffers(offer []*structs.Offer) (int, error)
		ReadChatsForMatching(enable int) ([]*structs.Chat, error)
		ReadNextOffer(chat *structs.Chat) (*structs.Offer, error)
		CleanFromExistOrders(offers map[uint64]string, siteName string) error
	}

	Bot interface {
		SendOffer(offer *structs.Offer, chatId int64) error
	}

	Site interface {
		Name() string
		FullHost() string
		Url() string
		Selector() string
		IdFromHref(href string) (uint64, error)
		ParseNewOffer(href string, exId uint64, doc *goquery.Document) *structs.Offer
	}

	Manager struct {
		st            Storage
		bot           Bot
		cnf           *configs.Config
		sitesForParse []Site
	}
)

// NewManager - initializes the manager
func NewManager(cnf *configs.Config, st Storage, bot Bot) *Manager {
	return &Manager{
		st:  st,
		bot: bot,
		cnf: cnf,
		sitesForParse: []Site{
			parser.DieselSite(),
			parser.LalafoSite(),
		},
	}
}

// StartGrabber - starts the process of finding new offers
func (m *Manager) StartGrabber() {
	m.grabber()
}

// StartGrabber - starts the search process for chats
func (m *Manager) StartMatcher() {
	m.matcher()
}
