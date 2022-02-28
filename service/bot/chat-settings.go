package bot

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/database"

	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v3"
)

// getChatSettings retrieves the chat settings from the database for the given
// chat.
//
// If there are not settings previously created, it creates a new set of
// settings based on some default values and returns it.
func (bot *telegramBot) getChatSettings(chat *tb.Chat) (chatSettings, error) {
	settings, err := bot.db.GetChatSettings(chat.ID)
	if err == database.ErrChatNotFound {
		err = nil
		// Chat settings not found, load default values
		settings = database.ChatSettings{
			BotEnabled:    true,
			OnJoinDelete:  false,
			OnLeaveDelete: false,
			OnJoinChinese: database.BotAction{
				Action:   database.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnJoinArabic: database.BotAction{
				Action:   database.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnMessageChinese: database.BotAction{
				Action:   database.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnMessageArabic: database.BotAction{
				Action:   database.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnBlacklistCAS: database.BotAction{
				Action:   database.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			ChatAdmins: database.ChatAdminList{},
		}

		// Private chats doesn't have admins, Telegram will reply with an error.
		if chat.Type != tb.ChatPrivate {
			chatAdmins, err := bot.telebot.AdminsOf(chat)
			if err != nil {
				return chatSettings{}, fmt.Errorf("failed to get admin list for chat: %w", err)
			}
			settings.ChatAdmins.SetFromChat(chatAdmins)
		}

		err = bot.db.SetChatSettings(chat.ID, settings)
		if err != nil {
			return chatSettings{}, fmt.Errorf("failed to save chat settings for new chat: %w", err)
		}
	}

	return chatSettings{
		ChatSettings: settings,
		logger:       bot.logger,
		b:            bot.telebot,
	}, err
}

type chatSettings struct {
	database.ChatSettings

	globalLog int64
	b         *tb.Bot
	logger    logrus.FieldLogger
}

// Log will log the action in the dedicated log channel for the group
func (s *chatSettings) Log(action string, by *tb.User, on *tb.User, description string) {
	if s.LogChannel > 0 || s.globalLog > 0 {
		strbuf := strings.Builder{}
		strbuf.WriteString("**Action**: ")
		strbuf.WriteString(action)
		strbuf.WriteString("\n")

		if by != nil {
			strbuf.WriteString("**By**: ")
			strbuf.WriteString(by.FirstName + " " + by.LastName)
			strbuf.WriteString("(")
			strbuf.WriteString("@" + by.Username + " " + strconv.FormatInt(int64(by.ID), 10))
			strbuf.WriteString(")")
			strbuf.WriteString("\n")
		}
		if on != nil {
			strbuf.WriteString("**On**: ")
			strbuf.WriteString(on.FirstName + " " + on.LastName)
			strbuf.WriteString("(")
			strbuf.WriteString("@" + on.Username + " " + strconv.FormatInt(int64(on.ID), 10))
			strbuf.WriteString(")")
			strbuf.WriteString("\n")
		}
		strbuf.WriteString("**Details**: \n")
		strbuf.WriteString(description)

		if s.LogChannel > 0 {
			_, err := s.b.Send(&tb.Chat{ID: s.LogChannel}, strbuf.String(), tb.ModeMarkdown)
			if err != nil {
				s.logger.WithError(err).WithField("chatid", s.LogChannel).Error("failed to send message to log channel")
			}
		}

		if s.globalLog > 0 {
			_, err := s.b.Send(&tb.Chat{ID: s.globalLog}, strbuf.String(), tb.ModeMarkdown)
			if err != nil {
				s.logger.WithError(err).WithField("chatid", s.globalLog).Error("failed to send message to log channel")
			}
		}
	}
}

// LogForward will forward a message to the log channel
func (s *chatSettings) LogForward(message *tb.Message) {
	if s.LogChannel > 0 {
		_, err := s.b.Forward(&tb.Chat{ID: s.LogChannel}, message)
		if err != nil {
			s.logger.WithError(err).WithField("chatid", s.LogChannel).Error("failed to forward message to log channel")
		}
	}
	if s.globalLog > 0 {
		_, err := s.b.Forward(&tb.Chat{ID: s.globalLog}, message)
		if err != nil {
			s.logger.WithError(err).WithField("chatid", s.globalLog).Error("failed to forward message to log channel")
		}
	}
}
