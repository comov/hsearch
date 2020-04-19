package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/comov/hsearch/background"
	"github.com/comov/hsearch/bot"
	"github.com/comov/hsearch/configs"
	"github.com/comov/hsearch/storage"
)

const (
	helpCommands = "hsearch: '%s' is not a command.\n" +
		"usage: go run main.go [migrate]\n\n" +
		"By default hsearch run offer manager and telegram" +
		" bot.\nFor example: go run main.go\n\n" +
		"Commands:\n" +
		"\tmigrate - the command for run migration and create DB if not" +
		" exist. Support\n\t the flag -dir for the directory of migrations\n"
)

func main() {
	cnf, err := configs.GetConf()
	if err != nil {
		log.Fatalln("[main.GetConf] error: ", err)
	}

	db, err := storage.New(cnf)
	if err != nil {
		log.Fatalln("[main.storage.New] error: ", err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			// run migrations and stop
			err = db.Migrate(migrationPath())
			if err != nil {
				log.Fatalln("[main.storage.Migrate] error: ", err)
			}
			return
		case "withmigrate":
			// like migration bud not stop
			err = db.Migrate(migrationPath())
			if err != nil {
				log.Fatalln("[main.storage.Migrate] error: ", err)
			}
		default:
			fmt.Printf(helpCommands, os.Args[1])
			return
		}
	}

	// Telegram bot и Background manager в дальнейшем нужно запускать как отдельные
	// сервисы, а главный поток оставить следить за ними. Таким образом можно
	// сделать graceful shutdown, reload config да и просто по приколу

	telegramBot := bot.NewTelegramBot(cnf, db)

	bgm := background.NewManager(cnf, db, telegramBot)
	go bgm.StartGrabber()
	go bgm.StartMatcher()

	telegramBot.Start()
}

// migrationPath - return path to migrations file
func migrationPath() string {
	pathToMigrations := "migrations"
	if len(os.Args) > 2 && strings.HasPrefix(os.Args[2], "-dir=") {
		pathToMigrations = strings.TrimPrefix(os.Args[2], "-dir=")
	}

	if !strings.HasPrefix(pathToMigrations, "/") {
		dir, _ := os.Getwd()
		pathToMigrations = path.Join(dir, pathToMigrations)
	}
	return pathToMigrations
}
