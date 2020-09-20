package storage

import (
	"context"
	"time"
)

// Feedback - write feedback from user
func (c *Connector) Feedback(ctx context.Context, chat int64, username, body string) error {
	_, err := c.Conn.Exec(
		ctx,
		"INSERT INTO feedback (created, chat, username, body) VALUES ($1, $2, $3, $4);",
		time.Now().Unix(),
		chat,
		username,
		body,
	)
	if err != nil && !regexContain.MatchString(err.Error()) {
		return err
	}
	return nil
}

// SaveMessage - when we send user or group offer, description or photos, we
//  save this message for subsequent removal from chat, if need.
func (c *Connector) SaveMessage(ctx context.Context, msgId int, offerId uint64, chat int64, kind string) error {
	_, err := c.Conn.Exec(
		ctx,
		`INSERT INTO tg_messages (message_id, offer_id, kind, chat, created) VALUES ($1, $2, $3, $4, $5);`,
		msgId,
		offerId,
		kind,
		chat,
		time.Now().Unix(),
	)

	return err
}

// CleanExpiredTGMessages - just clean tg_messages table
func (c *Connector) CleanExpiredTGMessages(ctx context.Context, expireDate int64) error {
	_, err := c.Conn.Exec(ctx, `DELETE FROM tg_messages WHERE created < $1`, expireDate)
	return err
}
