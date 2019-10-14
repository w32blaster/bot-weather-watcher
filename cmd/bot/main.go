package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/w32blaster/bot-weather-watcher/structs"

	"github.com/caarlos0/env"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/w32blaster/bot-weather-watcher/command"
)

func main() {

	// get ENV VAR
	var opts = structs.Opts{}
	if err := env.Parse(&opts); err != nil {
		panic("Can't parse ENV VARS: " + err.Error())
	}

	bot, err := tgbotapi.NewBotAPI(opts.BotToken)
	if err != nil {
		panic("Bot doesn't work. Reason: " + err.Error())
	}

	bot.Debug = opts.IsDebug

	if !opts.IsDebug {
		// Log as JSON instead of the default ASCII formatter.
		log.SetFormatter(&log.JSONFormatter{})
	}

	// run scheduler
	gocron.Every(1).Day().At("01:10").Loc(time.UTC).Do(func() {
		command.CheckWeather(bot, &opts, -1)
	})
	gocron.Start()

	log.WithField("username", bot.Self.UserName).Info("Authorized on account")
	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServe(":"+strconv.Itoa(opts.Port), nil)

	for update := range updates {

		if update.Message != nil {

			command.SetLog(log.WithFields(log.Fields{
				"app-name":  "Bot Weather Watcher",
				"user-id":   update.Message.From.ID,
				"user-name": update.Message.From.UserName,
				"chat-id":   update.Message.Chat.ID,
				"action":    "message",
				"raw-text":  update.Message.Text,
			}))

			if update.Message.IsCommand() {

				// This is a command starting with slash
				command.ProcessCommands(bot, update.Message, &opts)

			} else {

				// just a plain text
				command.ProcessPlainText(bot, update.Message)
			}

		} else if update.CallbackQuery != nil {

			command.SetLog(log.WithFields(log.Fields{
				"app-name":  "Bot Weather Watcher",
				"user-id":   update.CallbackQuery.From.ID,
				"user-name": update.CallbackQuery.From.UserName,
				"action":    "button-clicked",
				"raw-text":  update.CallbackQuery.Data,
			}))

			// this is the callback after a button click
			command.ProcessButtonCallback(bot, update.CallbackQuery, &opts)

		} else if update.InlineQuery != nil {

			command.SetLog(log.WithFields(log.Fields{
				"app-name":  "Bot Weather Watcher",
				"user-id":   update.InlineQuery.From.ID,
				"user-name": update.InlineQuery.From.UserName,
				"action":    "inline-query",
				"raw-text":  update.InlineQuery.Query,
			}))

			// this is inline query (it's like a suggestion while typing)
			command.ProcessInlineQuery(bot, update.InlineQuery)
		}

	}

}
