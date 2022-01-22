package database

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// IsUserBanned returns true if the given user ID is banned in the bot (G-Line).
func (db *Database) IsUserBanned(userid int64) (bool, error) {
	err := db.conn.HGet(context.TODO(), "banlist", strconv.FormatInt(userid, 10)).Err()
	if err == redis.Nil {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, err
}

// SetUserBanned marks the given user ID as banned in the bot (G-Line).
func (db *Database) SetUserBanned(userid int64) error {
	return db.conn.HSet(context.TODO(), "banlist", strconv.FormatInt(userid, 10), time.Now().String()).Err()
}

// RemoveUserBanned unmarks the user as banned in the bot (G-Line).
func (db *Database) RemoveUserBanned(userid int64) error {
	return db.conn.HDel(context.TODO(), "banlist", strconv.FormatInt(userid, 10), time.Now().String()).Err()
}
