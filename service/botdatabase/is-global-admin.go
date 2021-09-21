package botdatabase

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"strconv"
	"strings"
)

func (db *_botDatabase) IsGlobalAdmin(user *tb.User) bool {
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
			if ID == int64(user.ID) {
				return true
			}
		}
	}
	return false
}
