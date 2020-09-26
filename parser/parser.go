package parser

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/comov/hsearch/structs"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

type (
	Site interface {
		FullHost() string
		Url() string
		Selector() string
		IdFromHref(href string) (uint64, error)
		ParseNewOffer(href string, exId uint64, doc *goquery.Document) *structs.Offer
	}
)

// Diesel
//  [NOT]: Тема-негативка. Только факты. Арендаторам и Арендодателям внимание!
const negativeTheme = 2477961

var (
	intRegex  = regexp.MustCompile(`\d+`)
	textRegex = regexp.MustCompile(`[a-zA-Zа-яА-Я]+`)
)

// FindOffersLinksOnSite - load new offers from the site and all find offers
func FindOffersLinksOnSite(site Site) (map[uint64]string, error) {
	doc, err := GetDocumentByUrl(site.Url())
	if err != nil {
		return nil, err
	}

	offers := make(map[uint64]string)

	doc.Find(site.Selector()).Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}

		offerId, err := site.IdFromHref(href)
		if err != nil {
			log.Println("Can't get Id from href with an error", err)
			return
		}

		u, err := url.Parse(href)
		if err != nil {
			log.Println("Can't parse href to error with an error", err)
			return
		}

		offers[offerId] = fmt.Sprintf("%s%s", site.FullHost(), u.RequestURI())
	})

	delete(offers, negativeTheme)
	return offers, nil
}

type loadOffers struct {
	offers []*structs.Offer
	add    chan *structs.Offer
	wg     sync.WaitGroup
	ctx    context.Context
}

func (l *loadOffers) loadOffer(site Site, id uint64, href string) {
	defer l.wg.Done()

	doc, err := GetDocumentByUrl(href)
	if err != nil {
		log.Printf("Can't load offer %s with an error %s\f", href, err)
		return
	}

	offer := site.ParseNewOffer(href, id, doc)
	if offer != nil {
		l.add <- offer
	}
}

func (l *loadOffers) addOffer() {
	for {
		select {
		case offer := <-l.add:
			l.offers = append(l.offers, offer)
		case <-l.ctx.Done():
			return
		}
	}
}

// LoadOffersDetail - выгружает и парсит offers по href
func LoadOffersDetail(offersList map[uint64]string, site Site) []*structs.Offer {
	// fixme: это ёбаный костыль!
	lo := loadOffers{
		offers: make([]*structs.Offer, 0),
		add:    make(chan *structs.Offer, len(offersList)),
	}

	ctx, cancel := context.WithCancel(context.Background())
	lo.ctx = ctx
	defer cancel()
	defer close(lo.add)

	go lo.addOffer()

	for id, href := range offersList {
		lo.wg.Add(1)
		go lo.loadOffer(site, id, href)
	}

	lo.wg.Wait()
	time.Sleep(time.Second * 1) // fixme: особенно это. Типа ожидать чтоб добавить в список последний offer
	return lo.offers
}

// GetDocumentByUrl - получает страницу по http, читает и возвращет объект
// goquery.Document для парсинга
func GetDocumentByUrl(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Println("[GetDocumentByUrl.Get] error:", err)
		return nil, err
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Println("[GetDocumentByUrl.defer.Close] error:", err)
		}
	}()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	return goquery.NewDocumentFromReader(res.Body)
}
