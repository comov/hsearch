package storage

import (
	"context"
	"database/sql"

	"github.com/comov/hsearch/structs"
)

// StopSearch - disable receive message about new offers.
func (c *Connector) StopSearch(ctx context.Context, id int64) error {
	resp, err := c.Conn.Exec(ctx, "UPDATE chat SET enable = 0 WHERE id = $1;", id)
	if err != nil {
		return err
	}

	affect := resp.RowsAffected()
	if affect == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// StartSearch - register new user or group if not exist or enable receive new
//  offers.
func (c *Connector) StartSearch(ctx context.Context, id int64, username, title, cType string) error {
	chat, err := c.ReadChat(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.CreateChat(ctx, id, username, title, cType)
		}
		return err
	}

	_, err = c.Conn.Exec(ctx, "UPDATE chat SET enable = 1 WHERE id = $1;", chat.Id)
	return err
}

// UpdateSettings - updates all chat settings
func (c *Connector) UpdateSettings(ctx context.Context, chat *structs.Chat) error {
	_, err := c.Conn.Exec(
		ctx,
		`UPDATE chat SET
		enable = $1,
		diesel = $2,
		house = $3,
		lalafo = $4,
		photo = $5,
		kgs = $6,
		usd = $7
	WHERE id = $8
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
