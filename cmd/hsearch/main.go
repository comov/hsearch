package main

import (
	"context"
	"fmt"
	"log"

	"github.com/comov/hsearch/background"
	"github.com/comov/hsearch/bot"
	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/storage"
)

var Release string

func main() {
	cnf, err := configs.GetConf()
	if err != nil {
		log.Fatalln("[main.GetConf] error: ", err)
	}
	cnf.Release = Release
	fmt.Printf("Release: %s\n", cnf.Release)

	ctx := context.Background()
	db, err := storage.New(ctx, cnf)
	if err != nil {
		log.Fatalln("[main.storage.New] error: ", err)
	}

	defer db.Close()

	// Telegram bot и Background manager в дальнейшем нужно запускать как отдельные
	// сервисы, а главный поток оставить следить за ними. Таким образом можно
	// сделать graceful shutdown, reload config да и просто по приколу

	telegramBot := bot.NewTelegramBot(cnf, db)

	bgm := background.NewManager(cnf, db, telegramBot)
	go bgm.StartGrabber()
	go bgm.StartMatcher()

	telegramBot.Start()
}
