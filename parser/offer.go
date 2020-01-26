package parser

import "github.com/PuerkitoBio/goquery"

type (
	// todo: может стоит использовать структуру из пакета storage?!
	// Offer - временная структура для хранения и парсинга объявления
	Offer struct {
		Url        string
		Title      string
		Price      string
		Phone      string
		RoomNumber string
		Body       string
		Images     []string
		doc        *goquery.Document
	}
)

// ParseNewOffer - заполняет структуру объявления
func ParseNewOffer(href string, doc *goquery.Document) *Offer {
	offer := &Offer{
		Url: href,
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
	o.Title = o.doc.Find(".ipsType_pagetitle").Text()
	return o.Title
}

// parsePrice - находит цену их баджика
func (o *Offer) parsePrice() string {
	o.Price = o.doc.Find("span.field-value.badge.badge-green").Text()
	return o.Price
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
		o.RoomNumber = goquery.NewDocumentFromNode(roomNumberNodes[1]).Text()
	}
	return o.RoomNumber
}

// parseBody - находит тело объявления
func (o *Offer) parseBody() string {
	// todo: нужно почистить от картинок и html тегов
	o.Body = o.doc.Find(".post.entry-content").Text()
	return o.Body
}

// parseImages - находит все прикрепленные файлы в объявлении
func (o *Offer) parseImages() []string {
	o.doc.Find(".attach").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("src")
		if ok {
			o.Images = append(o.Images, href)
		}
	})
	return o.Images
}
