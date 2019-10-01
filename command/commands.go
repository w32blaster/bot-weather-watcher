package command

import (
	"bytes"
	"fmt"
	"github.com/asdine/storm"
	"github.com/w32blaster/bot-weather-watcher/structs"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"html"
	"log"
	"strings"
)

const (
	DbPath = "weather.db"
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

	case "add":
		StartProcessAddingNewLocation(bot, message)

	case "start":
		sendMsg(bot, chatID, "Hey! In order to begin, you should add at least one site location where you would like to observe a weather. Click /add")

	case "locations":
		PrintSavedLocations(bot, chatID, message.From.ID)

	default:
		sendMsg(bot, chatID, "Sorry, I don't recognize such command: "+command+", please call /help to get full list of commands I understand")
	}

}

func PrintSavedLocations(bot *tgbotapi.BotAPI, chatID int64, userID int) {

	db, err := storm.Open(DbPath)
	if err != nil {
		log.Printf("Error! Can't open the database, the error is %s", err.Error())
		sendMsg(bot, chatID, "Sorry, internal error occurred, can't open a database. Please try again later.")
		return
	}
	defer db.Close()

	var locations []structs.UsersLocationBookmark
	db.Find("UserID", userID, &locations)

	var buffer bytes.Buffer
	buffer.WriteString("Saved locations: ")
	for _, e := range locations {
		buffer.WriteString("● ")
		buffer.WriteString(e.LocationID)
		buffer.WriteString("\n")
	}

	sendMsg(bot, chatID, buffer.String())
}

// Initiate the process of adding a new location, create a new state
func StartProcessAddingNewLocation(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	db, err := storm.Open(DbPath)
	if err != nil {
		log.Printf("Error! Can't open the database, the error is %s", err.Error())
		sendMsg(bot, message.Chat.ID, "Sorry, internal error occurred, can't open a database. Please try again later.")
		return
	}

	defer db.Close()

	// to make sure we start from the beginning, clear all previous states if any
	DeleteStateForUser(db, message.From.ID)

	// and now start a new state machine
	sm, err := LoadStateMachineFor(message.From.ID, db)
	if err != nil {
		log.Printf("Can't initiate a state machine. Error is %s", err.Error())
		sendMsg(bot, message.Chat.ID, "Sorry, internal error occurred, please trt again later")
		return
	}

	if err := sm.CreateNewBookmark(); err != nil {
		log.Println(err.Error())
		sendMsg(bot, message.Chat.ID, "Sorry, internal error occurred, please trt again later")
		return
	}

	sendMsg(bot, message.Chat.ID, "Ok, let's add a location where you want to monitor a weather. "+
		"Start typing name following by the bot name and suggestions will appear. \n"+
		"Example: `@weather_observer_bot London`")
}

// Process a general text. The context should be retrieved from state machine
func ProcessPlainText(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	db, err := storm.Open(DbPath)
	defer db.Close()

	if err != nil {
		sendMsg(bot, message.Chat.ID, "Ouch, this is internal error, sorry")
		log.Println("Error opening the database, err: " + err.Error())
		return
	}

	stateMachine, err := LoadStateMachineFor(message.From.ID, db)
	if err != nil {
		sendMsg(bot, message.Chat.ID, "Ouch, this is internal error, sorry")
		log.Println("Error opening the database, err: " + err.Error())
		return
	}

	msg := stateMachine.ProcessNextState(message.Text)
	sendMsg(bot, message.Chat.ID, msg)
}

func ProcessInlineQuery(bot *tgbotapi.BotAPI, inlineQuery *tgbotapi.InlineQuery) {
	db, err := storm.Open(DbPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	// firstly, make query to TFL
	searchQuery := inlineQuery.Query
	var locations []structs.SiteLocation
	db.Prefix("Name", searchQuery, &locations, storm.Limit(10))

	var answers []interface{}

	for _, loc := range locations {

		// Build one line for inline answer (one result)
		strLocID := fmt.Sprint(loc.ID)
		answer := tgbotapi.NewInlineQueryResultArticleHTML(strLocID, loc.Name, strLocID)
		descr := loc.AuthArea + ", " + strings.ToUpper(loc.Region) + ", UK"
		if len(loc.NationalPark) > 0 {
			descr = loc.NationalPark + ", " + descr
		}
		answer.Description = html.EscapeString(descr)

		answers = append(answers, answer)
	}

	answer := tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.ID,
		CacheTime:     3,
		Results:       answers,
	}

	if resp, err := bot.AnswerInlineQuery(answer); err != nil {
		log.Fatal("ERROR! bot.answerInlineQuery:", err, resp)
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
