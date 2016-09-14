package main

import (
	"flag"
	"log"

	"os"
	"os/signal"
	"syscall"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type DataMapType map[string]map[string]int

var DataDict = make(DataMapType)

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

	flag.Parse()

	if botToken == "" {
		log.Panic("Missing Bot Token")
	}

	ds, err := loadDataset(datasetPath)
	if err != nil {
		log.Panic(err)
		return
	}
	DataDict = ds

	if importPath != "" {
		importFile(importPath)
	}

	// Init bot
	go log.Panic(runBot(botToken, devMode))

	// Shutdown Handler
	var done = make(chan bool, 1)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		err := saveDataset(datasetPath)
		if err != nil {
			log.Println("Error saving dataset:", err)
		}

		done <- true
	}()
	<-done
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
		}

		go trainMessage(update.Message.Text)
	}
	return nil
}

func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message.Command() == "test" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "yo")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}

	if update.Message.Command() == "markov" {
		log.Println(update.Message.CommandArguments())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, generateMarkovResponse(update.Message.CommandArguments()))
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}
}
