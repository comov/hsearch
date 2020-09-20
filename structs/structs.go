package structs

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

const (
	KindOffer       = "offer"
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

	// Offer - posted on the site.
	Offer struct {
		Id         uint64
		Created    int64
		Site       string
		Url        string
		Topic      string
		FullPrice  string
		Price      int
		Currency   string // all currency is lower
		Phone      string
		Rooms      string
		Area       string
		Floor      string
		District   string
		City       string
		RoomType   string
		Body       string
		Images     int
		ImagesList []string
	}

	// Answer - is a ManyToMany to store the user's reaction to the offer.
	Answer struct {
		Created int64
		Chat    uint64
		Offer   uint64
		Dislike bool
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
