package storage

import (
	"context"
	"regexp"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/tern/migrate"

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

// Migrate - Applies the changes recorded in the migration files to the
//  database.
func (c *Connector) Migrate(ctx context.Context, path string) error {
	conn, err := c.Conn.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "versions")
	if err != nil {
		return err
	}
	err = migrator.LoadMigrations(path)
	if err != nil {
		return err
	}
	return migrator.Migrate(ctx)
}

// Close - close connection with database
func (c *Connector) Close() {
	c.Conn.Close()
}
