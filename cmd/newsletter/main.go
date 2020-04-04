package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/comov/hsearch/bot"
	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/storage"
	"github.com/comov/hsearch/structs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var releases map[string]string

func main() {
	cnf, err := configs.GetConf()
	if err != nil {
		log.Fatalln("[main.GetConf] error: ", err)
	}

	db, err := storage.New(cnf)
	if err != nil {
		log.Fatalln("[main.storage.New] error: ", err)
	}

	chats, err := db.ReadChatsForMatching(-1)
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
			log.Fatalln("[tBot.Send] error: ", err)
		}
		updateChatsInfo(db, chat, msg)
	}
}

// todo: remove after update chat info
func updateChatsInfo(st *storage.Connector, chat *structs.Chat, m tgbotapi.Message) {
	if chat.Type == "" {
		title := m.Chat.Title
		if m.Chat.IsPrivate() {
			title = fmt.Sprintf("%s %s", m.Chat.FirstName, m.Chat.LastName)
		}
		err := st.UpdateChat(m.Chat.ID, title, m.Chat.Type)
		if err != nil {
			fmt.Printf("[UpdateChat] %s with an error: %s", m.Chat.UserName, err)
		}
	}
}
