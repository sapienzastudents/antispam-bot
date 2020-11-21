package botdatabase

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"strconv"
	"strings"
)

type BOTDatabase interface {
	IsGlobalAdmin(user *tb.User) bool

	GetChatSetting(b *tb.Bot, chat *tb.Chat) (ChatSettings, error)
	SetChatSettings(*tb.Chat, ChatSettings) error

	ListMyChatrooms() ([]*tb.Chat, error)

	UpdateMyChatroomList(c *tb.Chat) error
	LeftChatroom(c *tb.Chat) error

	DoCacheUpdate(b *tb.Bot) error
	DoCacheUpdateForChat(b *tb.Bot, chat *tb.Chat) error

	GetChatCategory(c *tb.Chat) (string, error)
	SetChatCategory(c *tb.Chat, cat string) error
}

func New(logger *logrus.Entry) (BOTDatabase, error) {
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
		logger:    logger,
	}, nil
}

type _botDatabase struct {
	redisconn *redis.Client
	logger    *logrus.Entry
}

func (db *_botDatabase) IsGlobalAdmin(user *tb.User) bool {
	admins, err := db.redisconn.HGet("global", "admins").Result()
	if err != nil {
		db.logger.WithError(err).Error("Cannot get global admin list")
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
