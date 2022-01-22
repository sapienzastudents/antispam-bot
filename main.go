package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/database"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/cas"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/i18n"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/bot"

	"github.com/ardanlabs/conf/v2"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// AppVersion is the app version injected by the compiler
var AppVersion = "dev"

// BuildDate is the app build date injected by the compiler
var BuildDate = "n/a"

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}

// Initializes and starts the bot.
func run() error {
	// Load configuration and defaults.
	cfg, err := getConfig()
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return nil
		}
		return err
	}

	// Inizialite logger.
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableQuote:     true,
		PadLevelText:     true,
	})
	log.SetOutput(os.Stdout)
	lvl, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	log.SetLevel(lvl)

	log.Info("Antispam Telegram Bot, version ", AppVersion, " build ", BuildDate)

	log.Debugf("Loaded configuration: %+v", cfg)

	// Initialize redis connection.
	log.Info("Initializing redis DB connection")

	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("failed to parse redis URL: %w", err)
	}
	redisDB := redis.NewClient(redisOptions)
	if err := redisDB.Ping(context.TODO()).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis server")
	}

	log.Info("Initializing database")
	botdb, err := database.New(redisDB)
	if err != nil {
		return fmt.Errorf("failed to create DB connection: %w", err)
	}
	if cfg.GlobalAdmin == 0 {
		log.Warn("No default bot admin given, some functionalities cannot be guaranteed")
	} else if err := botdb.AddBotAdmin(cfg.GlobalAdmin); err != nil {
		return fmt.Errorf("failed add default global admin: %w", err)
	}

	// Initialize CAS database.
	log.Info("Initializing CAS database")
	casDB, err := cas.New(cfg.CASUpdate, log, nil)
	if err != nil {
		return fmt.Errorf("failed to create CAS database: %w", err)
	}

	// Initialize i18n.
	log.Info("Initializing i18n support")
	bundle, err := i18n.New(log)
	if err != nil {
		return fmt.Errorf("failed to create i18n bundle: %w", err)
	}

	// Initialize Telegram bot.
	log.Info("Initializing Telegram bot connection")
	bot, err := bot.New(bot.Options{
		Logger:              log,
		Database:            botdb,
		Token:               cfg.BotToken,
		CAS:                 casDB,
		Bundle:              bundle,
		GitTemporaryDir:     cfg.Git.TmpDir,
		GitSSHKeyFile:       cfg.Git.SSHKey,
		GitSSHKeyPassphrase: cfg.Git.SSHKeyPass,
	})
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	go func() {
		// Temporary HTTP Server for metrics.
		log.Infof("Starting HTTP server for metrics on 0.0.0.0:3000")
		http.Handle("/metrics", bot.MetricsHandler())
		_ = http.ListenAndServe(":3000", nil)
	}()

	// Create buffered channel to catch SIGTERM signal.
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM)

	go func() {
		<-sigchan
		if err := bot.Close(); err != nil {
			log.WithError(err).Error("failed to gracefully stop bot")
		}
	}()

	if err := bot.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start bot")
	}
	return nil
}
