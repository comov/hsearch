package parser

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/PuerkitoBio/goquery"

	"github.com/comov/hsearch/structs"
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

// LoadOffersDetail - downloads the offers and provides a ready list from the offers structures
func LoadOffersDetail(offersList map[uint64]string, site Site) []*structs.Offer {
	var offers []*structs.Offer
	var wg sync.WaitGroup

	wg.Add(len(offersList))
	for id, href := range offersList {
		go func(site Site, id uint64, href string) {
			defer wg.Done()

			doc, err := GetDocumentByUrl(href)
			if err != nil {
				log.Printf("Can't load offer %s with an error %s\f", href, err)
				return
			}

			offer := site.ParseNewOffer(href, id, doc)
			if offer != nil {
				offers = append(offers, offer)
			}
		}(site, id, href)
	}

	wg.Wait()
	return offers
}

// GetDocumentByUrl - receives the page by http, reads and returns the goquery.Document object
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
