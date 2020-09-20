package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/comov/hsearch/bot"
	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var releases map[string]string

func main() {
	ctx := context.Background()

	cnf, err := configs.GetConf()
	if err != nil {
		log.Fatalln("[main.GetConf] error: ", err)
	}

	db, err := storage.New(ctx, cnf)
	if err != nil {
		log.Fatalln("[main.storage.New] error: ", err)
	}

	defer db.Close(ctx)

	chats, err := db.ReadChatsForMatching(ctx, -1)
	if err != nil {
		log.Fatalln("[db.ReadChatsForMatching] error: ", err)
	}

	changelogPath := os.Args[1]
	file, err := ioutil.ReadFile(changelogPath)
	if err != nil {
		log.Fatalln("[ioutil.ReadFile] error: ", err)
	}

	err = json.Unmarshal(file, &releases)
	if err != nil {
		log.Fatalln("[json.Unmarshal] error: ", err)
	}

	version := os.Args[2]
	tBot := bot.NewTelegramBot(cnf, db)
	for _, chat := range chats {
		message := tgbotapi.NewMessage(chat.Id, releases[version])
		message.ParseMode = tgbotapi.ModeMarkdown
		msg, err := tBot.Send(message)
		if err != nil {
			log.Printf("[tBot.Send] %d error: %s\n", chat.Id, err)
			continue
		}
		log.Printf("[tBot.Send] %s %s success\n", msg.Chat.FirstName, msg.Chat.LastName)
	}
}
