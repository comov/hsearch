package parser

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/comov/hsearch/structs"
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

func (s *Diesel) GetApartmentsMap(doc *goquery.Document) ApartmentsMap {
	return ApartmentsMap{}
}

// IdFromHref - find apartment Id from URL
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

// ParseNewApartment - parse html and fills the apartment with valid values
func (s *Diesel) ParseNewApartment(href string, exId uint64, doc *goquery.Document) *structs.Apartment {
	roomType := s.spanContainsStr(doc, "Тип помещения")
	isNotBlank := roomType != ""
	isNotFlat := strings.ToLower(roomType) != "квартира"
	if isNotBlank && isNotFlat {
		return nil
	}

	city := s.spanContainsStr(doc, "Город:")
	isNotBlank = city != ""
	isNotBishkek := strings.ToLower(city) != "бишкек"
	if isNotBlank && isNotBishkek {
		return nil
	}

	price, currency := s.parsePrice(doc)
	images := s.parseImages(doc)
	return &structs.Apartment{
		ExternalId:  exId,
		Site:        s.Site,
		Url:         href,
		Topic:       s.parseTitle(doc),
		Price:       price,
		Currency:    currency,
		Phone:       s.parsePhone(doc),
		Rooms:       s.spanContains(doc, "Количество комнат"),
		Area:        s.spanContains(doc, "Площадь (кв.м.)"),
		Floor:       0,
		District:    "",
		City:        city,
		RoomType:    roomType,
		Body:        s.parseBody(doc),
		ImagesCount: int32(len(images)),
		ImagesList:  images,
	}
}

// parseTitle - find topic title
func (s *Diesel) parseTitle(doc *goquery.Document) string {
	return doc.Find(".ipsType_pagetitle").Text()
}

// parsePrice - find price from badge
func (s *Diesel) parsePrice(doc *goquery.Document) (int32, int32) {
	fullPrice := doc.Find("span.field-value.badge.badge-green").Text()
	price := 0
	currencyStr := ""

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
		currencyStr = strings.ToLower(pCurrency[0])
	}

	currency := 0
	switch currencyStr {
	case "сом":
		currency = 2
	case "usd":
		currency = 1
	}

	return int32(price), int32(currency)
}

// parsePhone - find phone number from badge
func (s *Diesel) parsePhone(doc *goquery.Document) string {
	phone := doc.Find(".custom-field.md-phone > span.field-value").Text()
	if len(phone) >= 9 {
		phone = fmt.Sprintf("+996%s", phone[len(phone)-9:])
	}
	return phone
}

// spanContainsStr - find text value by contain selector
func (s *Diesel) spanContainsStr(doc *goquery.Document, text string) string {
	nodes := doc.Find("span:contains('" + text + "')").Parent().Children().Nodes
	if len(nodes) > 1 {
		return goquery.NewDocumentFromNode(nodes[1]).Text()
	}
	return ""
}

// spanContains - find text value by contain selector
func (s *Diesel) spanContains(doc *goquery.Document, text string) int32 {
	nodes := doc.Find("span:contains('" + text + "')").Parent().Children().Nodes
	if len(nodes) > 1 {
		str := goquery.NewDocumentFromNode(nodes[1]).Text()
		p, err := strconv.Atoi(str)
		if err != nil {
			log.Printf("[spanContains] %s with an error: %s", str, err)
			return 0
		}
		return int32(p)
	}
	return 0
}

// parseBody - find apartment body in page
func (s *Diesel) parseBody(doc *goquery.Document) string {
	messages := doc.Find(".post.entry-content").Nodes
	body := ""
	if len(messages) != 0 {
		body = goquery.NewDocumentFromNode(messages[0]).Text()
		reg := regexp.MustCompile(`Сообщение отредактировал.*`)
		body = reg.ReplaceAllString(body, "${1}")
		body = strings.Replace(body, "Прикрепленные изображения", "", -1)
		body = strings.Replace(body, "  ", "", -1)
		body = strings.TrimSpace(body)
	}
	return body
}

// parseImages - file all attachment in apartment
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
