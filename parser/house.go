package parser

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/getsentry/sentry-go"

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
		Target:       "https://www.house.kg/snyat-kvartiru?region=1&town=2&rental_term=3&sort_by=upped_at+desc&page=%d",
		MainSelector: "p.title > a",
	}
}

func (s *House) Name() string {
	return s.Site
}

func (s *House) FullHost() string {
	return s.Host
}

func (s *House) Url() string {
	return fmt.Sprintf(s.Target, 1)
}

func (s *House) Selector() string {
	return s.MainSelector
}

func (s *House) GetApartmentsMap(doc *goquery.Document) ApartmentsMap {
	var mapResponse = DefaultParser(s, doc)

	var lastPage = 1
	doc.Find(".page-link[data-page]").Each(func(i int, _s *goquery.Selection) {
		n, ok := _s.Attr("data-page")
		nInt, err := strconv.Atoi(n)
		if err != nil {
			sentry.CaptureException(err)
			return
		}

		if ok && nInt > lastPage {
			lastPage = nInt
		}
	})

	if lastPage == 1 {
		return mapResponse
	}

	for i := 2; i <= lastPage; i++ {
		doc, err := GetDocumentByUrl(fmt.Sprintf(s.Target, i))
		if err != nil {
			sentry.CaptureException(err)
			continue
		}
		for id, url := range DefaultParser(s, doc) {
			mapResponse[id] = url
		}
	}

	return mapResponse
}

// IdFromHref - find apartment Id from URL
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

// ParseNewApartment - parse html and fills the apartment with valid values
func (s *House) ParseNewApartment(href string, exId uint64, doc *goquery.Document) *structs.Apartment {
	price, currency := s.parsePrice(doc)
	floor, maxFloor := s.floor(doc)
	images := s.parseImages(doc)
	return &structs.Apartment{
		ExternalId:  exId,
		Site:        s.Site,
		Url:         href,
		Topic:       s.parseTitle(doc),
		Price:       price,
		Currency:    currency,
		Phone:       s.parsePhone(doc),
		Area:        s.area(doc),
		Floor:       floor,
		MaxFloor:    maxFloor,
		District:    s.district(doc),
		City:        "Бишкек", //city,
		RoomType:    "",       //roomType,
		Body:        s.parseBody(doc),
		ImagesCount: int32(len(images)),
		ImagesList:  images,
	}
}

// parseTitle - find topic title
func (s *House) parseTitle(doc *goquery.Document) string {
	return strings.TrimSpace(doc.Find(".left > h1").Text())
}

// parsePrice - find price from badge
func (s *House) parsePrice(doc *goquery.Document) (int32, int32) {
	fullPrice := doc.Find(".price-dollar").Text()
	price := 0

	pInt := intRegex.FindAllString(fullPrice, -1)
	if len(pInt) == 1 {
		p, err := strconv.Atoi(pInt[0])
		if err != nil {
			log.Printf("[parsePrice] %s with an error: %s", fullPrice, err)
			return 0, 0
		}
		price = p
	}

	return int32(price), 2
}

func (s *House) floor(doc *goquery.Document) (int32, int32) {
	floor := s.infoContains(doc, "Этаж")
	floor = strings.TrimSpace(strings.Replace(floor, "этаж ", "", -1))
	numData := strings.Split(floor, " из ")
	if len(numData) != 2 {
		return 0, 0
	}

	currentFloor, err := strconv.Atoi(numData[0])
	if err != nil {
		log.Printf("[floor.currentFloor] %s with an error: %s", floor, err)
		return 0, 0
	}

	maxFloor, err := strconv.Atoi(numData[1])
	if err != nil {
		log.Printf("[floor.maxFloor] %s with an error: %s", floor, err)
		return 0, 0
	}

	return int32(currentFloor), int32(maxFloor)
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

// parseBody - find apartment body in page
func (s *House) parseBody(doc *goquery.Document) string {
	return strings.TrimSpace(doc.Find(".description > p").Text())
}

// parseImages - file all attachment in apartment
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

func (s *House) area(doc *goquery.Document) int32 {
	areaString := s.infoContains(doc, "Площадь")
	if areaString != "" {
		r := intRegex.FindAllString(areaString, -1)
		if len(r) >= 1 {
			area, err := strconv.Atoi(r[0])
			if err == nil && area > 10 && area < 299 {
				return int32(area)
			}
		}
	}
	return 0
}
