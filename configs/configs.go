package configs

import (
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/getsentry/sentry-go"
	_ "github.com/joho/godotenv/autoload"
)

var Release string

// Config - the structure that contains all the customizable application
//  configurations
type Config struct {
	Release         string
	SentryDSN       string `env:"SENTRY_DSN"`
	ParserFrequency string `env:"PARSER_FREQUENCY"`
	OrderSkipDelay  string `env:"ORDER_SKIP_DELAY"`
	OrderRelevance  string `env:"ORDER_RELEVANCE"`
	TelegramToken   string `env:"T_TOKEN"`
	TelegramChatId  int64  `env:"T_CHAT_ID"`

	FrequencyTime time.Duration
	SkipDelayTime time.Duration
	RelevanceTime time.Duration
}

// GetConf - returns the application configuration
func GetConf() (*Config, error) {
	cfg := &Config{
		Release:         Release,
		ParserFrequency: "1m",
		OrderSkipDelay:  "3m",
		OrderRelevance:  "2m",
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:        cfg.SentryDSN,
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

	// SkipDelayTime
	cfg.SkipDelayTime, err = time.ParseDuration(cfg.OrderSkipDelay)
	if err != nil {
		return nil, err
	}

	// RelevanceTime
	cfg.RelevanceTime, err = time.ParseDuration(cfg.OrderRelevance)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
