package background

import (
	"context"

	"github.com/PuerkitoBio/goquery"

	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/parser"
	"github.com/comov/hsearch/structs"
)

type (
	Storage interface {
		WriteApartments(ctx context.Context, apartment []*structs.Apartment) (int, error)
		ReadChatsForMatching(ctx context.Context, enable int) ([]*structs.Chat, error)
		ReadNextApartment(ctx context.Context, chat *structs.Chat) (*structs.Apartment, error)
		CleanFromExistApartments(ctx context.Context, apartments map[uint64]string, siteName string) error

		UpdateSettings(ctx context.Context, chat *structs.Chat) error
	}

	Bot interface {
		SendApartment(ctx context.Context, apartment *structs.Apartment, chat *structs.Chat) error
		SendError(where string, err error, chatId int64)
	}

	Site interface {
		Name() string
		FullHost() string
		Url() string
		Selector() string

		GetApartmentsMap(doc *goquery.Document) parser.ApartmentsMap
		IdFromHref(href string) (uint64, error)
		ParseNewApartment(href string, exId uint64, doc *goquery.Document) *structs.Apartment
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

// StartGrabber - starts the process of finding new apartments
func (m *Manager) StartGrabber() {
	m.grabber()
}

// StartGrabber - starts the search process for chats
func (m *Manager) StartMatcher() {
	m.matcher()
}
