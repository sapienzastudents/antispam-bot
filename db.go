package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/patrickmn/go-cache"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"strconv"
	"strings"
	"time"
)

type BOTDatabase interface {
	IsGlobalAdmin(user *tb.User) bool

	GetChatSetting(*tb.Chat) (ChatSettings, error)
	SetChatSettings(*tb.Chat, ChatSettings) error

	ListMyChatrooms() ([]*tb.Chat, error)

	UpdateMyChatroomList(c *tb.Chat) error
	LeftChatroom(c *tb.Chat) error

	DoCacheUpdate() error
	DoCacheUpdateForChat(chat *tb.Chat) error
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

	return &_botDatabase{
		redisconn: redisDb,
		cache:     cache.New(7*24*time.Hour, 7*24*time.Hour),
	}, nil
}

type _botDatabase struct {
	redisconn *redis.Client
	cache     *cache.Cache
}

func (db *_botDatabase) IsGlobalAdmin(user *tb.User) bool {
	admins, err := db.redisconn.HGet("global", "admins").Result()
	if err != nil {
		logger.Critical("Cannot get global admin list:", err)
		return false
	}

	for _, sID := range strings.Split(admins, ",") {
		ID, err := strconv.ParseInt(sID, 10, 64)
		if err != nil {
			continue
		}
		if ID == int64(user.ID) {
			return true
		}
	}
	return false
}
