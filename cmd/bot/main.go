package main

import (
	"github.com/getsentry/sentry-go"
	"net/http"
	"strconv"
	"time"

	"github.com/w32blaster/bot-weather-watcher/structs"

	"github.com/caarlos0/env"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
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

	sentry.Init(sentry.ClientOptions{
		Dsn:   opts.SentryDSN,
		Debug: opts.IsDebug,
	})

	// run scheduler
	gocron.Every(1).Day().At("01:10").Loc(time.UTC).Do(func() {
		command.CheckWeather(bot, &opts, -1)
	})
	gocron.Start()

	sentry.CaptureMessage("Authorized on account " + bot.Self.UserName)
	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServe(":"+strconv.Itoa(opts.Port), nil)

	for update := range updates {

		if update.Message != nil {

			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetUser(sentry.User{
					ID:       strconv.Itoa(update.Message.From.ID),
					Username: update.Message.From.UserName})
				scope.SetTag("action", "message")
				scope.SetTag("raw-text", update.Message.Text)
			})

			if update.Message.IsCommand() {

				// This is a command starting with slash
				command.ProcessCommands(bot, update.Message, &opts)

			} else {

				// just a plain text
				command.ProcessPlainText(bot, update.Message)
			}

		} else if update.CallbackQuery != nil {

			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetUser(sentry.User{
					ID:       strconv.Itoa(update.Message.From.ID),
					Username: update.Message.From.UserName})
				scope.SetTag("action", "button-clicked")
				scope.SetTag("raw-text", update.CallbackQuery.Data)
			})

			// this is the callback after a button click
			command.ProcessButtonCallback(bot, update.CallbackQuery, &opts)

		} else if update.InlineQuery != nil {

			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetUser(sentry.User{
					ID:       strconv.Itoa(update.Message.From.ID),
					Username: update.Message.From.UserName})
				scope.SetTag("action", "inline-query")
				scope.SetTag("raw-text", update.InlineQuery.Query)
			})

			// this is inline query (it's like a suggestion while typing)
			command.ProcessInlineQuery(bot, update.InlineQuery)
		}

	}

}
