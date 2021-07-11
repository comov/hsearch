package storage

import (
	"context"
	"time"
)

// Feedback - write feedback from user
func (c *Connector) Feedback(ctx context.Context, chatId int64, username, body string) error {
	chat, _ := c.ReadChat(ctx, chatId)
	_, err := c.Conn.Exec(
		ctx,
		"INSERT INTO hsearch_feedback (created, chat_id, username, body) VALUES ($1, $2, $3, $4);",
		time.Now().Unix(),
		chat.Id,
		username,
		body,
	)
	if err != nil && !regexContain.MatchString(err.Error()) {
		return err
	}
	return nil
}

// SaveMessage - when we send user or group apartment, description or photos, we
//  save this message for subsequent removal from chat, if need.
func (c *Connector) SaveMessage(ctx context.Context, msgId int, apartmentId uint64, chatId int64, kind string) error {
	chat, _ := c.ReadChat(ctx, chatId)
	_, err := c.Conn.Exec(
		ctx,
		`INSERT INTO hsearch_tgmessage (message_id, apartment_id, kind, chat_id, created) VALUES ($1, $2, $3, $4, $5);`,
		msgId,
		apartmentId,
		kind,
		chat.Id,
		time.Now().Unix(),
	)

	return err
}
