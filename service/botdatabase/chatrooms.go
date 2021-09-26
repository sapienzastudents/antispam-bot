package botdatabase

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (db *_botDatabase) AddOrUpdateChat(c *tb.Chat) error {
	jsonbin, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return db.redisconn.HSet("chatrooms", fmt.Sprintf("%d", c.ID), string(jsonbin)).Err()
}

func (db *_botDatabase) DeleteChat(chatID int64) error {
	err := db.redisconn.HDel("public-links", fmt.Sprint(chatID)).Err()
	if err != nil {
		return err
	}
	err = db.redisconn.HDel("settings", fmt.Sprintf("%d", chatID)).Err()
	if err != nil {
		return err
	}
	return db.redisconn.HDel("chatrooms", fmt.Sprintf("%d", chatID)).Err()
}

func (db *_botDatabase) ChatroomsCount() (int64, error) {
	ret, err := db.redisconn.HLen("chatrooms").Result()
	if err == redis.Nil {
		return 0, nil
	}
	return ret, err
}

func (db *_botDatabase) ListMyChatrooms() ([]*tb.Chat, error) {
	var chatrooms []*tb.Chat

	var cursor uint64 = 0
	var err error
	var keys []string
	for {
		keys, cursor, err = db.redisconn.HScan("chatrooms", cursor, "", -1).Result()
		if err == redis.Nil {
			return chatrooms, nil
		}
		if err != nil {
			return nil, errors.Wrap(err, "error scanning chatrooms in redis")
		}

		for i := 0; i < len(keys); i += 2 {
			room := tb.Chat{}
			err = json.Unmarshal([]byte(keys[i+1]), &room)
			if err != nil {
				// TODO: skip?
				return nil, errors.Wrap(err, "error scanning chatroom "+keys[i+1])
			}

			chatrooms = append(chatrooms, &room)
		}

		// SCAN cycle end
		if cursor == 0 {
			break
		}
	}

	return chatrooms, nil
}
