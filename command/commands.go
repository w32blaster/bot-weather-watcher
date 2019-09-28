package command

import (
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"html"
	"log"
	"strings"
)

// ProcessCommands acts when user sent to a bot some command, for example "/command arg1 arg2"
func ProcessCommands(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {

	chatID := message.Chat.ID
	command := extractCommand(message.Command())
	log.Println("This is command " + command)

	switch command {

	case "about":
		about := `The bot is designed to notify you when the weather in selected places will be nice within 3 days.
		This can be helpful for everyone who tries to avoid rain and windy weather, such as motorcyclists, photographers, hikers and so on.
		Simply save few places you are interested in, specify range of wind speed and temperature range and the bot will notify you
		when a weather forecast matches your expectations. Have fun.

		This bot works in UK only and uses data from metoffice.gov.uk

		Please start with /start command.`
		sendMsg(bot, chatID, about)

	case "help":

		help := `This bot supports the following commands:
		     /start - shows start message
			 /help - this command
			 /add - add new place to watch
			 /now - collect the weather at current moment for all the saved places
             /forecast - show the forecast for all saved places within 3 days
			 /about - information about this bot
			 /reset - reset the inner state for current user
			 /deleteall - delete all saved places`
		sendMsg(bot, chatID, html.EscapeString(help))

	default:
		sendMsg(bot, chatID, "Sorry, I don't recognize such command: "+command+", please call /help to get full list of commands I understand")
	}

}

// properly extracts command from the input string, removing all unnecessary parts
// please refer to unit tests for details
func extractCommand(rawCommand string) string {

	command := rawCommand

	// remove slash if necessary
	if rawCommand[0] == '/' {
		command = command[1:]
	}

	// if command contains the name of our bot, remote it
	command = strings.Split(command, "@")[0]
	command = strings.Split(command, " ")[0]

	return command
}

// simply send a message to bot in Markdown format
func sendMsg(bot *tgbotapi.BotAPI, chatID int64, textMarkdown string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, textMarkdown)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	// send the message
	resp, err := bot.Send(msg)
	if err != nil {
		log.Println("bot.Send:", err, resp)
		return resp, err
	}

	return resp, err
}
