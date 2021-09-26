package botdatabase

import (
	"strconv"
	"strings"
)

func (db *_botDatabase) IsGlobalAdmin(userID int) bool {
	globalAdminListExists, err := db.redisconn.HExists("global", "admins").Result()
	if err != nil {
		db.logger.WithError(err).Error("Cannot check if global admin list exists")
		return false
	} else if globalAdminListExists {
		admins, err := db.redisconn.HGet("global", "admins").Result()
		if err != nil {
			db.logger.WithError(err).Error("Cannot get global admin list")
			return false
		}

		for _, sID := range strings.Split(admins, ",") {
			ID, err := strconv.ParseInt(sID, 10, 64)
			if err != nil {
				continue
			}
			if ID == int64(userID) {
				return true
			}
		}
	}
	return false
}
