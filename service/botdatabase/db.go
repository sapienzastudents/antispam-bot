package botdatabase

import (
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	tb "gopkg.in/tucnak/telebot.v3"
)

type Database interface {
	// IsGlobalAdmin returns true if the given user ID is a bot admin.
	IsGlobalAdmin(userID int64) (bool, error)

	// AddGlobalAdmin adds the given user as bot admin.
	AddGlobalAdmin(userID int64) error

	// GetChatSettings returns the chat settings of the bot for the given chat
	// ID.
	GetChatSettings(chatID int64) (ChatSettings, error)

	// SetChatSettings save the chat settings of the bot for the given chat ID.
	SetChatSettings(chatID int64, settings ChatSettings) error

	// ListMyChatrooms returns the list of chatrooms where the bot is.
	ListMyChats() ([]*tb.Chat, error)

	// ChatroomsCount returns the count of chatrooms where the bot is.
	ChatroomsCount() (int64, error)

	// AddOrUpdateChat adds or updates the chat info into the DB.
	//
	// As Telegram doesn't offer a way to track in which chatrooms the bot is,
	// we need to store it in Redis.
	AddChat(c *tb.Chat) error

	// DeleteChat removes all chatroom info.
	DeleteChat(int64) error

	// GetChatTree returns the chat tree (categories).
	GetChatTree() (ChatCategoryTree, error)

	// GetInviteLink returns the cached invite link.
	GetInviteLink(chatID int64) (string, error)

	// SetInviteLink saves the given invite link.
	SetInviteLink(chatID int64, inviteLink string) error

	// GetUUIDFromChat returns the UUID for the given chat ID.
	//
	// The UUID can be used e.g. in web links.
	GetUUIDFromChat(int64) (uuid.UUID, error)

	// GetChatIDFromUUID returns the chat ID for the given UUID.
	GetChatIDFromUUID(uuid.UUID) (int64, error)

	// IsUserBanned returns true if the given user ID is banned in the bot
	// (G-Line).
	IsUserBanned(userID int64) (bool, error)

	// SetUserBanned marks the given user ID as banned in the bot (G-Line).
	SetUserBanned(userID int64) error

	// RemoveUserBanned unmarks the given user ID as banned in the bot (G-Line).
	RemoveUserBanned(userID int64) error
}

// New returns a new instance of the bot database conforming to Database
// interface.
func New(redisclient *redis.Client) (Database, error) {
	if redisclient == nil {
		return nil, errors.New("no redis connection specified")
	}

	return &_botDatabase{
		redisconn: redisclient,
	}, nil
}

// _botDatabase is the concrete type that implements Database interface.
type _botDatabase struct {
	redisconn *redis.Client
}
