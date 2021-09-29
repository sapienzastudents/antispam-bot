package tbot

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"strconv"
	"strings"
)

// getChatSettings retrieves the chat settings from the database. If not settings were previously created, this command
// creates a new set of settings based on some default values
func (bot *telegramBot) getChatSettings(chat *tb.Chat) (chatSettings, error) {
	settings, err := bot.db.GetChatSettings(chat.ID)
	if err == botdatabase.ErrChatNotFound {
		err = nil
		// Chat settings not found, load default values
		settings = botdatabase.ChatSettings{
			BotEnabled:    true,
			OnJoinDelete:  false,
			OnLeaveDelete: false,
			OnJoinChinese: botdatabase.BotAction{
				Action:   botdatabase.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnJoinArabic: botdatabase.BotAction{
				Action:   botdatabase.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnMessageChinese: botdatabase.BotAction{
				Action:   botdatabase.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnMessageArabic: botdatabase.BotAction{
				Action:   botdatabase.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			OnBlacklistCAS: botdatabase.BotAction{
				Action:   botdatabase.ActionNone,
				Duration: 0,
				Delay:    0,
			},
			ChatAdmins: botdatabase.ChatAdminList{},
		}

		chatAdmins, err := bot.telebot.AdminsOf(chat)
		if err != nil {
			return chatSettings{}, errors.Wrap(err, "can't get admin list for chat")
		}
		settings.ChatAdmins.SetFromChat(chatAdmins)

		err = bot.db.SetChatSettings(chat.ID, settings)
		if err != nil {
			return chatSettings{}, errors.Wrap(err, "can't save chat settings for new chat")
		}
	}

	return chatSettings{
		ChatSettings: settings,
		logger:       bot.logger,
		b:            bot.telebot,
	}, err
}

type chatSettings struct {
	botdatabase.ChatSettings

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
				s.logger.WithError(err).WithField("chatid", s.LogChannel).Error("can't send message to log channel")
			}
		}

		if s.globalLog > 0 {
			_, err := s.b.Send(&tb.Chat{ID: s.globalLog}, strbuf.String(), tb.ModeMarkdown)
			if err != nil {
				s.logger.WithError(err).WithField("chatid", s.globalLog).Error("can't send message to log channel")
			}
		}
	}
}

// LogForward will forward a message to the log channel
func (s *chatSettings) LogForward(message *tb.Message) {
	if s.LogChannel > 0 {
		_, err := s.b.Forward(&tb.Chat{ID: s.LogChannel}, message)
		if err != nil {
			s.logger.WithError(err).WithField("chatid", s.LogChannel).Error("can't forward message to log channel")
		}
	}
	if s.globalLog > 0 {
		_, err := s.b.Forward(&tb.Chat{ID: s.globalLog}, message)
		if err != nil {
			s.logger.WithError(err).WithField("chatid", s.globalLog).Error("can't forward message to log channel")
		}
	}
}
