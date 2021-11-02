package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/botdatabase"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/cas"
	"gitlab.com/sapienzastudents/antispam-telegram-bot/service/tbot"

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

// Bot configuration.
type Config struct {
	BotToken      string `conf:"default:-,flag:bot-token,short:b,help:Bot token"`
	RedisURL      string `conf:"default:redis://127.0.0.1:6379,flag:redis-url,short:r,help:redis URL"`
	GitTmpDir     string `conf:"default:-,flag:git-dir,help:git temporary director"`
	GitSSHKey     string `conf:"default:-,flag:git-ssh-key,help:SSH key used with git"`
	GitSSHKeyPass string `conf:"default:-,flag:git-ssh-key-pass,help:SSH key's password"`
	CASUpdate     bool   `conf:"default:true,flag:cas-update,help:Update automatically CAS database"`
}

// Returns the loaded configuration, read from both command line and environment
// variables.
func getConfig() (Config, error) {
	cfg := Config{}
	help, err := conf.Parse("", &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Clean these variables for security purposes.
	if err := os.Setenv("GIT_SSH_KEY_PASS", ""); err != nil {
		return cfg, err
	}
	if err := os.Setenv("GIT_SSH_KEY", ""); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Initializes and starts the bot.
func run() error {
	// Initialize configuration.
	cfg, err := getConfig()
	if err != nil {
		return err
	}

	// Inizialite logger.
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)

	log.Info("Antispam Telegram Bot, version ", AppVersion, " build ", BuildDate)

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
	botdb, err := botdatabase.New(redisDB)
	if err != nil {
		return fmt.Errorf("failed to create DB connection: %w", err)
	}

	// Initialize CAS database.
	log.Info("Initializing CAS database")
	casDB, err := cas.New(cfg.CASUpdate, log, nil)
	if err != nil {
		return fmt.Errorf("failed to create CAS database: %w", err)
	}

	// Initialize Telegram bot.
	log.Info("Initializing Telegram bot connection")
	bot, err := tbot.New(tbot.Options{
		Logger:              log,
		Database:            botdb,
		Token:               cfg.BotToken,
		CAS:                 casDB,
		GitTemporaryDir:     cfg.GitTmpDir,
		GitSSHKeyFile:       cfg.GitSSHKey,
		GitSSHKeyPassphrase: cfg.GitSSHKeyPass,
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
