package tbot

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

type refreshDBInfoFunc func(*tb.Message, botdatabase.ChatSettings)

// refreshDBInfo wrapper is refreshing the info for the chat in the database
// (due the fact that Telegram APIs does not support listing chats)
func (bot *telegramBot) refreshDBInfo(actionHandler refreshDBInfoFunc) func(m *tb.Message) {
	return func(m *tb.Message) {
		if !m.Private() {
			err := bot.db.UpdateMyChatroomList(m.Chat)
			if err != nil {
				bot.logger.WithError(err).Error("Cannot update my chatroom list")
				return
			}

			settings, err := bot.db.GetChatSetting(bot.telebot, m.Chat)
			if err != nil {
				bot.logger.WithError(err).Error("Cannot get chat settings")
			} else if !settings.BotEnabled && !bot.db.IsGlobalAdmin(m.Sender.ID) {
				bot.logger.WithFields(logrus.Fields{
					"chatid":    m.Chat.ID,
					"chattitle": m.Chat.Title,
				}).Debugf("Bot not enabled for chat")
			} else {
				actionHandler(m, settings)
			}
		} else {
			actionHandler(m, botdatabase.ChatSettings{})
		}
	}
}
