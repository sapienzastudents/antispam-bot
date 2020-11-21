package botdatabase

import (
	"fmt"
	"github.com/go-redis/redis"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (db *_botDatabase) GetChatCategory(c *tb.Chat) (string, error) {
	category, err := db.redisconn.HGet("chat-categories", fmt.Sprint(c.ID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return category, err
}

func (db *_botDatabase) SetChatCategory(c *tb.Chat, cat string) error {
	return db.redisconn.HSet("chat-categories", fmt.Sprint(c.ID), cat).Err()
}
