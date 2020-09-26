package background

import (
	"context"
	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/parser"
	"github.com/comov/hsearch/structs"

	"github.com/PuerkitoBio/goquery"
)

type (
	Storage interface {
		WriteOffers(ctx context.Context, offer []*structs.Offer) (int, error)
		ReadChatsForMatching(ctx context.Context, enable int) ([]*structs.Chat, error)
		ReadNextOffer(ctx context.Context, chat *structs.Chat) (*structs.Offer, error)
		CleanFromExistOrders(ctx context.Context, offers map[uint64]string, siteName string) error

		// GarbageCollector methods
		CleanExpiredOffers(ctx context.Context, expireDate int64) error
		CleanExpiredImages(ctx context.Context, expireDate int64) error
		CleanExpiredAnswers(ctx context.Context, expireDate int64) error
		CleanExpiredTGMessages(ctx context.Context, expireDate int64) error
	}

	Bot interface {
		SendOffer(ctx context.Context, offer *structs.Offer, chatId int64) error
		SendError(where string, err error, chatId int64)
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

// NewManager - initializes the new background manager
func NewManager(cnf *configs.Config, st Storage, bot Bot) *Manager {
	return &Manager{
		st:  st,
		bot: bot,
		cnf: cnf,
		sitesForParse: []Site{
			parser.DieselSite(),
			parser.HouseSite(),
			parser.LalafoSite(),
		},
	}
}

// StartGarbageCollector - runs garbage collection in the form of old records that no longer make sense
func (m *Manager) StartGarbageCollector(ctx context.Context) {
	m.garbage(ctx)
}

// StartGrabber - starts the process of finding new offers
func (m *Manager) StartGrabber(ctx context.Context) {
	m.grabber(ctx)
}

// StartGrabber - starts the search process for chats
func (m *Manager) StartMatcher(ctx context.Context) {
	m.matcher(ctx)
}
