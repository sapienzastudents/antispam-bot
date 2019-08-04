package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"strconv"
	"strings"
)

type BOTDatabase interface {
	IsGlobalAdmin(user *tb.User) (bool, error)

	IsBotEnabled(c *tb.Chat) (bool, error)
	EnableBot(c *tb.Chat) error
	DisableBot(c *tb.Chat) error

	ListMyChatrooms() ([]*tb.Chat, error)

	UpdateMyChatroomList(c *tb.Chat) error
	LeftChatroom(c *tb.Chat) error
}

func NewBotDatabase() (BOTDatabase, error) {
	redisOptions, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return nil, errors.New(fmt.Sprint("Unable to parse REDIS_URL variable:", err))
	}
	redisDb := redis.NewClient(redisOptions)
	err = redisDb.Ping().Err()
	if err != nil {
		return nil, errors.New(fmt.Sprint("Unable to connect to Redis server:", err))
	}

	return &_botDatabase{redisconn: redisDb}, nil
}

type _botDatabase struct {
	redisconn *redis.Client
}

func (db *_botDatabase) IsGlobalAdmin(user *tb.User) (bool, error) {
	admins, err := db.redisconn.HGet("global", "admins").Result()
	if err != nil {
		return false, err
	}

	for _, sID := range strings.Split(admins, ",") {
		ID, err := strconv.ParseInt(sID, 10, 64)
		if err != nil {
			continue
		}
		if ID == int64(user.ID) {
			return true, nil
		}
	}
	return false, nil
}
