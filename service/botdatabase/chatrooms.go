package botdatabase

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
)

// AddOrUpdateChat adds or update the chat info into the DB. As Telegram doesn't offer a way to track in which
// chatrooms the bot is, we need to store it in Redis
func (db *_botDatabase) AddOrUpdateChat(c *tb.Chat) error {
	jsonbin, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return db.redisconn.HSet(context.TODO(), "chatrooms", strconv.FormatInt(c.ID, 10), string(jsonbin)).Err()
}

// DeleteChat remove all chatroom info by removing the named field in sets: "public-links", "settings" and "chatrooms"
func (db *_botDatabase) DeleteChat(chatID int64) error {
	err := db.redisconn.HDel(context.TODO(), "public-links", strconv.FormatInt(chatID, 10)).Err()
	if err != nil {
		return err
	}
	err = db.redisconn.HDel(context.TODO(), "settings", strconv.FormatInt(chatID, 10)).Err()
	if err != nil {
		return err
	}
	return db.redisconn.HDel(context.TODO(), "chatrooms", strconv.FormatInt(chatID, 10)).Err()
}

// ChatroomsCount returns the count of chatrooms where the bot is
func (db *_botDatabase) ChatroomsCount() (int64, error) {
	ret, err := db.redisconn.HLen(context.TODO(), "chatrooms").Result()
	if err == redis.Nil {
		return 0, nil
	}
	return ret, err
}

// ListMyChatrooms returns the list of chatrooms where the bot is by de-serializing the tb.Chat for each chatroom
func (db *_botDatabase) ListMyChatrooms() ([]*tb.Chat, error) {
	var chatrooms []*tb.Chat

	var cursor uint64 = 0
	var err error
	var keys []string
	for {
		keys, cursor, err = db.redisconn.HScan(context.TODO(), "chatrooms", cursor, "", -1).Result()
		if err == redis.Nil {
			return chatrooms, nil
		} else if err != nil {
			return nil, errors.Wrap(err, "error scanning chatrooms in redis")
		}

		for i := 0; i < len(keys); i += 2 {
			room := tb.Chat{}
			err = json.Unmarshal([]byte(keys[i+1]), &room)
			if err != nil {
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
