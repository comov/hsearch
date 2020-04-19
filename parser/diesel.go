package parser

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/comov/hsearch/structs"

	"github.com/PuerkitoBio/goquery"
)

type Diesel struct {
	Site         string
	Host         string
	Target       string
	MainSelector string
}

func DieselSite() *Diesel {
	return &Diesel{
		Site:         structs.SiteDiesel,
		Host:         "http://diesel.elcat.kg",
		Target:       "http://diesel.elcat.kg/index.php?showforum=305",
		MainSelector: ".topic_title",
	}
}

func (s *Diesel) Name() string {
	return s.Site
}

func (s *Diesel) FullHost() string {
	return s.Host
}

func (s *Diesel) Url() string {
	return s.Target
}

func (s *Diesel) Selector() string {
	return s.MainSelector
}

// IdFromHref - find offer Id from URL
func (s *Diesel) IdFromHref(href string) (uint64, error) {
	urlPath, err := url.Parse(href)
	if err != nil {
		return 0, err
	}
	id := urlPath.Query().Get("showtopic")
	if id == "" {
		return 0, fmt.Errorf("id not contain in href")
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return uint64(idInt), nil
}

// ParseNewOffer - parse html and fills the offer with valid values
func (s *Diesel) ParseNewOffer(href string, exId uint64, doc *goquery.Document) *structs.Offer {
	roomType := s.spanContains(doc, "Тип помещения")
	isNotBlank := roomType != ""
	isNotFlat := strings.ToLower(roomType) != "квартира"
	if isNotBlank && isNotFlat {
		return nil
	}

	city := s.spanContains(doc, "Город:")
	isNotBlank = city != ""
	isNotBishkek := strings.ToLower(city) != "бишкек"
	if isNotBlank && isNotBishkek {
		return nil
	}

	fullPrice, price, currency := s.parsePrice(doc)
	images := s.parseImages(doc)
	return &structs.Offer{
		Id:         exId,
		Site:       s.Site,
		Url:        href,
		Topic:      s.parseTitle(doc),
		FullPrice:  fullPrice,
		Price:      price,
		Currency:   currency,
		Phone:      s.parsePhone(doc),
		Rooms:      s.spanContains(doc, "Количество комнат"),
		Area:       s.spanContains(doc, "Площадь (кв.м.)"),
		Floor:      "",
		District:   "",
		City:       city,
		RoomType:   roomType,
		Body:       s.parseBody(doc),
		Images:     len(images),
		ImagesList: images,
	}
}

// parseTitle - find topic title
func (s *Diesel) parseTitle(doc *goquery.Document) string {
	return doc.Find(".ipsType_pagetitle").Text()
}

// parsePrice - find price from badge
func (s *Diesel) parsePrice(doc *goquery.Document) (string, int, string) {
	fullPrice := doc.Find("span.field-value.badge.badge-green").Text()
	price := 0
	currency := ""

	pInt := intRegex.FindAllString(fullPrice, -1)
	if len(pInt) == 1 {
		p, err := strconv.Atoi(pInt[0])
		if err != nil {
			log.Printf("[parsePrice] %s with an error: %s", fullPrice, err)
		}
		price = p
	}

	pCurrency := textRegex.FindAllString(fullPrice, -1)
	if len(pCurrency) == 1 {
		currency = strings.ToLower(pCurrency[0])
	}

	if currency == "сом" {
		currency = "kgs"
		fullPrice = fmt.Sprintf("%d %s", price, strings.ToUpper(currency))
	}

	return fullPrice, price, currency
}

// parsePhone - find phone number from badge
func (s *Diesel) parsePhone(doc *goquery.Document) string {
	phone := doc.Find(".custom-field.md-phone > span.field-value").Text()
	if len(phone) >= 9 {
		phone = fmt.Sprintf("+996%s", phone[len(phone)-9:])
	}
	return phone
}

// spanContains - find text value by contain selector
func (s *Diesel) spanContains(doc *goquery.Document, text string) string {
	nodes := doc.Find("span:contains('" + text + "')").Parent().Children().Nodes
	if len(nodes) > 1 {
		return goquery.NewDocumentFromNode(nodes[1]).Text()
	}
	return ""
}

// parseBody - find offer body in page
func (s *Diesel) parseBody(doc *goquery.Document) string {
	messages := doc.Find(".post.entry-content").Nodes
	body := ""
	if len(messages) != 0 {
		body = goquery.NewDocumentFromNode(messages[0]).Text()
		reg := regexp.MustCompile(`Сообщение отредактировал.*`)
		body = reg.ReplaceAllString(body, "${1}")
		body = strings.Replace(body, "Прикрепленные изображения", "", 1)
		body = strings.Replace(body, "  ", "", 1)
		body = strings.TrimSpace(body)
	}
	return body
}

// parseImages - file all attachment in offer
func (s *Diesel) parseImages(doc *goquery.Document) []string {
	images := make([]string, 0)
	doc.Find(".attach").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("src")
		if ok {
			images = append(images, href)
		}
	})
	return images
}
