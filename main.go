package main

import (
	"flag"
	"log"
	"math/rand"
	"sync"
	"time"

	"os"
	"os/signal"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"syscall"
)

type DataMapType map[[2]string][]string

type DataMap struct {
	sync.RWMutex
	Map DataMapType
}

var DataDict = DataMap{Map: make(DataMapType)}

func main() {
	// Flags
	var botToken string
	var importPath string
	var devMode bool
	var datasetPath string

	// Read flags

	flag.StringVar(&datasetPath, "file", "data.mrkv", "Path to markov dataset file")
	flag.StringVar(&botToken, "token", "", "Telegram token for bot")
	flag.StringVar(&importPath, "import", "", "Import raw text logs")
	flag.BoolVar(&devMode, "dev", false, "If enabled, bot debug mode is true")
	flag.IntVar(&MaxMessageLen, "maxlen", 100, "The max length of the generated messages")

	flag.Parse()

	rand.Seed(time.Now().Unix())

	if botToken == "" {
		log.Panic("Missing Bot Token")
	}

	ds, err := loadDataset(datasetPath)
	if err != nil {
		log.Panic(err)
		return
	}
	DataDict.Map = ds

	if importPath != "" {
		importFile(importPath)
	}


	// Init bot
	go runBot(botToken, devMode)

	// Shutdown Handler
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Starting Safe Quit")
		err := saveDataset(datasetPath)
		if err != nil {
			log.Println("Error saving dataset:", err)
		}
		done <- true
	}()

	<-done
	log.Println("Ended Safe Quit")

}

func runBot(token string, mode bool) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	bot.Debug = mode

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			go handleCommand(bot, update)
		} else {
			go trainMessage(update.Message.Text)
		}
	}
	return nil
}

func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message.Command() == "test" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "yo")
		msg.ReplyToMessageID = update.Message.MessageID
		_, sendErr := bot.Send(msg)
		if sendErr != nil {
			log.Println(sendErr)
		}
		return
	}

	if update.Message.Command() == "markov" {
		log.Println(update.Message.CommandArguments())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, generateMarkovResponse(update.Message.CommandArguments()))
		msg.ReplyToMessageID = update.Message.MessageID
		_, sendErr := bot.Send(msg)
		if sendErr != nil {
			log.Println(sendErr)
		}
		return
	}
}
