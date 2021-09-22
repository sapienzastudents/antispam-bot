package tbot

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func (bot *telegramBot) getChatSettings(chat *tb.Chat) (chatSettings, error) {
	settings, err := bot.db.GetChatSetting(bot.telebot, chat)
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
			strbuf.WriteString("@" + by.Username + " " + fmt.Sprint(by.ID))
			strbuf.WriteString(")")
			strbuf.WriteString("\n")
		}
		if on != nil {
			strbuf.WriteString("**On**: ")
			strbuf.WriteString(on.FirstName + " " + on.LastName)
			strbuf.WriteString("(")
			strbuf.WriteString("@" + on.Username + " " + fmt.Sprint(on.ID))
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
