package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/op/go-logging"
	tb "gopkg.in/tucnak/telebot.v2"
	"os"
	"time"
)

var APP_VERSION = "dev"

var b *tb.Bot = nil
var logger *logging.Logger = nil
var botdb BOTDatabase = nil

func main() {
	_ = godotenv.Load()
	var err error

	// Logging init
	logger = logging.MustGetLogger("goinfostudapi")

	// To add logging to file, edit this section and implement a system based on environmental variables
	backend1Leveled := logging.AddModuleLevel(logging.NewLogBackend(os.Stdout, "", 0))
	backend1Leveled.SetLevel(logging.DEBUG, "")
	logger.SetBackend(backend1Leveled)

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
	botdb, err = NewBotDatabase()
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
	b.Handle(tb.OnAddedToGroup, RefreshDBInfo(func(_ *tb.Message, _ ChatSettings) {}))
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
	b.Handle("/sighup", CheckGlobalAdmin(RefreshDBInfo(onSigHup)))
	b.Handle("/groupscheck", CheckGlobalAdmin(RefreshDBInfo(onGroupsPrivileges)))
	b.Handle("/version", CheckGlobalAdmin(RefreshDBInfo(onVersion)))

	// Utilities
	b.Handle("/id", RefreshDBInfo(func(m *tb.Message, _ ChatSettings) {
		_, _ = b.Send(m.Chat, fmt.Sprint("Your ID is: ", m.Sender.ID, "\nThis chat ID is: ", m.Chat.ID))
	}))

	logger.Info("Init ok, starting bot")

	// Cache updater
	go func() {
		t := time.NewTicker(10 * time.Minute)
		for {
			<-t.C
			err := botdb.DoCacheUpdate()
			if err != nil {
				logger.Critical(err)
			}
		}
	}()

	if os.Getenv("DISABLE_CAS") == "" {
		go CASWorker()
	}

	// Let's go!
	b.Start()
}

// This wrapper is refreshing the info for the chat in the database
// (due the fact that Telegram APIs does not support listing chats)
func RefreshDBInfo(actionHandler func(*tb.Message, ChatSettings)) func(m *tb.Message) {
	return func(m *tb.Message) {
		if !m.Private() {
			err := botdb.UpdateMyChatroomList(m.Chat)
			if err != nil {
				logger.Critical("Cannot update my chatroom list:", err)
				return
			}

			settings, err := botdb.GetChatSetting(m.Chat)
			if err != nil {
				logger.Critical("Cannot get chat settings:", err)
			} else if !settings.BotEnabled && !botdb.IsGlobalAdmin(m.Sender) {
				logger.Debugf("Bot not enabled for chat %d %s", m.Chat.ID, m.Chat.Title)
			} else {
				actionHandler(m, settings)
			}
		} else {
			actionHandler(m, ChatSettings{})
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

func CheckGroupAdmin(actionHandler func(*tb.Message, ChatSettings)) func(*tb.Message, ChatSettings) {
	return func(m *tb.Message, settings ChatSettings) {
		if !settings.ChatAdmins.IsAdmin(m.Sender) && !botdb.IsGlobalAdmin(m.Sender) {
			_, _ = b.Send(m.Chat, "Sorry, only group admins can use this command")
			return
		}
		actionHandler(m, settings)
	}
}
