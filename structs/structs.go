package structs

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

const (
	TypePrivate    = "private"
	TypeGroup      = "group"
	TypeSupergroup = "supergroup"
	TypeChannel    = "channel"

	KindApartment   = "apartment"
	KindPhoto       = "photo"
	KindDescription = "description"

	SiteDiesel = "diesel"
	SiteLalafo = "lalafo"
	SiteHouse  = "house"
)

type (
	// Price - is a custom type for storing the filter as a string.
	Price [2]int // {from, to}

	// Chat - all users and communicate with bot in chats. Chat can be group,
	//  supergroup or private (type).
	Chat struct {
		// information
		Id       int64
		ChatId   int64
		Username string
		Title    string // in private chats, this field save user full name
		Type     string
		Created  int64

		// settings
		Enable bool
		Diesel bool
		Lalafo bool
		House  bool

		// filters
		Photo bool
		USD   Price
		KGS   Price
	}

	// Apartment - posted on the site.
	Apartment struct {
		Id          uint64
		ExternalId  uint64
		Created     int64
		Site        string
		Url         string
		Topic       string
		Price       int32
		Currency    int32
		Phone       string
		Rooms       int32
		Area        int32
		Floor       int32
		MaxFloor    int32
		District    string
		City        string
		RoomType    string
		Body        string
		ImagesCount int32
		ImagesList  []string
		Lat         float64
		Lon         float64
	}

	// Answer - is a ManyToMany to store the user's reaction to the apartment.
	Answer struct {
		Created   int64
		Chat      uint64
		Apartment uint64
		Dislike   bool
	}

	// Feedback - a feedback structure hoping to get bug reports and not
	//  threats that I broke someone's business.
	Feedback struct {
		Username string
		Chat     int64
		Body     string
	}
)

// String - displays how the price was written in bd
func (p Price) String() string {
	return fmt.Sprintf("%d:%d", p[0], p[1])
}

// Value - leads to the format we need while saving the filter at a price.
func (p Price) Value() (driver.Value, error) {
	return p.String(), nil
}

// Scan - we read a line from the database and translate it into a Go object
//  as can work.
func (p *Price) Scan(value interface{}) error {
	v := value.(string)
	prices := strings.Split(v, ":")
	from, _ := strconv.Atoi(prices[0])
	to, _ := strconv.Atoi(prices[1])
	*p = Price{from, to}
	return nil
}

func (p *Chat) IsChannel() bool {
	return p.Type == TypeChannel
}

var priceMap = map[int32]string{
	1: "USD",
	2: "KGS",
}

func (a *Apartment) GetFullPrice() string {
	if a.Price > 0 {
		return fmt.Sprintf("%d %s", a.Price, priceMap[a.Price])
	}
	return ""
}
