package botdatabase

import (
	"context"
	"fmt"
	"strconv"
)

// AddGlobalAdmin adds the given user as bot admin.
func (db *_botDatabase) AddGlobalAdmin(userID int64) error {
	id := strconv.FormatInt(userID, 10)
	if err := db.redisconn.SAdd(context.TODO(), "global-admins", id).Err(); err != nil {
		return fmt.Errorf("on SADD: %w", err)
	}
	return nil
}
