package main

import (
	"encoding/json"
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (db *_botDatabase) ListMyChatrooms() ([]*tb.Chat, error) {
	chatrooms := []*tb.Chat{}

	var cursor uint64 = 0
	var err error
	var keys []string
	for {
		keys, cursor, err = db.redisconn.HScan("chatrooms", cursor, "*", 50).Result()
		if err != nil {
			return nil, err
		}

		for _, k := range keys {
			roombytes, err := db.redisconn.HGet("chatrooms", k).Result()
			if err != nil {
				// TODO: skip?
				return nil, err
			}

			room := tb.Chat{}
			err = json.Unmarshal([]byte(roombytes), &room)
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
