package tbot

import (
	"github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

// contextualChatSettingsFunc is the signature of the function that can be passed to refreshDBInfo as next handler in
// the chain
type contextualChatSettingsFunc func(*tb.Message, chatSettings)

// refreshDBInfo wrapper is refreshing the cache for chats in the database (due the fact that Telegram APIs does not
// support listing chats of bots, we need to keep track of all chats where we are)
func (bot *telegramBot) refreshDBInfo(actionHandler contextualChatSettingsFunc) func(m *tb.Message) {
	return func(m *tb.Message) {
		// Do not accept messages from channels
		if m.FromChannel() {
			return
		}

		if !m.Private() {
			// When the message is sent in a group, we need to:
			// 1. Update the chat info in the DB (or add the chat if it's new)
			// 2. Retrieve the chat settings
			// 3a. If the bot is not enabled (and the command is not from a global admin), ignore the message
			// 3b. Otherwise, we can call the next handler in the chain (i.e. actionHandler) and move on

			err := bot.db.AddOrUpdateChat(m.Chat)
			if err != nil {
				bot.logger.WithError(err).Error("Cannot update my chatroom list")
				return
			}

			settings, err := bot.getChatSettings(m.Chat)
			if err != nil {
				bot.logger.WithError(err).Error("Cannot get chat settings")
				return
			}

			isGlobalAdmin, err := bot.db.IsGlobalAdmin(m.Sender.ID)
			if err != nil {
				bot.logger.WithError(err).Error("can't check if the user is a global admin")
				return
			}

			if !settings.BotEnabled && !isGlobalAdmin {
				bot.logger.WithFields(logrus.Fields{
					"chatid":    m.Chat.ID,
					"chattitle": m.Chat.Title,
				}).Debugf("Bot not enabled for chat")
			} else {
				actionHandler(m, settings)
			}
		} else {
			// On private messages, no chat settings is available
			actionHandler(m, chatSettings{})
		}
	}
}
