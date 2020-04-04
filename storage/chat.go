package storage

import (
	"log"
	"strings"
	"time"

	"github.com/comov/hsearch/structs"
)

// todo: remove after update chat info
func (c *Connector) UpdateChat(id int64, title, cType string) error {
	_, err := c.DB.Exec("UPDATE chat SET title = ?, c_type = ? WHERE id = ?",
		title,
		cType,
		id,
	)
	return err
}

// CreateChat - creates a chat room and sets the default settings.
func (c *Connector) CreateChat(id int64, username, title, cType string) error {
	_, err := c.DB.Exec(
		`INSERT INTO chat (id, username, title, enable, c_type, created)
						VALUES (?, ?, ?, ?, ?, ?);`,
		id,
		username,
		title,
		true,
		cType,
		time.Now().Unix(),
	)
	return err
}

// ReadChat - return chat with user or with group if exist or return error.
func (c *Connector) ReadChat(id int64) (*structs.Chat, error) {
	chat := &structs.Chat{}
	err := c.DB.QueryRow(`
	SELECT
		id,
		username,
		title,
		c_type,
		created,
		enable,
		diesel,
		lalafo,
		photo,
		usd,
		kgs,
		up_track
	FROM chat
	WHERE id = ?;
	`,
		id,
	).Scan(
		&chat.Id,
		&chat.Username,
		&chat.Title,
		&chat.Type,
		&chat.Created,
		&chat.Enable,
		&chat.Diesel,
		&chat.Lalafo,
		&chat.Photo,
		&chat.USD,
		&chat.KGS,
		&chat.UpTrack,
	)
	return chat, err
}

// ReadChatsForMatching - read all the chats for which the mailing list has
//  to be done
func (c *Connector) ReadChatsForMatching(enable int) ([]*structs.Chat, error) {
	var query strings.Builder
	query.WriteString(`
	SELECT DISTINCT
		c.id,
		c.username,
		c.title,
		c.c_type,
		c.created,
		c.enable,
		c.diesel,
		c.lalafo,
		c.photo,
		c.usd,
		c.kgs,
		c.up_track
	FROM chat c
`)

	switch enable {
	case 1:
		query.WriteString(" WHERE c.enable = 1")
	case 0:
		query.WriteString(" WHERE c.enable = 0")
	}

	rows, err := c.DB.Query(query.String())
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
			&chat.Type,
			&chat.Created,
			&chat.Enable,
			&chat.Diesel,
			&chat.Lalafo,
			&chat.Photo,
			&chat.USD,
			&chat.KGS,
			&chat.UpTrack,
		)
		if err != nil {
			log.Println("[ReadChatsForMatching.Scan] error:", err)
			continue
		}

		chats = append(chats, chat)
	}

	return chats, nil
}
