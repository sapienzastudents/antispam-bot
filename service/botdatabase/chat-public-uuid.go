package botdatabase

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var ErrChatUUIDNotFound = errors.New("chat uuid not found")

func (db *_botDatabase) GetUUIDFromChat(chatID int64) (uuid.UUID, error) {
	chatUUIDString, err := db.redisconn.HGet("public-links", fmt.Sprint(chatID)).Result()
	if err == redis.Nil {
		// Not found
		chatUUID := uuid.New()
		err = db.redisconn.HSet("public-links", fmt.Sprint(chatID), chatUUID.String()).Err()
		return chatUUID, err
	} else if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(chatUUIDString)
}

func (db *_botDatabase) GetChatIDFromUUID(lookupUUID uuid.UUID) (int64, error) {
	var cursor uint64 = 0
	var err error
	var keys []string
	for {
		keys, cursor, err = db.redisconn.HScan("public-links", cursor, "", -1).Result()
		if err == redis.Nil {
			return 0, ErrChatUUIDNotFound
		}
		if err != nil {
			return 0, errors.Wrap(err, "error scanning uuids in redis")
		}

		for i := 0; i < len(keys); i += 2 {
			chatID, err := strconv.ParseInt(keys[i], 10, 64)
			if err != nil {
				continue
			}
			chatUUID, err := uuid.Parse(keys[i+1])
			if err != nil {
				continue
			}
			if chatUUID == lookupUUID {
				return chatID, nil
			}
		}

		// SCAN cycle end
		if cursor == 0 {
			break
		}
	}
	return 0, ErrChatUUIDNotFound
}
