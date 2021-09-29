package botdatabase

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Database interface {
	// IsGlobalAdmin checks if the user ID is a bot admin.
	//
	// Time complexity: O(n) where "n" is the length of the global admin list
	IsGlobalAdmin(userID int) (bool, error)

	// GetChatSettings returns the chat settings of the bot for the given chat.
	//
	// Time complexity: O(1)
	GetChatSettings(chatID int64) (ChatSettings, error)

	// SetChatSettings save the chat settings of the bot for the given chat.
	//
	// Time complexity: O(1)
	SetChatSettings(chatID int64, settings ChatSettings) error

	// ListMyChatrooms returns the list of chatrooms where the bot is.
	//
	// Time complexity: O(n) where "n" is the number of chat where the bot is
	ListMyChatrooms() ([]*tb.Chat, error)

	// ChatroomsCount returns the count of chatrooms where the bot is
	//
	// Time complexity: O(1)
	ChatroomsCount() (int64, error)

	// AddOrUpdateChat adds or update the chat info into the DB. As Telegram doesn't offer a way to track in which
	// chatrooms the bot is, we need to store it in Redis
	//
	// Time complexity: O(1)
	AddOrUpdateChat(c *tb.Chat) error

	// DeleteChat remove all chatroom info
	//
	// Time complexity: O(1)
	DeleteChat(int64) error

	// GetChatTree returns the chat tree (categories)
	//
	// Time complexity: O(n) where "n" is the number of chatroom where the bot is
	GetChatTree() (ChatCategoryTree, error)

	// GetInviteLink returns the cached invite link
	//
	// Time complexity: O(1)
	GetInviteLink(chatID int64) (string, error)

	// SetInviteLink save the invite link
	//
	// Time complexity: O(1)
	SetInviteLink(chatID int64, inviteLink string) error

	// GetUUIDFromChat returns the UUID for the given chat ID. The UUID can be used e.g. in web links
	//
	// Time complexity: O(1)
	GetUUIDFromChat(int64) (uuid.UUID, error)

	// GetChatIDFromUUID returns the chat ID for the given UUID
	//
	// Time complexity: O(n) where "n" is the number of chatrooms where the bot is
	GetChatIDFromUUID(uuid.UUID) (int64, error)

	// IsUserBanned checks if the user is banned in the bot (G-Line)
	//
	// Time complexity: O(1)
	IsUserBanned(int64) (bool, error)

	// SetUserBanned mark the user as banned in the bot (G-Line)
	//
	// Time complexity: O(1)
	SetUserBanned(int64) error

	// RemoveUserBanned unmark the user as banned in the bot (G-Line)
	//
	// Time complexity: O(1)
	RemoveUserBanned(int64) error
}

// New returns a new instance of the bot database conforming to Database interface
func New(redisclient *redis.Client) (Database, error) {
	if redisclient == nil {
		return nil, errors.New("no redis connection specified")
	}

	return &_botDatabase{
		redisconn: redisclient,
	}, nil
}

type _botDatabase struct {
	redisconn *redis.Client
}
