package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ardanlabs/conf/v2"
	"gopkg.in/yaml.v3"
)

// BotConfig descrives the bot's configuration.
type BotConfig struct {
	Path     string `conf:"default:./config.yml,flag:config,short:c,help:configuration file"`
	BotToken string `conf:"default:-,flag:bot-token,short:b,help:Bot token"`
	RedisURL string `conf:"default:redis://127.0.0.1:6379,flag:redis-url,short:r,help:redis URL"`
	Git      struct {
		TmpDir     string `conf:"default:-,flag:git-dir,help:git temporary director"`
		SSHKey     string `conf:"default:-,flag:git-ssh-key,help:SSH key used with git"`
		SSHKeyPass string `conf:"default:-,flag:git-ssh-key-pass,help:SSH key's password"`
	}
	CASUpdate bool   `conf:"default:true,flag:cas-update,help:Update automatically CAS database"`
	LogLevel  string `conf:"default:info,flag:log-level,short:l,help:Minimium log level"`
}

// getConfig returns a BotConfig struct with loaded values from environment
// variables, command line arguments and a config file.
func getConfig() (BotConfig, error) {
	const prefix = "ANTISPAM" // Environment's variables prefix.
	cfg := BotConfig{}

	// Load configuration from environment variables and command line arguments.
	if help, err := conf.Parse(prefix, &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return cfg, conf.ErrHelpWanted
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Override values from YAML if specified and if it exists.
	data, err := os.ReadFile(cfg.Path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cfg, fmt.Errorf("failed to read config file, while it exists: %w")
	}
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to unmarshal config file: %w", err)
		}
	}

	// Clean these variables for security purposes.
	if err := os.Setenv(prefix+"_GIT_SSH_KEY_PASS", ""); err != nil {
		return cfg, err
	}
	if err := os.Setenv(prefix+"_GIT_SSH_KEY", ""); err != nil {
		return cfg, err
	}

	return cfg, nil
}
