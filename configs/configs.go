package configs

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/getsentry/sentry-go"
	_ "github.com/joho/godotenv/autoload"
)

// Config - the structure that contains all the customizable application
//  configurations
type Config struct {
	Release         string
	ParserFrequency string `env:"PARSER_FREQUENCY"`
	OrderRelevance  string `env:"ORDER_RELEVANCE"`
	TelegramToken   string `env:"T_TOKEN"`
	TelegramChatId  int64  `env:"T_CHAT_ID"`
	PgPassword      string `env:"GO_DB_PASSWORD"`
	PgHost          string `env:"GO_DB_HOST"`
	PgPort          int32  `env:"GO_DB_PORT"`

	FrequencyTime time.Duration
	RelevanceTime time.Duration

	ExpireDays   int
	PgConnString string
}

// GetConf - returns the application configuration
func GetConf() (*Config, error) {
	cfg := &Config{
		ParserFrequency: "1m",
		OrderRelevance:  "2m",
		PgPassword:      "hsearch",
		PgHost:          "localhost",
		PgPort:          5432,
		ExpireDays:      7,
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	err = sentry.Init(sentry.ClientOptions{
		SampleRate: 0.5,
	})

	if err != nil {
		return nil, err
	}

	//// In the settings we set the delay time as line 1m or 12h, then parse
	////  in time.

	// RelevanceTime
	cfg.FrequencyTime, err = time.ParseDuration(cfg.ParserFrequency)
	if err != nil {
		return nil, err
	}

	// RelevanceTime
	cfg.RelevanceTime, err = time.ParseDuration(cfg.OrderRelevance)
	if err != nil {
		return nil, err
	}

	cfg.PgConnString = fmt.Sprintf("user=hsearch password=%s host=%s port=%d dbname=hsearch",
		cfg.PgPassword,
		cfg.PgHost,
		cfg.PgPort,
	)

	return cfg, nil
}
