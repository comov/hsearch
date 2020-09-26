package parser

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/comov/hsearch/structs"
)

type House struct {
	Site         string
	Host         string
	Target       string
	MainSelector string
}

func HouseSite() *House {
	return &House{
		Site:         structs.SiteHouse,
		Host:         "https://www.house.kg",
		Target:       "https://www.house.kg/snyat-kvartiru?region=1&town=2&rental_term=3&sort_by=upped_at+desc",
		MainSelector: ".left-image > a",
	}
}

func (s *House) Name() string {
	return s.Site
}

func (s *House) FullHost() string {
	return s.Host
}

func (s *House) Url() string {
	return s.Target
}

func (s *House) Selector() string {
	return s.MainSelector
}

// IdFromHref - find offer Id from URL
func (s *House) IdFromHref(href string) (uint64, error) {
	res := strings.Split(href, "-")
	if len(res) == 2 {
		idInt, err := strconv.Atoi(res[1])
		if err != nil {
			return 0, err
		}
		return uint64(idInt), nil
	}
	return 0, fmt.Errorf("can't find id from href %s", href)
}

// ParseNewOffer - parse html and fills the offer with valid values
func (s *House) ParseNewOffer(href string, exId uint64, doc *goquery.Document) *structs.Offer {
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
		Area:       s.area(doc),
		Floor:      s.floor(doc),
		District:   s.district(doc),
		City:       "Бишкек", //city,
		RoomType:   "",       //roomType,
		Body:       s.parseBody(doc),
		Images:     len(images),
		ImagesList: images,
	}
}

// parseTitle - find topic title
func (s *House) parseTitle(doc *goquery.Document) string {
	return strings.TrimSpace(doc.Find(".left > h1").Text())
}

// parsePrice - find price from badge
func (s *House) parsePrice(doc *goquery.Document) (string, int, string) {
	fullPrice := doc.Find(".price-dollar").Text()
	price := 0

	pInt := intRegex.FindAllString(fullPrice, -1)
	if len(pInt) == 1 {
		p, err := strconv.Atoi(pInt[0])
		if err != nil {
			log.Printf("[parsePrice] %s with an error: %s", fullPrice, err)
		}
		price = p
	}

	return fmt.Sprintf("%d USD", price), price, "usd"
}

func (s *House) floor(doc *goquery.Document) string {
	floor := s.infoContains(doc, "Этаж")
	floor = strings.Replace(floor, "этаж ", "", -1)
	return strings.TrimSpace(floor)
}

func (s *House) district(doc *goquery.Document) string {
	district := strings.Replace(doc.Find("div.adress").Text(), "Бишкек, ", "", -1)
	return strings.TrimSpace(district)
}

// parsePhone - find phone number from badge
func (s *House) parsePhone(doc *goquery.Document) string {
	phone := doc.Find(".number").Text()
	phone = strings.Replace(phone, "-", "", -1)
	phone = strings.Replace(phone, " ", "", -1)
	if len(phone) >= 9 {
		phone = fmt.Sprintf("+996%s", phone[len(phone)-9:])
	}
	return phone
}

// spanContains - find text value by contain selector
func (s *House) infoContains(doc *goquery.Document, text string) string {
	nodes := doc.Find("div.label:contains('" + text + "')").Parent().Children().Nodes
	if len(nodes) > 1 {
		return goquery.NewDocumentFromNode(nodes[1]).Text()
	}
	return ""
}

// parseBody - find offer body in page
func (s *House) parseBody(doc *goquery.Document) string {
	return strings.TrimSpace(doc.Find(".description > p").Text())
}

// parseImages - file all attachment in offer
func (s *House) parseImages(doc *goquery.Document) []string {
	images := make([]string, 0)
	doc.Find(".fotorama > a").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("data-full")
		if ok {
			images = append(images, href)
		}
	})
	return images
}

func (s *House) area(doc *goquery.Document) string {
	areaString := s.infoContains(doc, "Площадь")
	if areaString != "" {
		r := intRegex.FindAllString(areaString, -1)
		if len(r) >= 1 {
			area, err := strconv.Atoi(r[0])
			if err == nil && area > 10 && area < 299 {
				return fmt.Sprintf("%d м2", area)
			}
		}
	}
	return ""
}
