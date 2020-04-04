package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/comov/hsearch/structs"
)

// WriteOffer - records Offer in the database with the pictures and returns Id
//  to the structure.
func (c *Connector) WriteOffer(offer *structs.Offer) error {
	_, err := c.DB.Exec(`INSERT INTO offer (
		id,
		url,
		topic,
		full_price,
		price,
		currency,
		phone,
		room_numbers,
		area,
		city,
		room_type,
		body,
		images,
		created) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
		offer.Id,
		offer.Url,
		offer.Topic,
		offer.FullPrice,
		offer.Price,
		offer.Currency,
		offer.Phone,
		offer.Rooms,
		offer.Area,
		offer.City,
		offer.RoomType,
		offer.Body,
		offer.Images,
		time.Now().Unix(),
	)
	if err != nil && !regexContain.MatchString(err.Error()) {
		return err
	}
	return c.writeImages(strconv.Itoa(int(offer.Id)), offer.ImagesList)
}

// WriteOffers - writes bulk from offers along with pictures to the fd.
func (c *Connector) WriteOffers(offers []*structs.Offer) (int, error) {
	newOffersCount := 0
	// TODO: как видно, сейчас это сделано через простой цикл, но лучше
	//  предоставить это самому хранилищу. Сделать bulk insert, затем запросить
	//  Id по ExtId и записать картины. Не было времени сделать это сразу
	for i := range offers {
		offer := offers[i]
		err := c.WriteOffer(offer)
		if err != nil {
			return newOffersCount, err
		}

		newOffersCount += 1
	}
	return newOffersCount, nil
}

// writeImages - так как картинки храняться в отдельной таблице, то пишем мы их
// отдельно
func (c *Connector) writeImages(offerId string, images []string) error {
	if len(images) <= 0 {
		return nil
	}

	params := make([]interface{}, 0)

	paramsPattern := ""
	sep := ""
	for _, image := range images {
		paramsPattern += sep + "(?, ?)"
		sep = ", "
		params = append(params, offerId, image)
	}

	query := "INSERT INTO image (offer_id, path) VALUES " + paramsPattern
	_, err := c.DB.Exec(query, params...)
	if err != nil && !regexContain.MatchString(err.Error()) {
		return err
	}

	return nil
}

// CleanFromExistOrders - clears the map of offers that are already in
//  the database
func (c *Connector) CleanFromExistOrders(offers map[uint64]string) error {
	params := make([]interface{}, 0)

	paramsPattern := ""
	sep := ""
	for id := range offers {
		paramsPattern += sep + "?"
		sep = ", "
		params = append(params, id)
	}

	query := fmt.Sprintf("SELECT id FROM offer WHERE id IN (%s)", paramsPattern)
	rows, err := c.DB.Query(query, params...)
	if err != nil {
		return err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Println("[CleanFromExistOrders.Close] error:", err)
		}
	}()

	for rows.Next() {
		exId := uint64(0)
		err := rows.Scan(&exId)
		if err != nil {
			log.Println("[CleanFromExistOrders.Scan] error:", err)
			continue
		}

		delete(offers, exId)
	}

	return nil
}

// Dislike - mark offer as bad for user or group and return all message ids
//  (description and photos) for delete from chat.
func (c *Connector) Dislike(msgId int, chatId int64) ([]int, error) {
	offerId := uint64(0)
	msgIds := make([]int, 0)
	err := c.DB.QueryRow(
		`SELECT offer_id
				FROM tg_messages
				WHERE message_id = ?
					AND chat = ?;`,
		msgId,
		chatId,
	).Scan(
		&offerId,
	)
	if err != nil {
		return msgIds, err
	}

	_, _ = c.DB.Exec(
		`INSERT INTO answer (chat, offer_id, dislike, created)
				VALUES (?, ?, ?, ?);`,
		chatId,
		offerId,
		1,
		time.Now().Unix(),
	)

	// load all message with offerId and delete
	rows, err := c.DB.Query(
		`SELECT message_id FROM tg_messages WHERE offer_id = ? AND chat = ?;`,
		offerId,
		chatId,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return msgIds, nil
		}
		return msgIds, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Println("[Dislike] error:", err)
		}
	}()

	for rows.Next() {
		var mId int
		err := rows.Scan(&mId)
		if err != nil {
			log.Println("[Dislike.Scan] error:", err)
			continue
		}
		msgIds = append(msgIds, mId)
	}

	return msgIds, err
}

func (c *Connector) Skip(msgId int, chatId int64) error {
	offerId := uint64(0)
	err := c.DB.QueryRow(
		"SELECT offer_id FROM tg_messages WHERE message_id = ? AND chat = ?;",
		msgId,
		chatId,
	).Scan(
		&offerId,
	)
	if err != nil {
		return err
	}

	skipTime := time.Now().Add(c.skipDelayTime).Unix()
	_, err = c.DB.Exec(
		"INSERT INTO answer (chat, offer_id, skip, created) VALUES (?, ?, ?, ?);",
		chatId,
		offerId,
		skipTime,
		time.Now().Unix(),
	)
	return err
}

func (c *Connector) ReadNextOffer(chat *structs.Chat) (*structs.Offer, error) {
	offer := new(structs.Offer)
	now := time.Now()

	var query strings.Builder
	query.WriteString(`
	SELECT DISTINCT
		of.id,
		of.url,
		of.topic,
		of.full_price,
		of.price,
		of.currency,
		of.phone,
		of.room_numbers,
		of.area,
		of.city,
		of.room_type,
		of.images,
		of.body
	FROM offer of
	LEFT JOIN answer u on (of.id = u.offer_id AND u.chat = ?)
	LEFT JOIN tg_messages sm on (of.id = sm.offer_id AND sm.chat = ?)
	WHERE of.created >= ?
		AND (u.dislike = 0 OR u.dislike IS NULL)
		AND sm.created IS NULL
	`)

	if chat.Photo {
		query.WriteString(" AND of.images != 0")
	}

	if chat.KGS.String() != "0:0" || chat.USD.String() != "0:0" {
		query.WriteString(priceFilter(chat.USD, chat.KGS))
	}

	query.WriteString(" 	ORDER BY of.created;")

	err := c.DB.QueryRow(
		query.String(),
		chat.Id,
		chat.Id,
		now.Add(-c.relevanceTime).Unix(),
	).Scan(
		&offer.Id,
		&offer.Url,
		&offer.Topic,
		&offer.FullPrice,
		&offer.Price,
		&offer.Currency,
		&offer.Phone,
		&offer.Rooms,
		&offer.Area,
		&offer.City,
		&offer.RoomType,
		&offer.Images,
		&offer.Body,
	)

	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}

	return offer, err
}

func priceFilter(usd, kgs structs.Price) string {
	var f strings.Builder
	f.WriteString(" AND(")
	if usd.String() == "0:0" {
		f.WriteString(" of.currency = 'usd'")
	} else {
		f.WriteString(fmt.Sprintf(" (of.price between %d and %d and of.currency = 'usd')", usd[0], usd[1]))
	}

	if kgs.String() == "0:0" {
		f.WriteString(" or of.currency = 'сом'")
	} else {
		f.WriteString(fmt.Sprintf(" or (of.price between %d and %d and of.currency = 'сом')", kgs[0], kgs[1]))
	}

	f.WriteString(" )")
	return f.String()
}

func (c *Connector) ReadOfferDescription(msgId int, chatId int64) (uint64, string, error) {
	offerId := uint64(0)
	err := c.DB.QueryRow(
		"SELECT offer_id FROM tg_messages WHERE message_id = ? AND chat = ?;",
		msgId,
		chatId,
	).Scan(
		&offerId,
	)
	if err != nil {
		return offerId, "", err
	}

	description := ""
	err = c.DB.QueryRow(`SELECT body FROM offer of WHERE of.id = ?;`,
		offerId,
	).Scan(
		&description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return offerId, "Предложение не найдено, возможно было удалено", nil
		}
		return offerId, "", err
	}

	return offerId, description, nil
}

func (c *Connector) ReadOfferImages(msgId int, chatId int64) (uint64, []string, error) {
	offerId := uint64(0)
	images := make([]string, 0)

	err := c.DB.QueryRow(
		"SELECT offer_id FROM tg_messages WHERE message_id = ? AND chat = ?;",
		msgId,
		chatId,
	).Scan(
		&offerId,
	)
	if err != nil {
		return offerId, images, err
	}

	rows, err := c.DB.Query(`SELECT path FROM image im WHERE im.offer_id = ?;`, offerId)
	if err != nil {
		if err == sql.ErrNoRows {
			return offerId, images, nil
		}
		return offerId, images, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Println("[ReadOfferImages] error:", err)
		}
	}()

	for rows.Next() {
		image := ""
		err := rows.Scan(
			&image,
		)
		if err != nil {
			log.Println("[ReadOfferImages.Scan] error:", err)
			continue
		}
		images = append(images, image)
	}

	return offerId, images, nil
}
