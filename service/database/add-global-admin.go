package database

import (
	"context"
	"fmt"
	"strconv"
)

// AddGlobalAdmin adds the given user as bot admin.
func (db *Database) AddGlobalAdmin(userID int64) error {
	id := strconv.FormatInt(userID, 10)
	if err := db.conn.SAdd(context.TODO(), "global-admins", id).Err(); err != nil {
		return fmt.Errorf("on SADD: %w", err)
	}
	return nil
}
