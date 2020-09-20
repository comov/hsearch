package storage

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/comov/hsearch/structs"
)

// CreateChat - creates a chat room and sets the default settings.
func (c *Connector) CreateChat(ctx context.Context, id int64, username, title, cType string) error {
	_, err := c.Conn.Exec(
		ctx,
		`INSERT INTO chat (id, username, title, enable, c_type, created) VALUES ($1, $2, $3, $4, $5, $6);`,
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
func (c *Connector) ReadChat(ctx context.Context, id int64) (*structs.Chat, error) {
	chat := &structs.Chat{}
	err := c.Conn.QueryRow(
		ctx, `SELECT
		id,
		username,
		title,
		c_type,
		created,
		enable,
		diesel,
		house,
		lalafo,
		photo,
		usd,
		kgs
	FROM chat
	WHERE id = $1
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
		&chat.House,
		&chat.Lalafo,
		&chat.Photo,
		&chat.USD,
		&chat.KGS,
	)
	return chat, err
}

// ReadChatsForMatching - read all the chats for which the mailing list has
//  to be done
func (c *Connector) ReadChatsForMatching(ctx context.Context, enable int) ([]*structs.Chat, error) {
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
		c.house,
		c.lalafo,
		c.photo,
		c.usd,
		c.kgs
	FROM chat c
`)

	switch enable {
	case 1:
		query.WriteString(" WHERE c.enable is true")
	case 0:
		query.WriteString(" WHERE c.enable is false")
	}

	rows, err := c.Conn.Query(ctx, query.String())
	if err != nil {
		return nil, err
	}

	defer rows.Close()

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
			&chat.House,
			&chat.Lalafo,
			&chat.Photo,
			&chat.USD,
			&chat.KGS,
		)
		if err != nil {
			log.Println("[ReadChatsForMatching.Scan] error:", err)
			continue
		}

		chats = append(chats, chat)
	}

	return chats, nil
}
