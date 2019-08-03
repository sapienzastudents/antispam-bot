package main

import (
	"encoding/json"
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
)

// As Telegram doesn't offer a way to track in which chatrooms the bot is, we need to store it in Redis every time a
// new message is seen
func updateMyChatList(c *tb.Chat) {
	jsonbin, err := json.Marshal(c)
	if err != nil {
		logger.Critical("Cannot JSONize chat %s (%d)", c.Title, c.ID)
		return
	}
	redisDb.Set(fmt.Sprintf("chat.%d", c.ID), string(jsonbin), 0)
}
