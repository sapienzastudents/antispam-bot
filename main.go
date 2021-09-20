package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
)

// AppVersion is the app version injected by the compiler
var AppVersion = "dev"

var b *tb.Bot = nil
var logger *logrus.Entry = nil
var botdb botdatabase.BOTDatabase = nil

var globaleditcat = cache.New(1*time.Minute, 1*time.Minute)

func main() {
	rand.Seed(time.Now().UnixNano())
	_ = godotenv.Load()
	var err error

	go mainHTTP()

	/*discordbot, err := NewDiscordBot()
	if err != nil {
		panic(err)
	}
	err = discordbot.Test()
	if err != nil {
		panic(err)
	}

	time.Sleep(60 * time.Second)

	discordbot.Stop()
	return*/

	// Logging init
	logrusLogger := logrus.New()
	logrusLogger.SetOutput(os.Stdout)
	logrusLogger.SetLevel(logrus.DebugLevel) // TODO
	// TODO: logger.SetFormatter(&logrus.JSONFormatter{})

	hostname, _ := os.Hostname()
	logger = logrusLogger.WithFields(logrus.Fields{
		"hostname": hostname,
	})

	logger.Info("Initializing...")
	// Initial checks
	if os.Getenv("BOT_TOKEN") == "" {
		logger.Fatal("BOT_TOKEN environment variable not set - exiting")
		return
	}
	if os.Getenv("REDIS_URL") == "" {
		logger.Fatal("REDIS_URL environment variable not set - exiting")
		return
	}

	// Initialize Redis database
	botdb, err = botdatabase.New(logger)
	if err != nil {
		logger.Fatalf("Unable to initialize the database, exiting: %s", err.Error())
		return
	}

	// Initialize bot library
	b, err = tb.NewBot(tb.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		logger.Fatal(err)
		return
	}

	// Registering internal/utils handlers (mostly for: spam detection, chat refresh)
	b.Handle(tb.OnVoice, metrics(refreshDBInfo(onAnyMessage)))
	b.Handle(tb.OnVideo, metrics(refreshDBInfo(onAnyMessage)))
	b.Handle(tb.OnEdited, metrics(refreshDBInfo(onAnyMessage)))
	b.Handle(tb.OnDocument, metrics(refreshDBInfo(onAnyMessage)))
	b.Handle(tb.OnAudio, metrics(refreshDBInfo(onAnyMessage)))
	b.Handle(tb.OnPhoto, metrics(refreshDBInfo(onAnyMessage)))
	b.Handle(tb.OnText, metrics(refreshDBInfo(onAnyMessage)))
	b.Handle(tb.OnSticker, metrics(refreshDBInfo(onAnyMessage)))
	b.Handle(tb.OnAnimation, metrics(refreshDBInfo(onAnyMessage)))
	b.Handle(tb.OnUserJoined, metrics(refreshDBInfo(onUserJoined)))
	b.Handle(tb.OnAddedToGroup, metrics(refreshDBInfo(func(_ *tb.Message, _ botdatabase.ChatSettings) {})))
	b.Handle(tb.OnUserLeft, metrics(func(m *tb.Message) {
		if m.Sender.ID == b.Me.ID {
			return
		}
		refreshDBInfo(onUserLeft)(m)
	}))

	// Register commands
	b.Handle("/help", metrics(refreshDBInfo(onHelp)))
	b.Handle("/start", metrics(refreshDBInfo(onHelp)))
	b.Handle("/groups", metrics(refreshDBInfo(onGroups)))
	b.Handle("/gruppi", metrics(refreshDBInfo(onGroups)))
	b.Handle("/dont", metrics(refreshDBInfo(func(m *tb.Message, _ botdatabase.ChatSettings) {
		defer func() {
			err = b.Delete(m)
			if err != nil {
				logger.WithError(err).Error("Failed to delete message")
			}
		}()

		if !m.IsReply() {
			return
		}

		_, err := b.Reply(m.ReplyTo, "https://dontasktoask.com\nNon chiedere di chiedere, chiedi pure :)")
		if err != nil {
			logger.WithError(err).Error("Failed to reply")
			return
		}
	})))

	// Chat-admin commands
	b.Handle("/impostazioni", metrics(refreshDBInfo(checkGroupAdmin(onSettings))))
	b.Handle("/settings", metrics(refreshDBInfo(checkGroupAdmin(onSettings))))
	b.Handle("/terminate", metrics(refreshDBInfo(checkGroupAdmin(onTerminate))))
	b.Handle("/reload", metrics(refreshDBInfo(checkGroupAdmin(onReloadGroup))))
	b.Handle("/sigterm", metrics(refreshDBInfo(checkGroupAdmin(onSigTerm))))

	// Global-administrative commands
	b.Handle("/sighup", metrics(checkGlobalAdmin(refreshDBInfo(onSigHup))))
	b.Handle("/groupscheck", metrics(checkGlobalAdmin(refreshDBInfo(onGroupsPrivileges))))
	b.Handle("/version", metrics(checkGlobalAdmin(refreshDBInfo(onVersion))))
	b.Handle("/updatewww", metrics(checkGlobalAdmin(refreshDBInfo(onGlobalUpdateWww))))
	b.Handle("/gline", metrics(checkGlobalAdmin(refreshDBInfo(onGLine))))
	b.Handle("/remove_gline", metrics(checkGlobalAdmin(refreshDBInfo(onRemoveGLine))))
	// Global-administrative commands (legacy, we should replace them as soon as "admin fallback" feature is ready)
	b.Handle("/cut", metrics(checkGlobalAdmin(refreshDBInfo(onCut))))
	b.Handle("/emergency_remove", metrics(checkGlobalAdmin(refreshDBInfo(onEmergencyRemove))))
	b.Handle("/emergency_elevate", metrics(checkGlobalAdmin(refreshDBInfo(onEmergencyElevate))))
	b.Handle("/emergency_reduce", metrics(checkGlobalAdmin(refreshDBInfo(onEmergencyReduce))))

	// Utilities
	b.Handle("/id", metrics(refreshDBInfo(func(m *tb.Message, _ botdatabase.ChatSettings) {
		botCommandsRequestsTotal.WithLabelValues("id").Inc()
		_, _ = b.Send(m.Chat, fmt.Sprint("Your ID is: ", m.Sender.ID, "\nThis chat ID is: ", m.Chat.ID))
	})))

	logger.Info("Init ok, starting bot")

	// Cache updater
	go func() {
		t := time.NewTicker(10 * time.Minute)
		for {
			<-t.C
			startms := time.Now()
			err := botdb.DoCacheUpdate(b, groupUserCount)
			if err != nil {
				logger.WithError(err).Error("error cycling for data refresh")
			}
			backgroundRefreshElapsed.Set(float64(time.Since(startms) / time.Millisecond))
		}
	}()

	if os.Getenv("DISABLE_CAS") == "" {
		go casWorker()
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM)
	go func() {
		<-sigchan
		b.Stop()
	}()

	// Let's go!
	b.Start()
}

// refreshDBInfo wrapper is refreshing the info for the chat in the database
// (due the fact that Telegram APIs does not support listing chats)
func refreshDBInfo(actionHandler func(*tb.Message, botdatabase.ChatSettings)) func(m *tb.Message) {
	return func(m *tb.Message) {
		if !m.Private() {
			err := botdb.UpdateMyChatroomList(m.Chat)
			if err != nil {
				logger.WithError(err).Error("Cannot update my chatroom list")
				return
			}

			settings, err := botdb.GetChatSetting(b, m.Chat)
			if err != nil {
				logger.WithError(err).Error("Cannot get chat settings")
			} else if !settings.BotEnabled && !botdb.IsGlobalAdmin(m.Sender) {
				logger.Debugf("Bot not enabled for chat %d %s", m.Chat.ID, m.Chat.Title)
			} else {
				actionHandler(m, settings)
			}
		} else {
			actionHandler(m, botdatabase.ChatSettings{})
		}
	}
}

// checkGlobalAdmin is a "firewall" for global admin only functions
func checkGlobalAdmin(actionHandler func(m *tb.Message)) func(m *tb.Message) {
	return func(m *tb.Message) {
		if !botdb.IsGlobalAdmin(m.Sender) {
			return
		}
		actionHandler(m)
	}
}

// checkGroupAdmin is a "firewall" for group admin only functions
func checkGroupAdmin(actionHandler func(*tb.Message, botdatabase.ChatSettings)) func(*tb.Message, botdatabase.ChatSettings) {
	return func(m *tb.Message, settings botdatabase.ChatSettings) {
		if m.Private() || (!m.Private() && settings.ChatAdmins.IsAdmin(m.Sender)) || botdb.IsGlobalAdmin(m.Sender) {
			actionHandler(m, settings)
			return
		}
		_ = b.Delete(m)
		msg, _ := b.Send(m.Chat, "Sorry, only group admins can use this command")
		setMessageExpiration(msg, 10*time.Second)
	}
}
