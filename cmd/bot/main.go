package main

import (
	"github.com/jasonlvhit/gocron"
	"github.com/w32blaster/bot-weather-watcher/structs"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/caarlos0/env"
	"github.com/w32blaster/bot-weather-watcher/command"
	"gopkg.in/telegram-bot-api.v4"
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

	// run scheduler
	gocron.Every(1).Day().At("01:10").Loc(time.UTC).Do(func() {
		command.CheckWeather(bot)
	})
	gocron.Start()

	log.Printf("Authorized on account %s", bot.Self.UserName)
	updates := bot.ListenForWebhook("/" + bot.Token)

	go http.ListenAndServe(":"+strconv.Itoa(opts.Port), nil)

	for update := range updates {

		if update.Message != nil {

			if update.Message.IsCommand() {

				// This is a command starting with slash
				command.ProcessCommands(bot, update.Message, &opts)

			} else {

				// just a plain text
				command.ProcessPlainText(bot, update.Message)
			}

		} else if update.CallbackQuery != nil {

			// this is the callback after a button click
			command.ProcessButtonCallback(bot, update.CallbackQuery, &opts)

		} else if update.InlineQuery != nil {

			// this is inline query (it's like a suggestion while typing)
			command.ProcessInlineQuery(bot, update.InlineQuery)
		}

	}

}
