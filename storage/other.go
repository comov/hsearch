package storage

import "time"

// Feedback - write feedback from user
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

// SaveMessage - when we send user or group offer, description or photos, we
//  save this message for subsequent removal from chat, if need.
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
