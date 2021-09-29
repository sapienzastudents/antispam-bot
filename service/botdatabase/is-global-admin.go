package botdatabase

import (
	"strconv"
	"strings"
)

// IsGlobalAdmin checks if the user ID is a bot admin in redis hash key global -> admins. If the key doesn't exist,
// the function always returns false
func (db *_botDatabase) IsGlobalAdmin(userID int) (bool, error) {
	globalAdminListExists, err := db.redisconn.HExists("global", "admins").Result()
	if err != nil {
		return false, err
	}

	if globalAdminListExists {
		admins, err := db.redisconn.HGet("global", "admins").Result()
		if err != nil {
			return false, err
		}

		// Check every item in global -> admins HSET against the user ID
		for _, sID := range strings.Split(admins, ",") {
			ID, err := strconv.ParseInt(sID, 10, 64)
			if err != nil {
				continue
			}
			// Currently, we need to cast the user ID here to int64 as telebot API uses int and not int64 for user IDs
			// (this is a technical debt in the library). Once we update to the new version we should be able to either
			// use int64 direcly (as hopefully the library will be fixed) or we will cast the userID outside this
			// function
			if ID == int64(userID) {
				return true, nil
			}
		}
	}
	return false, nil
}
