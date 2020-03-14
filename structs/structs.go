package structs

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	KindOffer       = "offer"
	KindPhoto       = "photo"
	KindDescription = "description"
)

var priceRegex = regexp.MustCompile(`\d+`)
var currencyRegex = regexp.MustCompile(`[a-zA-Zа-яА-Я]+`)

type (
	// Chat - all users and communicate with bot in chats. Chat can be group,
	// supergroup or private (type)
	Chat struct {
		Id       int64
		Username string
		Title    string // in private chats, this field save user full name
		Enable   bool
		Type     string
	}

	// Offer - posted on the site.
	Offer struct {
		Id         uint64
		Created    int64
		Url        string
		Topic      string
		FullPrice  string
		Price      int
		Currency   string // all currency is lower
		Phone      string
		Rooms      string
		Body       string
		Images     int
		ImagesList []string
		doc        *goquery.Document
	}

	// Answer - это ManyToMany для хранения реакции пользователя на
	// объявдение
	Answer struct {
		Created int64
		Chat    uint64
		Offer   uint64
		Like    bool
		Dislike bool
		Skip    uint64
	}

	// Feedback - структура для обратной связи в надежде получать баг репорты
	// а не угрозы что я бизнес чей-то сломал
	Feedback struct {
		Username string
		Chat     int64
		Body     string
	}
)

// TODO: это должно быть в парсере 🤦‍
// ParseNewOffer - заполняет структуру объявления
func ParseNewOffer(href string, exId uint64, doc *goquery.Document) *Offer {
	offer := &Offer{
		Url: href,
		Id:  exId,
		doc: doc,
	}

	offer.parseTitle()
	offer.parsePrice()
	offer.parsePhone()
	offer.parseRoomNumber()
	offer.parseBody()
	offer.parseImages()
	return offer
}

// parseTitle - находит заголовок
func (o *Offer) parseTitle() string {
	o.Topic = o.doc.Find(".ipsType_pagetitle").Text()
	return o.Topic
}

// parsePrice - find price from badge
func (o *Offer) parsePrice() string {
	o.FullPrice = o.doc.Find("span.field-value.badge.badge-green").Text()

	pInt := priceRegex.FindAllString(o.FullPrice, -1)
	if len(pInt) == 1 {
		p, err := strconv.Atoi(pInt[0])
		if err != nil {
			log.Printf("[parsePrice] %s with an error: %s", o.FullPrice, err)
		}
		o.Price = p
	}

	pCurrency := currencyRegex.FindAllString(o.FullPrice, -1)
	if len(pCurrency) == 1 {
		o.Currency = strings.ToLower(pCurrency[0])
	}

	return o.FullPrice
}

// parsePhone - находит номер телефона из настроек обхявления
func (o *Offer) parsePhone() string {
	o.Phone = o.doc.Find(".custom-field.md-phone > span.field-value").Text()
	return o.Phone
}

// parseRoomNumber - находит количество комнат из настроек объявления
func (o *Offer) parseRoomNumber() string {
	roomNumberNodes := o.doc.Find("span:contains('Количество комнат')").Parent().Children().Nodes
	if len(roomNumberNodes) > 1 {
		o.Rooms = goquery.NewDocumentFromNode(roomNumberNodes[1]).Text()
	}
	return o.Rooms
}

// parseBody - находит тело объявления
func (o *Offer) parseBody() string {
	messages := o.doc.Find(".post.entry-content").Nodes
	if len(messages) != 0 {
		body := goquery.NewDocumentFromNode(messages[0]).Text()
		reg := regexp.MustCompile(`Сообщение отредактировал.*`)
		body = reg.ReplaceAllString(body, "${1}")
		body = strings.Replace(body, "Прикрепленные изображения", "", 1)
		body = strings.Replace(body, "  ", "", 1)
		body = strings.TrimSpace(body)
		o.Body = body
	}
	return o.Body
}

// parseImages - находит все прикрепленные файлы в объявлении
func (o *Offer) parseImages() []string {
	o.doc.Find(".attach").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("src")
		if ok {
			o.ImagesList = append(o.ImagesList, href)
			o.Images += 1
		}
	})
	return o.ImagesList
}
