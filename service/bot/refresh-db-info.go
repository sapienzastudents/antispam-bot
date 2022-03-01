package bot

import (
	"github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

// contextualChatSettingsFunc is the signature of the function that can be
// passed to refreshDBInfo as next handler in the chain.
type contextualChatSettingsFunc func(tb.Context, chatSettings)

// refreshDBInfo returns an HandlerFunc suited to be passed on bot. It wraps the
// given handler so it can refresh the cache for chats in the database.
//
// Telegram APis does not support listing chats of bots, we need to keep track
// of all of chats where we are.
func (bot *telegramBot) refreshDBInfo(handler contextualChatSettingsFunc) tb.HandlerFunc {
	return func(ctx tb.Context) error {
		chat := ctx.Chat()
		if chat == nil {
			bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Chat, ignored")
			return nil
		}

		// Update from channels, private or public, are ignored.
		if chat.Type == tb.ChatChannel || chat.Type == tb.ChatChannelPrivate {
			bot.logger.WithFields(logrus.Fields{
				"chatid":    chat.ID,
				"chattitle": chat.Title,
			}).Debug("Update from a public or private channel, ignored")
			return nil
		}

		settings := chatSettings{}

		// Updates from groups need special care.
		if chat.Type != tb.ChatPrivate {
			// Update chat info in the DB (or add the chat if it is new).
			if err := bot.db.AddChat(chat); err != nil {
				bot.logger.WithError(err).Error("Failed to update my chatroom list")
				return nil
			}

			// Retrieve chat's settings.
			var err error
			settings, err = bot.getChatSettings(chat)
			if err != nil {
				bot.logger.WithError(err).Error("Failed to get chat settings")
				return nil
			}

			// Retrieve message's sender.
			sender := ctx.Sender()
			if sender == nil {
				bot.logger.WithFields(logrus.Fields{
					"chatid":    chat.ID,
					"chattitle": chat.Title,
				}).Warn("Update with nil on Sender, ignored")
				return nil
			}

			isGlobalAdmin, err := bot.db.IsBotAdmin(sender.ID)
			if err != nil {
				bot.logger.WithError(err).Error("Failed to check if the user is a global admin")
				return nil
			}

			// If the bot is not enabled (and the command is not from a global
			// admin), ignore the message.
			if !settings.BotEnabled && !isGlobalAdmin {
				bot.logger.WithFields(logrus.Fields{
					"chatid":    chat.ID,
					"chattitle": chat.Title,
				}).Debugf("Bot not enabled for chat")
				return nil
			}
		}

		// Updates from private chats don't need chat settings, only groups.
		handler(ctx, settings)
		return nil
	}
}
