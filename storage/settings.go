package storage

import (
	"database/sql"

	"github.com/comov/hsearch/structs"
)

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

// StartSearch - register new user or group if not exist or enable receive new
//  offers.
func (c *Connector) StartSearch(id int64, username, title, cType string) error {
	chat, err := c.ReadChat(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.CreateChat(id, username, title, cType)
		}
		return err
	}

	_, err = c.DB.Exec("UPDATE chat SET enable = 1 WHERE id = ?;", chat.Id)
	return err
}

// UpdateSettings - updates all chat settings
func (c *Connector) UpdateSettings(chat *structs.Chat) error {
	_, err := c.DB.Exec(`
	UPDATE chat SET
		enable = ?,
		photo = ?,
		kgs = ?,
		usd = ?
	WHERE id = ?
	`,
		chat.Enable,
		chat.Photo,
		chat.KGS,
		chat.USD,
		chat.Id,
	)
	return err
}
