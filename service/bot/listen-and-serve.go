package bot

import (
	"fmt"
	"time"

	tb "gopkg.in/telebot.v3"
)

// simpleHandler adds the given function as handler for the given endpoint. It
// wraps fn with metrics and chat cache refresh. Use this for registering bot
// actions when no filters are required.
func (bot *telegramBot) simpleHandler(endpoint interface{}, fn contextualChatSettingsFunc) {
	bot.telebot.Handle(endpoint, bot.metrics(bot.refreshDBInfo(fn)))
}

// ListenAndServe register all handlers and starts the bot. It's a blocker function: terminates when the bot is stopped
func (bot *telegramBot) ListenAndServe() error {
	// Registering internal/utils handlers (mostly for: spam detection, chat refresh)
	bot.simpleHandler(tb.OnVoice, bot.onAnyMessage)
	bot.simpleHandler(tb.OnVideo, bot.onAnyMessage)
	bot.simpleHandler(tb.OnEdited, bot.onAnyMessage)
	bot.simpleHandler(tb.OnDocument, bot.onAnyMessage)
	bot.simpleHandler(tb.OnAudio, bot.onAnyMessage)
	bot.simpleHandler(tb.OnPhoto, bot.onAnyMessage)
	bot.simpleHandler(tb.OnText, bot.onAnyMessage)
	bot.simpleHandler(tb.OnSticker, bot.onAnyMessage)
	bot.simpleHandler(tb.OnAnimation, bot.onAnyMessage)
	bot.simpleHandler(tb.OnUserJoined, bot.onUserJoined)
	bot.simpleHandler(tb.OnAddedToGroup, bot.onAddedToGroup)
	bot.simpleHandler(tb.OnUserLeft, bot.onUserLeft)

	// Register general commands
	bot.simpleHandler("/help", bot.onHelp)
	bot.simpleHandler("/start", bot.onHelp)
	bot.simpleHandler("/groups", bot.onGroups)
	bot.simpleHandler("/gruppi", bot.onGroups)
	bot.simpleHandler("/dont", bot.onDont)

	// Chat-admin commands
	bot.chatAdminHandler("/impostazioni", bot.onSettings)
	bot.chatAdminHandler("/settings", bot.onSettings)
	bot.chatAdminHandler("/terminate", bot.onTerminate)
	bot.chatAdminHandler("/reload", bot.onReloadGroup)
	bot.chatAdminHandler("/sigterm", bot.onSigTerm)

	// Global-administrative commands
	bot.globalAdminHandler("/sighup", bot.onSigHup)
	bot.globalAdminHandler("/groupscheck", bot.onGroupsPrivileges)
	bot.globalAdminHandler("/updatewww", bot.onGlobalUpdateWWW)
	bot.globalAdminHandler("/gline", bot.onGLine)
	bot.globalAdminHandler("/remove_gline", bot.onRemoveGLine)

	// Utilities
	bot.simpleHandler("/id", func(ctx tb.Context, settings chatSettings) {
		bot.botCommandsRequestsTotal.WithLabelValues("id").Inc()
		m := ctx.Message()
		if m == nil {
			bot.logger.WithField("updateid", ctx.Update().ID).Warn("Update with nil on Message, ignored")
			return
		}
		_ = ctx.Send(fmt.Sprint("Your ID is: ", m.Sender.ID, "\nThis chat ID is: ", m.Chat.ID))
	})

	bot.logger.Info("Init ok, starting bot")

	// Cache updater
	go func() {
		t := time.NewTicker(10 * time.Minute)
		for {
			<-t.C
			startms := time.Now()
			err := bot.DoCacheUpdate()
			if err != nil {
				bot.logger.WithError(err).Error("error cycling for data refresh")
			}
			bot.backgroundRefreshElapsed.Set(float64(time.Since(startms) / time.Millisecond))
		}
	}()

	// Let's go!
	bot.telebot.Start()
	return nil
}
