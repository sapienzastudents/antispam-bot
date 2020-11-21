package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/botdatabase"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var APP_VERSION = "dev"

var b *tb.Bot = nil
var logger *logrus.Entry = nil
var botdb botdatabase.BOTDatabase = nil

var globaleditcat = map[int]int64{}

func main() {
	_ = godotenv.Load()
	var err error

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
	b.Handle(tb.OnVoice, RefreshDBInfo(onAnyMessage))
	b.Handle(tb.OnVideo, RefreshDBInfo(onAnyMessage))
	b.Handle(tb.OnEdited, RefreshDBInfo(onAnyMessage))
	b.Handle(tb.OnDocument, RefreshDBInfo(onAnyMessage))
	b.Handle(tb.OnAudio, RefreshDBInfo(onAnyMessage))
	b.Handle(tb.OnPhoto, RefreshDBInfo(onAnyMessage))
	b.Handle(tb.OnText, RefreshDBInfo(onAnyMessage))
	b.Handle(tb.OnUserJoined, RefreshDBInfo(onUserJoined))
	b.Handle(tb.OnAddedToGroup, RefreshDBInfo(func(_ *tb.Message, _ botdatabase.ChatSettings) {}))
	b.Handle(tb.OnUserLeft, RefreshDBInfo(onUserLeft))

	// Register commands
	b.Handle("/help", RefreshDBInfo(onHelp))
	b.Handle("/start", RefreshDBInfo(onHelp))
	b.Handle("/groups", RefreshDBInfo(onGroups))

	// Chat-admin commands
	b.Handle("/settings", RefreshDBInfo(CheckGroupAdmin(onSettings)))
	b.Handle("/terminate", RefreshDBInfo(CheckGroupAdmin(onTerminate)))
	b.Handle("/reload", RefreshDBInfo(CheckGroupAdmin(onReloadGroup)))

	// Global-administrative commands
	b.Handle("/emergency_remove", CheckGlobalAdmin(RefreshDBInfo(onEmergencyRemove)))
	b.Handle("/emergency_elevate", CheckGlobalAdmin(RefreshDBInfo(onEmergencyElevate)))
	b.Handle("/emergency_reduce", CheckGlobalAdmin(RefreshDBInfo(onEmergencyReduce)))
	b.Handle("/sighup", CheckGlobalAdmin(RefreshDBInfo(onSigHup)))
	b.Handle("/groupscheck", CheckGlobalAdmin(RefreshDBInfo(onGroupsPrivileges)))
	b.Handle("/version", CheckGlobalAdmin(RefreshDBInfo(onVersion)))
	b.Handle("/updatewww", CheckGlobalAdmin(RefreshDBInfo(onGlobalUpdateWww)))

	// Utilities
	b.Handle("/id", RefreshDBInfo(func(m *tb.Message, _ botdatabase.ChatSettings) {
		_, _ = b.Send(m.Chat, fmt.Sprint("Your ID is: ", m.Sender.ID, "\nThis chat ID is: ", m.Chat.ID))
	}))

	logger.Info("Init ok, starting bot")

	// Cache updater
	go func() {
		t := time.NewTicker(10 * time.Minute)
		for {
			<-t.C
			err := botdb.DoCacheUpdate(b)
			if err != nil {
				logger.WithError(err).Error("erorr cycling for data refresh")
			}
		}
	}()

	if os.Getenv("DISABLE_CAS") == "" {
		go CASWorker()
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

// This wrapper is refreshing the info for the chat in the database
// (due the fact that Telegram APIs does not support listing chats)
func RefreshDBInfo(actionHandler func(*tb.Message, botdatabase.ChatSettings)) func(m *tb.Message) {
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

func CheckGlobalAdmin(actionHandler func(m *tb.Message)) func(m *tb.Message) {
	return func(m *tb.Message) {
		if !botdb.IsGlobalAdmin(m.Sender) {
			return
		}
		actionHandler(m)
	}
}

func CheckGroupAdmin(actionHandler func(*tb.Message, botdatabase.ChatSettings)) func(*tb.Message, botdatabase.ChatSettings) {
	return func(m *tb.Message, settings botdatabase.ChatSettings) {
		if m.Private() || (!m.Private() && settings.ChatAdmins.IsAdmin(m.Sender)) || botdb.IsGlobalAdmin(m.Sender) {
			actionHandler(m, settings)
			return
		}
		_, _ = b.Send(m.Chat, "Sorry, only group admins can use this command")
	}
}
