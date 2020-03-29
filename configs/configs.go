package configs

import (
	"time"

	"github.com/caarlos0/env/v6"
)

// Config - структура содержащая все изменяемые конфигурации приложения
type Config struct {
	// TODO: нейминг страдает
	MaxPage           int    `env:"MAXPAGE"`
	TelegramToken     string `env:"TOKEN"`
	ParserSleepTime   string `env:"PARSER_SLEEP_TIME"`
	SkipTimeString    string `env:"SKIP_TIME"`
	FreshOffersString string `env:"FRESH_ORDER"`
	AdminChatId       int64  `env:"ADMIN_CHAT_ID"`

	ManagerDelay time.Duration
	SkipTime     time.Duration
	FreshOffers  time.Duration
}

// GetConf - возвращет конфигурацию приложения
func GetConf() (*Config, error) {
	cfg := &Config{
		// здесь можно выставить параметры по умолчанию
		MaxPage:           2,
		ParserSleepTime:   "1m",
		SkipTimeString:    "3m",
		FreshOffersString: "2m",
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	// ManagerDelay
	// Здесь немного не явный момент, в настройках мы задаем время задержки как
	// строка 1m илил 12h, потом парсим в time.Duration
	cfg.ManagerDelay, err = time.ParseDuration(cfg.ParserSleepTime)
	if err != nil {
		return nil, err
	}

	// SkipTime
	cfg.SkipTime, err = time.ParseDuration(cfg.SkipTimeString)
	if err != nil {
		return nil, err
	}

	// FreshOffers
	cfg.FreshOffers, err = time.ParseDuration(cfg.FreshOffersString)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
