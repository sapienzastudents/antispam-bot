package tbot

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

func (bot *telegramBot) simpleHandler(endpoint interface{}, fn refreshDBInfoFunc) {
	bot.telebot.Handle(endpoint, bot.metrics(bot.refreshDBInfo(fn)))
}

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
	bot.simpleHandler(tb.OnAddedToGroup, func(_ *tb.Message, _ chatSettings) {})
	bot.telebot.Handle(tb.OnUserLeft, bot.metrics(func(m *tb.Message) {
		if m.Sender.ID == bot.telebot.Me.ID {
			return
		}
		bot.refreshDBInfo(bot.onUserLeft)(m)
	}))

	// Register commands
	bot.simpleHandler("/help", bot.onHelp)
	bot.simpleHandler("/start", bot.onHelp)
	bot.simpleHandler("/groups", bot.onGroups)
	bot.simpleHandler("/gruppi", bot.onGroups)
	bot.simpleHandler("/dont", func(m *tb.Message, _ chatSettings) {
		defer func() {
			err := bot.telebot.Delete(m)
			if err != nil {
				bot.logger.WithError(err).Error("Failed to delete message")
			}
		}()

		if !m.IsReply() {
			return
		}

		_, err := bot.telebot.Reply(m.ReplyTo, "https://dontasktoask.com\nNon chiedere di chiedere, chiedi pure :)")
		if err != nil {
			bot.logger.WithError(err).Error("Failed to reply")
			return
		}
	})

	// Chat-admin commands
	bot.chatAdminHandler("/impostazioni", bot.onSettings)
	bot.chatAdminHandler("/settings", bot.onSettings)
	bot.chatAdminHandler("/terminate", bot.onTerminate)
	bot.chatAdminHandler("/reload", bot.onReloadGroup)
	bot.chatAdminHandler("/sigterm", bot.onSigTerm)

	// Global-administrative commands
	bot.globalAdminHandler("/sighup", bot.onSigHup)
	bot.globalAdminHandler("/groupscheck", bot.onGroupsPrivileges)
	bot.globalAdminHandler("/updatewww", bot.onGlobalUpdateWww)
	bot.globalAdminHandler("/gline", bot.onGLine)
	bot.globalAdminHandler("/remove_gline", bot.onRemoveGLine)
	// Global-administrative commands (legacy, we should replace them as soon as "admin fallback" feature is ready)
	bot.globalAdminHandler("/cut", bot.onCut)
	bot.globalAdminHandler("/emergency_remove", bot.onEmergencyRemove)
	bot.globalAdminHandler("/emergency_elevate", bot.onEmergencyElevate)
	bot.globalAdminHandler("/emergency_reduce", bot.onEmergencyReduce)

	// Utilities
	bot.telebot.Handle("/id", bot.metrics(bot.refreshDBInfo(func(m *tb.Message, _ chatSettings) {
		bot.botCommandsRequestsTotal.WithLabelValues("id").Inc()
		_, _ = bot.telebot.Send(m.Chat, fmt.Sprint("Your ID is: ", m.Sender.ID, "\nThis chat ID is: ", m.Chat.ID))
	})))

	bot.logger.Info("Init ok, starting bot")

	// Cache updater
	go func() {
		t := time.NewTicker(10 * time.Minute)
		for {
			<-t.C
			startms := time.Now()
			err := bot.DoCacheUpdate(bot.groupUserCount)
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
