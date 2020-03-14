package storage

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/structs"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose"
)

type (
	// Connector - структура для храниения и управления подключением к бд
	Connector struct {
		DB              *sql.DB
		skipTime        time.Duration
		freshOffersTime time.Duration
	}
)

var regexContain = regexp.MustCompile(`UNIQUE constraint failed*`)

// New - возвращает коннектор для подключения к базе данных. Код не должен знать
// какая бд или какой драйвер используется для работы с базой.
func New(cnf *configs.Config) (*Connector, error) {
	db, err := sql.Open("sqlite3", "hsearch.db?cache=shared")
	if err != nil {
		return nil, err
	}

	// Для БД sqlite максимальное кол-во соединений желательно иметь 1, так как
	// если писать в бд будут 2 потока, то файл может быть покрашен, что несет
	// потерб данных. Это не production проет, по этому нет смысла в PG ил MySql
	db.SetMaxOpenConns(1)

	return &Connector{
		DB:              db,
		skipTime:        cnf.SkipTime,
		freshOffersTime: cnf.FreshOffers,
	}, nil
}

func (c *Connector) Migrate(path string) error {
	err := goose.SetDialect("sqlite3")
	if err != nil {
		return err
	}
	err = goose.Run("up", c.DB, path)
	if err == goose.ErrNoNextVersion {
		return nil
	}

	return err
}

// WriteOffer - записывает Offer в базу вместе с картинками и вазвращает Id в
// структуру
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
    	body,
		images,
		created) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`,
		offer.Id,
		offer.Url,
		offer.Topic,
		offer.FullPrice,
		offer.Price,
		offer.Currency,
		offer.Phone,
		offer.Rooms,
		offer.Body,
		offer.Images,
		time.Now().Unix(),
	)
	if err != nil && !regexContain.MatchString(err.Error()) {
		return err
	}
	return c.writeImages(strconv.Itoa(int(offer.Id)), offer.ImagesList)
}

// WriteOffers - пишет пачку из offers вместе с картинками в бд
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

// CleanFromExistOrders - очищает map от offers которые уже есть в базе
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

// ReadChatsForMatching - read all the chats for which the mailing list has
// to be done
func (c *Connector) ReadChatsForMatching() ([]*structs.Chat, error) {
	rows, err := c.DB.Query(`
	SELECT DISTINCT
		c.id,
		c.username,
		c.title
	FROM chat c
	LEFT JOIN answer uto on (c.id = uto.chat and uto.dislike = 0)
	WHERE c.enable = 1;`)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Println("[ReadChatsForMatching.Close] error:", err)
		}
	}()

	chats := make([]*structs.Chat, 0)
	for rows.Next() {
		chat := new(structs.Chat)
		err := rows.Scan(
			&chat.Id,
			&chat.Username,
			&chat.Title,
		)
		if err != nil {
			log.Println("[ReadChatsForMatching.Scan] error:", err)
			continue
		}

		chats = append(chats, chat)
	}

	return chats, nil
}

// StartSearch - register new user or group if not exist or enable receive new
// offers.
func (c *Connector) StartSearch(id int64, username, title, cType string) error {
	chat, err := c.ReadChat(id)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err := c.DB.Exec(
				`INSERT INTO chat (id, username, title, enable, c_type)
						VALUES (?, ?, ?, ?, ?);`,
				id,
				username,
				title,
				true,
				cType,
			)
			return err
		}
		return err
	}

	_, err = c.DB.Exec("UPDATE chat SET enable = 1 WHERE id = ?;", chat.Id)
	return err
}

// ReadChat - return chat with user or with group if exist or return error.
func (c *Connector) ReadChat(id int64) (*structs.Chat, error) {
	chat := &structs.Chat{}
	err := c.DB.QueryRow(
		`SELECT id, username, title, enable, c_type
				FROM chat
				WHERE id = ?;`,
		id,
	).Scan(
		&chat.Id,
		&chat.Username,
		&chat.Title,
		&chat.Enable,
		&chat.Type,
	)
	return chat, err
}

// StopSearch - disable receive message about new offers.
func (c *Connector) StopSearch(id int64) error {
	resp, err := c.DB.Exec("UPDATE chat SET enable = 0 WHERE id = ?;", id)
	if err != nil {
		return err
	}

	affect, err := resp.RowsAffected()
	if err != nil {
		return err
	}

	if affect == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SaveMessage - when we send user or group offer, description or photos, we
// save this message for subsequent removal from chat, if need.
func (c *Connector) SaveMessage(msgId int, offerId uint64, chat int64, kind string) error {
	_, err := c.DB.Exec(
		`INSERT INTO tg_messages (message_id, offer_id, kind, chat, created)
 				VALUES (?, ?, ?, ?, ?);`,
		msgId,
		offerId,
		kind,
		chat,
		time.Now().Unix(),
	)

	return err
}

// Dislike - mark offer as bad for user or group and return all message ids
// (description and photos) for delete from chat.
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

	_, err = c.DB.Exec(
		`INSERT INTO answer (chat, offer_id, dislike, created)
				VALUES (?, ?, ?, ?);`,
		chatId,
		offerId,
		1,
		time.Now().Unix(),
	)

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

	skipTime := time.Now().Add(c.skipTime).Unix()
	_, err = c.DB.Exec(
		"INSERT INTO answer (chat, offer_id, skip, created) VALUES (?, ?, ?, ?);",
		chatId,
		offerId,
		skipTime,
		time.Now().Unix(),
	)
	return err
}

func (c *Connector) ReadNextOffer(chatId int64) (*structs.Offer, error) {
	offer := new(structs.Offer)
	now := time.Now()

	err := c.DB.QueryRow(`
	SELECT DISTINCT
	   id,
       url,
       topic,
       full_price,
       price,
       currency,
       phone,
       room_numbers,
	   images,
	   body
	FROM offer of
	LEFT JOIN answer u on (of.id = u.offer_id AND u.chat = ?)
	LEFT JOIN tg_messages sm on (of.id = sm.offer_id AND sm.chat = ?)
	WHERE of.created >= ?
	  AND (u.dislike = 0 OR u.dislike IS NULL)
	  AND sm.created IS NULL
	ORDER BY of.created;
	`,
		chatId,
		chatId,
		now.Add(-c.freshOffersTime).Unix(),
	).Scan(
		&offer.Id,
		&offer.Url,
		&offer.Topic,
		&offer.FullPrice,
		&offer.Price,
		&offer.Currency,
		&offer.Phone,
		&offer.Rooms,
		&offer.Images,
		&offer.Body,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return offer, nil
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

func (c *Connector) Feedback(chat int64, username, body string) error {
	_, err := c.DB.Exec(
		"INSERT INTO feedback (created, chat, username, body) VALUES (?, ?, ?, ?);",
		time.Now().Unix(),
		chat,
		username,
		body,
	)
	return err
}
