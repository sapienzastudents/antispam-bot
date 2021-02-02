package botdatabase

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

type BOTDatabase interface {
	IsGlobalAdmin(user *tb.User) bool

	GetChatSetting(b *tb.Bot, chat *tb.Chat) (ChatSettings, error)
	SetChatSettings(*tb.Chat, ChatSettings) error

	ListMyChatrooms() ([]*tb.Chat, error)
	ChatroomsCount() (int64, error)

	UpdateMyChatroomList(c *tb.Chat) error
	LeftChatroom(c *tb.Chat) error

	// I know that it's wrong to have this whole function here. It's also wrong
	// to have bot<->database and prometheus<->database inter-dependencies.
	// You're free to submit a patch for this, I'm too lazy to fix it now
	DoCacheUpdate(b *tb.Bot, g *prometheus.GaugeVec) error
	DoCacheUpdateForChat(b *tb.Bot, chat *tb.Chat) error

	GetChatTree(b *tb.Bot) (ChatCategoryTree, error)

	GetInviteLink(chatID int64) (string, error)
	SetInviteLink(chatID int64, inviteLink string) error

	GetUUIDFromChat(int64) (uuid.UUID, error)
	GetChatIDFromUUID(uuid.UUID) (int64, error)
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
