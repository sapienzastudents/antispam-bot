package botdatabase

import (
	"context"
	"strconv"
	"strings"
)

// IsGlobalAdmin checks if the given user ID is a bot admin.
//
// Time complexity: O(n) where "n" is the length of the global admin list.
func (db *_botDatabase) IsGlobalAdmin(userID int64) (bool, error) {
	globalAdminListExists, err := db.redisconn.HExists(context.TODO(), "global", "admins").Result()
	if err != nil {
		return false, err
	}

	if globalAdminListExists {
		// If the key doesn't exist, the function always returns false.
		admins, err := db.redisconn.HGet(context.TODO(), "global", "admins").Result()
		if err != nil {
			return false, err
		}

		// Check every item in global -> admins HSET against the user ID
		for _, sID := range strings.Split(admins, ",") {
			ID, err := strconv.ParseInt(sID, 10, 64)
			if err != nil {
				continue
			}
			if ID == userID {
				return true, nil
			}
		}
	}
	return false, nil
}
