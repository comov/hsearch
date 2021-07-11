package storage

import (
	"context"
	"regexp"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/comov/hsearch/configs"
)

type (
	// Connector - the interface to the storage.
	Connector struct {
		ctx           context.Context
		Conn          *pgxpool.Pool
		relevanceTime time.Duration
	}
)

// regexContain - database does not contain a certain type of error, so we have
//  to search through the text to understand what error was caused.
var regexContain = regexp.MustCompile(`ERROR: duplicate key value violates unique constraint*`)

// New - creates a connection to the base and returns the interface to work
//  with the storage.
func New(ctx context.Context, cnf *configs.Config) (*Connector, error) {
	conn, err := pgxpool.Connect(ctx, cnf.PgConnString)
	if err != nil {
		return nil, err
	}

	return &Connector{
		Conn:          conn,
		relevanceTime: cnf.RelevanceTime,
	}, nil
}

// Close - close connection with database
func (c *Connector) Close() {
	c.Conn.Close()
}
