package storage

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/comov/hsearch/structs"
)

// WriteApartment - records Apartment in the database with the pictures and returns Id
//  to the structure.
func (c *Connector) WriteApartment(ctx context.Context, apartment *structs.Apartment) error {
	lat, lon := 0.0, 0.0
	err := c.Conn.QueryRow(ctx, `INSERT INTO hsearch_apartment (
		external_id,
		created,
		site,
		url,
		topic,
		price,
		currency,
		phone,
		rooms,
		area,
		floor,
		max_floor,
		district,
		city,
		room_type,
		body,
		images_count,
		lat,
		lon
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19) RETURNING id;`,
		apartment.ExternalId,
		time.Now().Unix(),
		apartment.Site,
		apartment.Url,
		apartment.Topic,
		apartment.Price,
		apartment.Currency,
		apartment.Phone,
		apartment.Rooms,
		apartment.Area,
		apartment.Floor,
		apartment.MaxFloor,
		apartment.District,
		apartment.City,
		apartment.RoomType,
		apartment.Body,
		apartment.ImagesCount,
		lat,
		lon,
	).Scan(&apartment.Id)
	if err != nil && !regexContain.MatchString(err.Error()) {
		return err
	}
	return c.writeImages(ctx, strconv.Itoa(int(apartment.Id)), apartment.ImagesList)
}

// WriteApartments - writes bulk from apartments along with pictures to the fd.
func (c *Connector) WriteApartments(ctx context.Context, apartments []*structs.Apartment) (int, error) {
	newApartmentsCount := 0
	// TODO: как видно, сейчас это сделано через простой цикл, но лучше
	//  предоставить это самому хранилищу. Сделать bulk insert, затем запросить
	//  Id по ExtId и записать картины. Не было времени сделать это сразу
	for i := range apartments {
		apartment := apartments[i]
		err := c.WriteApartment(ctx, apartment)
		if err != nil {
			return newApartmentsCount, err
		}

		newApartmentsCount += 1
	}
	return newApartmentsCount, nil
}

// writeImages - так как картинки хранятся в отдельной таблице, то пишем мы их отдельно
func (c *Connector) writeImages(ctx context.Context, apartmentId string, images []string) error {
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
		params = append(params, apartmentId, image, now)
		paramsNum += 3
	}

	query := "INSERT INTO hsearch_image (apartment_id, path, created) VALUES " + paramsPattern
	_, err := c.Conn.Exec(ctx, query, params...)
	if err != nil && !regexContain.MatchString(err.Error()) {
		return err
	}

	return nil
}

// CleanFromExistApartments - clears the map of offers that are already in
//  the database
func (c *Connector) CleanFromExistApartments(ctx context.Context, apartments map[uint64]string, siteName string) error {
	params := make([]interface{}, 0)

	paramsPattern := ""
	paramsNum := 1
	sep := ""
	for id := range apartments {
		paramsPattern += fmt.Sprintf("%s$%d", sep, paramsNum)
		sep = ", "
		params = append(params, id)
		paramsNum += 1
	}

	params = append(params, siteName)

	query := fmt.Sprintf(`
	SELECT external_id
	FROM hsearch_apartment
	WHERE external_id IN (%s)
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

		delete(apartments, exId)
	}

	return nil
}

// Dislike - mark apartment as bad for user or group and return all message ids
//  (description and photos) for delete from chat.
func (c *Connector) Dislike(ctx context.Context, msgId int, chatId int64) ([]int, error) {
	apartmentId := uint64(0)
	exChatId := uint64(0)
	msgIds := make([]int, 0)
	err := c.Conn.QueryRow(
		ctx,
		`SELECT tgm.apartment_id, hc.id
				FROM hsearch_tgmessage tgm
				left join hsearch_chat hc on hc.id = tgm.chat_id
				WHERE tgm.message_id = $1
					AND hc.chat_id = $2;`,
		msgId,
		chatId,
	).Scan(
		&apartmentId,
		&exChatId,
	)
	if err != nil {
		return msgIds, err
	}

	_, _ = c.Conn.Exec(
		ctx,
		`INSERT INTO hsearch_answer (chat_id, apartment_id, dislike, created) VALUES ($1, $2, $3, $4);`,
		exChatId,
		apartmentId,
		true,
		time.Now().Unix(),
	)

	// load all message with apartmentId and delete
	rows, err := c.Conn.Query(
		ctx,
		`SELECT message_id FROM hsearch_tgmessage WHERE apartment_id = $1 AND chat_id = $2;`,
		apartmentId,
		exChatId,
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

func (c *Connector) ReadNextApartment(ctx context.Context, chat *structs.Chat) (*structs.Apartment, error) {
	apartment := new(structs.Apartment)
	now := time.Now()

	var query strings.Builder
	query.WriteString(`
	SELECT
		of.id,
		of.external_id,
		of.site,
		of.url,
		of.topic,
		of.price,
		of.currency,
		of.phone,
		of.rooms,
		of.area,
		of.floor,
		of.max_floor,
		of.district,
		of.city,
		of.room_type,
		of.body,
		of.images_count
	FROM hsearch_apartment of
	LEFT JOIN hsearch_answer u on (of.id = u.apartment_id AND u.chat_id = $1)
	LEFT JOIN hsearch_tgmessage sm on (of.id = sm.apartment_id AND sm.chat_id = $2)
	WHERE of.created >= $3
		AND (u.dislike is false OR u.dislike IS NULL)
		AND sm.created IS NULL
	`)

	if chat.Photo {
		query.WriteString(" AND of.images_count != 0")
	}

	if chat.KGS.String() != "0:0" || chat.USD.String() != "0:0" {
		query.WriteString(priceFilter(chat.USD, chat.KGS))
	}

	query.WriteString(siteFilter(chat.Diesel, chat.House, chat.Lalafo))
	query.WriteString(" 	ORDER BY of.created;")

	err := c.Conn.QueryRow(
		ctx,
		query.String(),
		chat.ChatId,
		chat.ChatId,
		now.Add(-c.relevanceTime).Unix(),
	).Scan(
		&apartment.Id,
		&apartment.ExternalId,
		&apartment.Site,
		&apartment.Url,
		&apartment.Topic,
		&apartment.Price,
		&apartment.Currency,
		&apartment.Phone,
		&apartment.Rooms,
		&apartment.Area,
		&apartment.Floor,
		&apartment.MaxFloor,
		&apartment.District,
		&apartment.City,
		&apartment.RoomType,
		&apartment.Body,
		&apartment.ImagesCount,
	)

	if err != nil && err == pgx.ErrNoRows {
		return nil, nil
	}

	return apartment, err
}

func priceFilter(usd, kgs structs.Price) string {
	var f strings.Builder
	f.WriteString(" AND(")
	if usd.String() == "0:0" {
		f.WriteString(" of.currency = 1")
	} else {
		f.WriteString(fmt.Sprintf(" (of.price between %d and %d and of.currency = 1)", usd[0], usd[1]))
	}

	if kgs.String() == "0:0" {
		f.WriteString(" or of.currency = 2")
	} else {
		f.WriteString(fmt.Sprintf(" or (of.price between %d and %d and of.currency = 2)", kgs[0], kgs[1]))
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

func (c *Connector) ReadApartmentDescription(ctx context.Context, msgId int, chatId int64) (uint64, string, error) {
	apartmentId := uint64(0)
	err := c.Conn.QueryRow(
		ctx,
		"SELECT tgm.apartment_id FROM hsearch_tgmessage tgm left join hsearch_chat hc on hc.id = tgm.chat_id WHERE message_id = $1 AND hc.chat_id = $2;",
		msgId,
		chatId,
	).Scan(
		&apartmentId,
	)
	if err != nil {
		return apartmentId, "", err
	}

	description := ""
	err = c.Conn.QueryRow(ctx, `SELECT body FROM hsearch_apartment of WHERE of.id = $1;`,
		apartmentId,
	).Scan(
		&description,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return apartmentId, "Предложение не найдено, возможно было удалено", nil
		}
		return apartmentId, "", err
	}

	return apartmentId, description, nil
}

func (c *Connector) ReadApartmentImages(ctx context.Context, msgId int, chatId int64) (uint64, []string, error) {
	apartmentId := uint64(0)
	images := make([]string, 0)

	err := c.Conn.QueryRow(
		ctx,
		"SELECT apartment_id FROM hsearch_tgmessage tgm left join hsearch_chat hc on hc.id = tgm.chat_id WHERE tgm.message_id = $1 AND hc.chat_id = $2;",
		msgId,
		chatId,
	).Scan(
		&apartmentId,
	)
	if err != nil {
		return apartmentId, images, err
	}

	rows, err := c.Conn.Query(ctx, `SELECT path FROM hsearch_image im WHERE im.apartment_id = $1;`, apartmentId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return apartmentId, images, nil
		}
		return apartmentId, images, err
	}

	defer rows.Close()

	for rows.Next() {
		image := ""
		err := rows.Scan(
			&image,
		)
		if err != nil {
			log.Println("[ReadApartmentImages.Scan] error:", err)
			continue
		}
		images = append(images, image)
	}

	return apartmentId, images, nil
}
