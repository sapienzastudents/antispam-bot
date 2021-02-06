package botdatabase

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func (db *_botDatabase) IsUserBanned(userid int64) (bool, error) {
	err := db.redisconn.HGet("banlist", fmt.Sprint(userid)).Err()
	if err == redis.Nil {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, err
}

func (db *_botDatabase) SetUserBanned(userid int64) error {
	return db.redisconn.HSet("banlist", fmt.Sprint(userid), time.Now().String()).Err()
}
