package storage

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/comov/hsearch/structs"

	"github.com/jackc/pgx/v4"
)

// WriteOffer - records Offer in the database with the pictures and returns Id
//  to the structure.
func (c *Connector) WriteOffer(ctx context.Context, offer *structs.Offer) error {
	_, err := c.Conn.Exec(ctx, `INSERT INTO offer (
		id,
		created,
		site,
		url,
		topic,
		full_price,
		price,
		currency,
		phone,
		room_numbers,
		area,
		floor,
		district,
		city,
		room_type,
		body,
		images) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);`,
		offer.Id,
		time.Now().Unix(),
		offer.Site,
		offer.Url,
		offer.Topic,
		offer.FullPrice,
		offer.Price,
		offer.Currency,
		offer.Phone,
		offer.Rooms,
		offer.Area,
		offer.Floor,
		offer.District,
		offer.City,
		offer.RoomType,
		offer.Body,
		offer.Images,
	)
	if err != nil && !regexContain.MatchString(err.Error()) {
		return err
	}
	return c.writeImages(ctx, strconv.Itoa(int(offer.Id)), offer.ImagesList)
}

// WriteOffers - writes bulk from offers along with pictures to the fd.
func (c *Connector) WriteOffers(ctx context.Context, offers []*structs.Offer) (int, error) {
	newOffersCount := 0
	// TODO: как видно, сейчас это сделано через простой цикл, но лучше
	//  предоставить это самому хранилищу. Сделать bulk insert, затем запросить
	//  Id по ExtId и записать картины. Не было времени сделать это сразу
	for i := range offers {
		offer := offers[i]
		err := c.WriteOffer(ctx, offer)
		if err != nil {
			return newOffersCount, err
		}

		newOffersCount += 1
	}
	return newOffersCount, nil
}

// writeImages - так как картинки храняться в отдельной таблице, то пишем мы их
// отдельно
func (c *Connector) writeImages(ctx context.Context, offerId string, images []string) error {
	if len(images) <= 0 {
		return nil
	}

	params := make([]interface{}, 0)
	now := time.Now().Unix()

	paramsPattern := ""
	paramsNum := 1
	sep := ""
	for _, image := range images {
		paramsPattern += sep + fmt.Sprintf("($%d, $%d, $%d)", paramsNum, paramsNum+1, paramsNum+2) // todo: fixed
		sep = ", "
		params = append(params, offerId, image, now)
		paramsNum += 3
	}

	query := "INSERT INTO image (offer_id, path, created) VALUES " + paramsPattern
	_, err := c.Conn.Exec(ctx, query, params...)
	if err != nil && !regexContain.MatchString(err.Error()) {
		return err
	}

	return nil
}

// CleanFromExistOrders - clears the map of offers that are already in
//  the database
func (c *Connector) CleanFromExistOrders(ctx context.Context, offers map[uint64]string, siteName string) error {
	params := make([]interface{}, 0)

	paramsPattern := ""
	paramsNum := 1
	sep := ""
	for id := range offers {
		paramsPattern += fmt.Sprintf("%s$%d", sep, paramsNum)
		sep = ", "
		params = append(params, id)
		paramsNum += 1
	}

	params = append(params, siteName)

	query := fmt.Sprintf(`
	SELECT id
	FROM offer
	WHERE id IN (%s)
		AND site = $%d
	`,
		paramsPattern,
		paramsNum,
	)
	rows, err := c.Conn.Query(ctx, query, params...)
	if err != nil {
		return err
	}

	defer rows.Close()

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
func (c *Connector) Dislike(ctx context.Context, msgId int, chatId int64) ([]int, error) {
	offerId := uint64(0)
	msgIds := make([]int, 0)
	err := c.Conn.QueryRow(
		ctx,
		`SELECT offer_id
				FROM tg_messages
				WHERE message_id = $1
					AND chat = $2;`,
		msgId,
		chatId,
	).Scan(
		&offerId,
	)
	if err != nil {
		return msgIds, err
	}

	_, _ = c.Conn.Exec(
		ctx,
		`INSERT INTO answer (chat, offer_id, dislike, created)
				VALUES ($1, $2, $3, $4);`,
		chatId,
		offerId,
		true,
		time.Now().Unix(),
	)

	// load all message with offerId and delete
	rows, err := c.Conn.Query(
		ctx,
		`SELECT message_id FROM tg_messages WHERE offer_id = $1 AND chat = $2;`,
		offerId,
		chatId,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return msgIds, nil
		}
		return msgIds, err
	}

	defer rows.Close()

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

func (c *Connector) ReadNextOffer(ctx context.Context, chat *structs.Chat) (*structs.Offer, error) {
	offer := new(structs.Offer)
	now := time.Now()

	var query strings.Builder
	query.WriteString(`
	SELECT
		of.id,
		of.site,
		of.url,
		of.topic,
		of.full_price,
		of.price,
		of.currency,
		of.phone,
		of.room_numbers,
		of.area,
		of.city,
		of.floor,
		of.district,
		of.room_type,
		of.images,
		of.body
	FROM offer of
	LEFT JOIN answer u on (of.id = u.offer_id AND u.chat = $1)
	LEFT JOIN tg_messages sm on (of.id = sm.offer_id AND sm.chat = $2)
	WHERE of.created >= $3
		AND (u.dislike is false OR u.dislike IS NULL)
		AND sm.created IS NULL
	`)

	if chat.Photo {
		query.WriteString(" AND of.images != 0")
	}

	if chat.KGS.String() != "0:0" || chat.USD.String() != "0:0" {
		query.WriteString(priceFilter(chat.USD, chat.KGS))
	}

	query.WriteString(siteFilter(chat.Diesel, chat.House, chat.Lalafo))
	query.WriteString(" 	ORDER BY of.created;")

	err := c.Conn.QueryRow(
		ctx,
		query.String(),
		chat.Id,
		chat.Id,
		now.Add(-c.relevanceTime).Unix(),
	).Scan(
		&offer.Id,
		&offer.Site,
		&offer.Url,
		&offer.Topic,
		&offer.FullPrice,
		&offer.Price,
		&offer.Currency,
		&offer.Phone,
		&offer.Rooms,
		&offer.Area,
		&offer.City,
		&offer.Floor,
		&offer.District,
		&offer.RoomType,
		&offer.Images,
		&offer.Body,
	)

	if err != nil && err == pgx.ErrNoRows {
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
		f.WriteString(" or of.currency = 'kgs'")
	} else {
		f.WriteString(fmt.Sprintf(" or (of.price between %d and %d and of.currency = 'kgs')", kgs[0], kgs[1]))
	}

	f.WriteString(" )")
	return f.String()
}

func siteFilter(diesel, house, lalafo bool) string {
	var sites []string
	if diesel {
		sites = append(sites, structs.SiteDiesel)
	}
	if house {
		sites = append(sites, structs.SiteHouse)
	}
	if lalafo {
		sites = append(sites, structs.SiteLalafo)
	}
	switch len(sites) {
	case 1:
		return fmt.Sprintf(" AND of.site == '%s'", sites[0])
	case 2:
		sitesStr, sep := "", ""
		for _, site := range sites {
			sitesStr += fmt.Sprintf("%s'%s'", sep, site)
			sep = ", "
		}
		return fmt.Sprintf(" AND of.site in (%s)", sitesStr)
	}
	return ""
}

func (c *Connector) ReadOfferDescription(ctx context.Context, msgId int, chatId int64) (uint64, string, error) {
	offerId := uint64(0)
	err := c.Conn.QueryRow(
		ctx,
		"SELECT offer_id FROM tg_messages WHERE message_id = $1 AND chat = $2;",
		msgId,
		chatId,
	).Scan(
		&offerId,
	)
	if err != nil {
		return offerId, "", err
	}

	description := ""
	err = c.Conn.QueryRow(ctx, `SELECT body FROM offer of WHERE of.id = $1;`,
		offerId,
	).Scan(
		&description,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return offerId, "Предложение не найдено, возможно было удалено", nil
		}
		return offerId, "", err
	}

	return offerId, description, nil
}

func (c *Connector) ReadOfferImages(ctx context.Context, msgId int, chatId int64) (uint64, []string, error) {
	offerId := uint64(0)
	images := make([]string, 0)

	err := c.Conn.QueryRow(
		ctx,
		"SELECT offer_id FROM tg_messages WHERE message_id = $1 AND chat = $2;",
		msgId,
		chatId,
	).Scan(
		&offerId,
	)
	if err != nil {
		return offerId, images, err
	}

	rows, err := c.Conn.Query(ctx, `SELECT path FROM image im WHERE im.offer_id = $1;`, offerId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return offerId, images, nil
		}
		return offerId, images, err
	}

	defer rows.Close()

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

// CleanExpiredOffers - just clean offer table
func (c *Connector) CleanExpiredOffers(ctx context.Context, expireDate int64) error {
	_, err := c.Conn.Exec(ctx, `DELETE FROM offer WHERE created < $1`, expireDate)
	return err
}

// CleanExpiredImages - just clean image table
func (c *Connector) CleanExpiredImages(ctx context.Context, expireDate int64) error {
	_, err := c.Conn.Exec(ctx, `DELETE FROM image WHERE created < $1`, expireDate)
	return err
}

// CleanExpiredAnswers - just clean answer table
func (c *Connector) CleanExpiredAnswers(ctx context.Context, expireDate int64) error {
	_, err := c.Conn.Exec(ctx, `DELETE FROM answer WHERE created < $1`, expireDate)
	return err
}
