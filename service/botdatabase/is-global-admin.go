package botdatabase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// IsGlobalAdmin returns true if the given userID is a bot admin.
func (db *_botDatabase) IsGlobalAdmin(userID int64) (bool, error) {
	// Migrate old database.
	oldGlobalAdmins, err := db.redisconn.HExists(context.TODO(), "global", "admins").Result()
	if err != nil {
		return false, err
	}
	if oldGlobalAdmins {
		admins, err := db.redisconn.HGet(context.TODO(), "global", "admins").Result()
		if err != nil {
			return false, fmt.Errorf("on HGET from old database: %w", err)
		}

		// Migrate every item.
		for _, sID := range strings.Split(admins, ",") {
			ID, err := strconv.ParseInt(sID, 10, 64)
			if err != nil {
				continue
			}
			if err := db.AddGlobalAdmin(ID); err != nil {
				return false, fmt.Errorf("during migration from old database: %w", err)
			}
		}

		// Delete old database, otherwhise the migration is done at every call
		// of IsGlobalAdmin.
		if err := db.redisconn.HDel(context.TODO(), "global", "admins").Err(); err != nil {
			return false, fmt.Errorf("on HDEL on old database: %w", err)
		}
	}

	id := strconv.FormatInt(userID, 10)
	is, err := db.redisconn.SIsMember(context.TODO(), "global-admins", id).Result()
	if err != nil {
		return false, fmt.Errorf("on SISMEMBER: %w", err)
	}
	return is, nil
}
