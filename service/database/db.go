package database

import (
	"errors"

	"github.com/go-redis/redis/v8"
)

// Database represents an abstraction over the underlying database connection,
// it implements methods related to the bot's business logic.
type Database struct {
	conn *redis.Client
}

// New returns a new Database that uses the given redis client.
func New(client *redis.Client) (*Database, error) {
	if client == nil {
		return nil, errors.New("no redis connection specified")
	}
	return &Database{conn: client}, nil
}
