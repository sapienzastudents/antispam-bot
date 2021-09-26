package botdatabase

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Database interface {
	IsGlobalAdmin(userID int) bool

	GetChatSettings(chatID int64) (ChatSettings, error)
	SetChatSettings(chatID int64, settings ChatSettings) error

	ListMyChatrooms() ([]*tb.Chat, error)
	ChatroomsCount() (int64, error)

	UpdateMyChatroomList(c *tb.Chat) error
	LeftChatroom(chatID int64) error

	GetChatTree() (ChatCategoryTree, error)

	GetInviteLink(chatID int64) (string, error)
	SetInviteLink(chatID int64, inviteLink string) error

	GetUUIDFromChat(int64) (uuid.UUID, error)
	GetChatIDFromUUID(uuid.UUID) (int64, error)

	IsUserBanned(int64) (bool, error)
	SetUserBanned(int64) error
	RemoveUserBanned(int64) error

	DeleteChat(int64) error
}

type Options struct {
	Logger logrus.FieldLogger
	Redis  *redis.Client
}

func New(opts Options) (Database, error) {
	if opts.Logger == nil {
		return nil, errors.New("no logger specified")
	}
	if opts.Redis == nil {
		return nil, errors.New("no redis connection specified")
	}

	return &_botDatabase{
		redisconn: opts.Redis,
		logger:    opts.Logger,
	}, nil
}

type _botDatabase struct {
	redisconn *redis.Client
	logger    logrus.FieldLogger
}
