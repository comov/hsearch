package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/getsentry/sentry-go"

	"github.com/comov/hsearch/structs"
)

type Lalafo struct {
	Site         string
	Host         string
	Target       string
	MainSelector string
}

type MainPageResponse struct {
	Props struct {
		InitialState struct {
			Listing struct {
				ListingFeed struct {
					Items []struct {
						ExternalID uint64 `json:"id"`
						URL        string `json:"url"`
					} `json:"items"`
				} `json:"listingFeed"`
			} `json:"listing"`
		} `json:"initialState"`
	} `json:"props"`
}

func LalafoSite() *Lalafo {
	return &Lalafo{
		Site:         structs.SiteLalafo,
		Host:         "https://lalafo.kg",
		Target:       "https://lalafo.kg/kyrgyzstan/kvartiry/arenda-kvartir/dolgosrochnaya-arenda-kvartir",
		MainSelector: "#__NEXT_DATA__",
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

func (s *Lalafo) GetApartmentsMap(doc *goquery.Document) ApartmentsMap {
	var mapResponse = make(ApartmentsMap, 0)

	doc.Find("#__NEXT_DATA__").Each(func(i int, _s *goquery.Selection) {
		var fromTheNext = new(MainPageResponse)

		err := json.Unmarshal([]byte(_s.Text()), fromTheNext)
		if err != nil {
			sentry.CaptureException(err)
			return
		}

		for _, i := range fromTheNext.Props.InitialState.Listing.ListingFeed.Items {
			mapResponse[i.ExternalID] = fmt.Sprintf("%s%s", s.FullHost(), i.URL)
		}
	})

	return mapResponse
}

// IdFromHref - find apartment Id from URL
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

// ParseNewApartment - parse html and fills the apartment with valid values
func (s *Lalafo) ParseNewApartment(href string, exId uint64, doc *goquery.Document) *structs.Apartment {
	apartment := s.findAndParseJsonApartment(doc)

	isNotBlank := apartment.City != ""
	isNotBishkek := strings.ToLower(apartment.City) != "бишкек"
	if isNotBishkek && isNotBlank {
		return nil
	}

	currency := 0
	switch strings.ToLower(apartment.Currency) {
	case "сом":
		currency = 1
	case "usd":
		currency = 2
	}

	floor, maxFloor := apartment.floor()
	return &structs.Apartment{
		ExternalId:          exId,
		Site:        s.Site,
		Url:         href,
		Topic:       strings.ReplaceAll(apartment.Title, "Сдается квартира: ", ""),
		Price:       apartment.Price,
		Currency:    int32(currency),
		Phone:       apartment.Mobile,
		Rooms:       apartment.rooms(),
		Area:        apartment.area(),
		Floor:       floor,
		MaxFloor:    maxFloor,
		District:    apartment.district(),
		City:        apartment.City,
		RoomType:    "",
		Body:        apartment.Description,
		ImagesCount: int32(len(apartment.Images)),
		ImagesList:  apartment.imagesAsString(),
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
	Item LalafoApartment `json:"item"`
}

const (
	roomsId       = 69
	areaId        = 70
	floorNumberId = 226
	floorTotalId  = 229
	districtId    = 357
)

type LalafoApartment struct {
	Mobile       string `json:"mobile"`
	IsNegotiable bool   `json:"is_negotiable"`
	Params       []struct {
		ID      int         `json:"id"`
		Name    string      `json:"name"`
		Value   interface{} `json:"value"`
		ValueID int         `json:"value_id"`
	} `json:"params"`
	ParamsMap map[int]string
	Price     int32  `json:"price"`
	City      string `json:"city"`
	Currency  string `json:"currency"`
	Title     string `json:"title"`
	Images    []struct {
		OriginalURL string `json:"original_url"`
	} `json:"images"`
	Description string `json:"description"`
}

func (o *LalafoApartment) rooms() int32 {
	r := intRegex.FindAllString(o.ParamsMap[roomsId], -1)
	if len(r) == 0 {
		return 0
	}
	rooms, err := strconv.Atoi(r[0])
	if err != nil {
		log.Printf("[rooms] %s with an error: %s", o.ParamsMap[roomsId], err)
		return 0
	}
	return int32(rooms)
}

func (o *LalafoApartment) area() int32 {
	areaString, ok := o.ParamsMap[areaId]
	if ok {
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

func (o *LalafoApartment) floor() (int32, int32) {
	numberStr, ok := o.ParamsMap[floorNumberId]
	if ok && numberStr != "" {
		number, err := strconv.Atoi(numberStr)
		if err != nil {
			return 0, 0
		}

		totalStr, ok := o.ParamsMap[floorTotalId]
		if ok && totalStr != "" {
			total, err := strconv.Atoi(totalStr)
			if err == nil {
				return int32(number), int32(total)
			}
		}
		return int32(number), 0
	}
	return 0, 0
}

func (o *LalafoApartment) district() string {
	return o.ParamsMap[districtId]
}

func (o *LalafoApartment) paramsToMap() {
	o.ParamsMap = make(map[int]string)
	for _, param := range o.Params {
		o.ParamsMap[param.ID] = strings.TrimSpace(fmt.Sprintf("%v", param.Value))
	}
}

func (o *LalafoApartment) imagesAsString() []string {
	images := make([]string, 0)
	for _, img := range o.Images {
		images = append(images, img.OriginalURL)
	}
	return images
}

func (s *Lalafo) findAndParseJsonApartment(doc *goquery.Document) LalafoApartment {
	foundJson := JsonStruct{}

	doc.Find("#__NEXT_DATA__").Each(func(i int, s *goquery.Selection) {
		err := json.Unmarshal([]byte(s.Text()), &foundJson)
		if err != nil {
			log.Printf("[findAndParseJsonApartment] fail with an error: %s\n", err)
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
