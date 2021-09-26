package botdatabase

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Database interface {
	// IsGlobalAdmin checks if the user ID is a bot admin
	IsGlobalAdmin(userID int) bool

	// GetChatSettings returns the chat settings of the bot for the given chat
	GetChatSettings(chatID int64) (ChatSettings, error)

	// SetChatSettings save the chat settings of the bot for the given chat
	SetChatSettings(chatID int64, settings ChatSettings) error

	// ListMyChatrooms returns the list of chatrooms where the bot is
	ListMyChatrooms() ([]*tb.Chat, error)

	// ChatroomsCount returns the count of chatrooms where the bot is
	ChatroomsCount() (int64, error)

	// AddOrUpdateChat adds or update the chat info into the DB. As Telegram doesn't offer a way to track in which
	// chatrooms the bot is, we need to store it in Redis
	AddOrUpdateChat(c *tb.Chat) error

	// DeleteChat remove all chatroom info
	DeleteChat(int64) error

	// GetChatTree returns the chat tree (categories)
	GetChatTree() (ChatCategoryTree, error)

	// GetInviteLink returns the cached invite link
	GetInviteLink(chatID int64) (string, error)

	// SetInviteLink save the invite link
	SetInviteLink(chatID int64, inviteLink string) error

	// GetUUIDFromChat returns the UUID for the given chat ID. The UUID can be used e.g. in web links
	GetUUIDFromChat(int64) (uuid.UUID, error)

	// GetChatIDFromUUID returns the chat ID for the given UUID
	GetChatIDFromUUID(uuid.UUID) (int64, error)

	// IsUserBanned checks if the user is banned in the bot (G-Line)
	IsUserBanned(int64) (bool, error)

	// SetUserBanned mark the user as banned in the bot (G-Line)
	SetUserBanned(int64) error

	// RemoveUserBanned unmark the user as banned in the bot (G-Line)
	RemoveUserBanned(int64) error
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
