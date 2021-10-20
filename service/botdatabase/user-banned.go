package botdatabase

import (
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

// IsUserBanned returns true if the given user ID is banned in the bot (G-Line).
//
// Time complexity: O(1).
func (db *_botDatabase) IsUserBanned(userid int64) (bool, error) {
	err := db.redisconn.HGet("banlist", strconv.FormatInt(userid, 10)).Err()
	if err == redis.Nil {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, err
}

// SetUserBanned marks the given user ID as banned in the bot (G-Line).
//
// Time complexity: O(1).
func (db *_botDatabase) SetUserBanned(userid int64) error {
	return db.redisconn.HSet("banlist", strconv.FormatInt(userid, 10), time.Now().String()).Err()
}

// RemoveUserBanned unmarks the user as banned in the bot (G-Line).
//
// Time complexity: O(1).
func (db *_botDatabase) RemoveUserBanned(userid int64) error {
	return db.redisconn.HDel("banlist", strconv.FormatInt(userid, 10), time.Now().String()).Err()
}
