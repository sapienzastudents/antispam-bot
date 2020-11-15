package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (db *_botDatabase) ListMyChatrooms() ([]*tb.Chat, error) {
	chatrooms := []*tb.Chat{}

	var cursor uint64 = 0
	var err error
	var keys []string
	for {
		keys, cursor, err = db.redisconn.HScan("chatrooms", cursor, "", -1).Result()
		if err == redis.Nil {
			return chatrooms, nil
		}
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(keys); i += 2 {
			room := tb.Chat{}
			err = json.Unmarshal([]byte(keys[i+1]), &room)
			if err != nil {
				// TODO: skip?
				return nil, err
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

// As Telegram doesn't offer a way to track in which chatrooms the bot is, we need to store it in Redis every time a
// new message is seen
func (db *_botDatabase) UpdateMyChatroomList(c *tb.Chat) error {
	jsonbin, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return db.redisconn.HSet("chatrooms", fmt.Sprintf("%d", c.ID), string(jsonbin)).Err()
}

func (db *_botDatabase) LeftChatroom(c *tb.Chat) error {
	return db.redisconn.HDel("chatrooms", fmt.Sprintf("%d", c.ID)).Err()
}
