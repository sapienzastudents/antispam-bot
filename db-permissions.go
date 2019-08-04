package main

import (
	"fmt"
	"github.com/go-redis/redis"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (db *_botDatabase) IsBotEnabled(c *tb.Chat) (bool, error) {
	enabled, err := db.redisconn.HGet("settings:enabled", fmt.Sprintf("%d", c.ID)).Int()
	if err == redis.Nil {
		enabled = 1
		err = nil
	}
	return enabled != 0, err
}

func (db *_botDatabase) EnableBot(c *tb.Chat) error {
	return db.redisconn.HSet("settings:enabled", fmt.Sprintf("%d", c.ID), 1).Err()
}

func (db *_botDatabase) DisableBot(c *tb.Chat) error {
	return db.redisconn.HSet("settings:enabled", fmt.Sprintf("%d", c.ID), 1).Err()
}
