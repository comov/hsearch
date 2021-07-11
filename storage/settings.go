package storage

import (
	"context"

	"github.com/jackc/pgx/v4"

	"github.com/comov/hsearch/structs"
)

// StartSearch - register new user or group if not exist or enable receive new
//  apartments.
func (c *Connector) StartSearch(ctx context.Context, id int64, username, title, cType string) error {
	chat, err := c.ReadChat(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.CreateChat(ctx, id, username, title, cType, true)
		}
		return err
	}

	_, err = c.Conn.Exec(ctx, "UPDATE hsearch_chat SET enable = 1 WHERE chat_id = $1;", chat.Id)
	return err
}

// UpdateSettings - updates all chat settings
func (c *Connector) UpdateSettings(ctx context.Context, chat *structs.Chat) error {
	_, err := c.Conn.Exec(
		ctx,
		`UPDATE hsearch_chat SET
		enable = $1,
		diesel = $2,
		house = $3,
		lalafo = $4,
		photo = $5,
		kgs = $6,
		usd = $7
	WHERE chat_id = $8
	`,
		chat.Enable,
		chat.Diesel,
		chat.House,
		chat.Lalafo,
		chat.Photo,
		chat.KGS,
		chat.USD,
		chat.Id,
	)
	return err
}
