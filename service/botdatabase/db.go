package botdatabase

import (
	"errors"
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

	IsUserBanned(int64) (bool, error)
	SetUserBanned(int64) error
	RemoveUserBanned(int64) error

	DeleteChat(int64) error
}

type Options struct {
	Logger logrus.FieldLogger
	Redis  *redis.Client
}

func New(opts Options) (BOTDatabase, error) {
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
