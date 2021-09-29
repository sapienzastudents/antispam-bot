package botdatabase

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

// IsUserBanned checks if the user is banned in the bot (G-Line)
func (db *_botDatabase) IsUserBanned(userid int64) (bool, error) {
	err := db.redisconn.HGet("banlist", strconv.FormatInt(userid, 10)).Err()
	if err == redis.Nil {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, err
}

// SetUserBanned mark the user as banned in the bot (G-Line)
func (db *_botDatabase) SetUserBanned(userid int64) error {
	return db.redisconn.HSet("banlist", strconv.FormatInt(userid, 10), time.Now().String()).Err()
}

// RemoveUserBanned unmark the user as banned in the bot (G-Line)
func (db *_botDatabase) RemoveUserBanned(userid int64) error {
	return db.redisconn.HDel("banlist", strconv.FormatInt(userid, 10), time.Now().String()).Err()
}
