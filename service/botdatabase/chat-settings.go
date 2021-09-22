package botdatabase

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	ActionNone      = 0
	ActionMute      = 1
	ActionKick      = 2
	ActionBan       = 3
	ActionDeleteMsg = 4
)

type ChatAdminList []int64

func (list *ChatAdminList) IsAdmin(user *tb.User) bool {
	if list == nil {
		return false
	}
	for _, v := range *list {
		if v == int64(user.ID) {
			return true
		}
	}
	return false
}

func (list *ChatAdminList) SetFromChat(admins []tb.ChatMember) {
	*list = ChatAdminList{}
	for _, u := range admins {
		*list = append(*list, int64(u.User.ID))
	}
}

type BotAction struct {
	// Action in effect
	Action int `json:"action"`

	// Action effect, in seconds (if applicable)
	// Zero means forever
	Duration uint `json:"duration"`

	// Action delay, zero means "immediately"
	Delay uint `json:"delay"`
}

type ChatSettings struct {
	// Whether the bot is enabled for this chat
	BotEnabled bool `json:"bot_enabled"`

	Hidden bool `json:"hidden"`

	OnJoinDelete  bool `json:"on_join_delete"`
	OnLeaveDelete bool `json:"on_leave_delete"`

	OnJoinChinese BotAction `json:"on_join_chinese"`
	OnJoinArabic  BotAction `json:"on_join_arabic"`

	// Action on specific messages patterns
	OnMessageChinese BotAction `json:"on_message_chinese"`
	OnMessageArabic  BotAction `json:"on_message_arabic"`
	//OnMessageSpam    BotAction `json:"on_message_spam"`

	OnBlacklistCAS BotAction `json:"on_blacklist_cas"`

	// Chat admins
	ChatAdmins ChatAdminList `json:"chat_admins"`

	MainCategory string `json:"main_category"`
	SubCategory  string `json:"sub_category"`

	LogChannel int64 `json:"log_channel"`
}

func (db *_botDatabase) GetChatSetting(b *tb.Bot, chat *tb.Chat) (ChatSettings, error) {
	settings := ChatSettings{}
	jsonb, err := db.redisconn.HGet("settings", fmt.Sprintf("%d", chat.ID)).Result()
	if err == redis.Nil {
		// Settings not found, load default values
		settings = ChatSettings{
			BotEnabled:    true,
			OnJoinDelete:  false,
			OnLeaveDelete: false,
			OnJoinChinese: BotAction{
				Action:   ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnJoinArabic: BotAction{
				Action:   ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnMessageChinese: BotAction{
				Action:   ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnMessageArabic: BotAction{
				Action:   ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnBlacklistCAS: BotAction{
				Action:   ActionNone,
				Duration: 0,
				Delay:    0,
			},
			ChatAdmins: ChatAdminList{},
		}

		chatAdmins, err := b.AdminsOf(chat)
		if err != nil {
			return ChatSettings{}, errors.Wrap(err, "can't get admin list for chat")
		}
		settings.ChatAdmins.SetFromChat(chatAdmins)

		err = db.SetChatSettings(chat.ID, settings)
		if err != nil {
			return ChatSettings{}, errors.Wrap(err, "can't save chat settings for new chat")
		}
	} else if err != nil {
		return ChatSettings{}, err
	} else {
		err = json.Unmarshal([]byte(jsonb), &settings)
		if err != nil {
			return ChatSettings{}, errors.Wrap(err, "error decoding chat settings from JSON")
		}
	}

	return settings, err
}

func (db *_botDatabase) SetChatSettings(chatID int64, settings ChatSettings) error {
	jsonb, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return db.redisconn.HSet("settings", fmt.Sprintf("%d", chatID), jsonb).Err()
}
