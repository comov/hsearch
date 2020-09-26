package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/comov/hsearch/structs"
)

type Lalafo struct {
	Site         string
	Host         string
	Target       string
	MainSelector string
}

func LalafoSite() *Lalafo {
	return &Lalafo{
		Site:         structs.SiteLalafo,
		Host:         "https://lalafo.kg",
		Target:       "https://lalafo.kg/kyrgyzstan/kvartiry/arenda-kvartir/dolgosrochnaya-arenda-kvartir",
		MainSelector: ".adTile-mainInfo",
	}
}

func (s *Lalafo) Name() string {
	return s.Site
}

func (s *Lalafo) FullHost() string {
	return s.Host
}

func (s *Lalafo) Url() string {
	return s.Target
}

func (s *Lalafo) Selector() string {
	return s.MainSelector
}

// IdFromHref - find offer Id from URL
func (s *Lalafo) IdFromHref(href string) (uint64, error) {
	slice := strings.Split(href, "-")
	if len(slice) == 0 {
		return 0, fmt.Errorf("can't get id from href %s", href)
	}
	idInt, err := strconv.Atoi(slice[len(slice)-1])
	if err != nil {
		return 0, err
	}
	return uint64(idInt), nil
}

// ParseNewOffer - parse html and fills the offer with valid values
func (s *Lalafo) ParseNewOffer(href string, exId uint64, doc *goquery.Document) *structs.Offer {
	offer := s.findAndParseJsonOffer(doc)

	isNotBlank := offer.City != ""
	isNotBishkek := strings.ToLower(offer.City) != "бишкек"
	if isNotBishkek && isNotBlank {
		return nil
	}

	return &structs.Offer{
		Id:         exId,
		Site:       s.Site,
		Url:        href,
		Topic:      strings.ReplaceAll(offer.Title, "Сдается квартира: ", ""),
		FullPrice:  offer.fullPrice(),
		Price:      offer.Price,
		Currency:   strings.ToLower(offer.Currency),
		Phone:      offer.Mobile,
		Rooms:      offer.rooms(),
		Area:       offer.area(),
		Floor:      offer.floor(),
		District:   offer.district(),
		City:       offer.City,
		RoomType:   "",
		Body:       offer.Description,
		Images:     len(offer.Images),
		ImagesList: offer.imagesAsString(),
	}
}

type JsonStruct struct {
	Props struct {
		InitialState struct {
			Feed struct {
				AdDetails map[string]json.RawMessage `json:"adDetails"`
			} `json:"feed"`
		} `json:"initialState"`
	} `json:"props"`
}

type Item struct {
	Item LalafoOffer `json:"item"`
}

const (
	roomsId       = 69
	areaId        = 70
	floorNumberId = 226
	floorTotalId  = 229
	districtId    = 357
)

type LalafoOffer struct {
	Mobile       string `json:"mobile"`
	IsNegotiable bool   `json:"is_negotiable"`
	Params       []struct {
		ID      int         `json:"id"`
		Name    string      `json:"name"`
		Value   interface{} `json:"value"`
		ValueID int         `json:"value_id"`
	} `json:"params"`
	ParamsMap map[int]string
	Price     int    `json:"price"`
	City      string `json:"city"`
	Currency  string `json:"currency"`
	Title     string `json:"title"`
	Images    []struct {
		OriginalURL string `json:"original_url"`
	} `json:"images"`
	Description string `json:"description"`
}

func (o *LalafoOffer) fullPrice() string {
	if o.IsNegotiable {
		return "Договорная"
	}

	var b strings.Builder
	if o.Price != 0 {
		b.WriteString(strconv.Itoa(o.Price))
	}
	if o.Currency != "" {
		b.WriteString(" ")
		b.WriteString(o.Currency)
	}
	return b.String()
}

func (o *LalafoOffer) rooms() string {
	r := intRegex.FindAllString(o.ParamsMap[roomsId], -1)
	if len(r) == 0 {
		return "0"
	}
	return r[0]
}

func (o *LalafoOffer) area() string {
	areaString, ok := o.ParamsMap[areaId]
	if ok {
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

func (o *LalafoOffer) floor() string {
	number, ok := o.ParamsMap[floorNumberId]
	if ok && number != "" {
		total, ok := o.ParamsMap[floorTotalId]
		if ok && total != "" {
			return fmt.Sprintf("%s из %s", number, total)
		}
		return number
	}
	return ""
}

func (o *LalafoOffer) district() string {
	return o.ParamsMap[districtId]
}

func (o *LalafoOffer) paramsToMap() {
	o.ParamsMap = make(map[int]string)
	for _, param := range o.Params {
		o.ParamsMap[param.ID] = strings.TrimSpace(fmt.Sprintf("%v", param.Value))
	}
}

func (o *LalafoOffer) imagesAsString() []string {
	images := make([]string, 0)
	for _, img := range o.Images {
		images = append(images, img.OriginalURL)
	}
	return images
}

func (s *Lalafo) findAndParseJsonOffer(doc *goquery.Document) LalafoOffer {
	foundJson := JsonStruct{}

	doc.Find("#__NEXT_DATA__").Each(func(i int, s *goquery.Selection) {
		err := json.Unmarshal([]byte(s.Text()), &foundJson)
		if err != nil {
			log.Printf("[findAndParseJsonOffer] fail with an error: %s\n", err)
		}
	})

	item := Item{}
	for _, v := range foundJson.Props.InitialState.Feed.AdDetails {
		/* this is hack, because we receive same response
		"adDetails": {
			"70426297": {"item": {}},
			"currentAdId": 70426297
		}
		*/
		_ = json.Unmarshal(v, &item)
	}
	item.Item.paramsToMap()
	return item.Item
}
