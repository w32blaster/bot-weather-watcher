package command

import (
	"bytes"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/w32blaster/bot-weather-watcher/structs"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/msgpack"
	"github.com/asdine/storm/q"
	"github.com/sirupsen/logrus"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var log logrus.Entry

const (
	DbPath                = "weather.db"
	LocationIDPrefix      = "LocationID:"
	Separator             = "#"
	ButtonDaysPrefix      = "D"
	ButtonLocationPrefix  = "L"
	ButtonDeleteMsgPrefix = "dM"
)

func SetLog(entry *logrus.Entry) {
	log = *entry
}

// ProcessCommands acts when user sent to a bot some command, for example "/command arg1 arg2"
func ProcessCommands(bot *tgbotapi.BotAPI, message *tgbotapi.Message, opts *structs.Opts) {

	chatID := message.Chat.ID
	command := extractCommand(message.Command())

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
			 /locations - list all the saved locations
			 /about - information about this bot
			 /deleteall - delete all saved places`
		sendMsg(bot, chatID, html.EscapeString(help))

	case "add":
		StartProcessAddingNewLocation(bot, message)

	case "start":
		sendMsg(bot, chatID, "Hey! In order to begin, you should add at least one site location where you would like to observe a weather. Click /add")

	case "locations":
		PrintSavedLocations(bot, chatID, message.From.ID)

	case "deleteall":
		DeleteLocations(bot, message)

	default:
		sendMsg(bot, chatID, "Sorry, I don't recognize such command: "+command+", please call /help to get full list of commands I understand")
	}
}

func DeleteLocations(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	if err != nil {
		log.WithError(err).Warn("Can't open database")
		sendMsg(bot, message.Chat.ID, "Sorry, internal error occurred, can't open a database. Please try again later.")
		return
	}
	defer db.Close()

	if err := DeleteAllForThisUser(db, message.From.ID); err != nil {
		log.WithError(err).WithField("user-id", message.From.ID).Error("Can't delete all user bookmarks (saved locations)")
	}

	log.WithField("user-id", message.From.ID).Info("User bookmarks were deleted")
	sendMsg(bot, message.Chat.ID, "Deleted")
}

func PrintSavedLocations(bot *tgbotapi.BotAPI, chatID int64, userID int) {

	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	if err != nil {
		log.WithError(err).Warn("Can't open database")
		sendMsg(bot, chatID, "Sorry, internal error occurred, can't open a database. Please try again later.")
		return
	}
	defer db.Close()

	var locations []structs.UsersLocationBookmark
	db.Find("UserID", userID, &locations)

	if len(locations) == 0 {
		sendMsg(bot, chatID, "No saved locations yet. Please type /add to add one")
		return
	}

	// load locations, build a map
	mapLocs := getMapOfLocations(locations, db)

	msg, _ := sendMsg(bot, chatID, "Here is your saved locations:")
	renderLocationsButtons(bot, chatID, msg.MessageID, locations, mapLocs)
}

func getMapOfLocations(locations []structs.UsersLocationBookmark, db *storm.DB) map[string]structs.SiteLocation {
	ids := make([]string, len(locations))
	for i, loc := range locations {
		ids[i] = loc.LocationID
	}
	var locs []structs.SiteLocation
	db.Select(q.In("ID", ids)).Find(&locs)
	mapLocations := make(map[string]structs.SiteLocation)
	for _, loc := range locs {
		mapLocations[loc.ID] = loc
	}

	return mapLocations
}

// Initiate the process of adding a new location, create a new state
func StartProcessAddingNewLocation(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	if err != nil {
		log.WithError(err).Warn("Can't open database")
		sendMsg(bot, message.Chat.ID, "Sorry, internal error occurred, can't open a database. Please try again later.")
		return
	}

	defer db.Close()

	// to make sure we start from the beginning, clear all previous states if any
	DeleteStateForUser(db, message.From.ID)

	// Delete all the bookmarks that this user has not finished if any
	DeleteAllUnfinishedBookmarksForThisUser(db, message.From.ID)

	// and now start a new state machine
	sm, err := LoadStateMachineFor(message.From.ID, db)
	if err != nil {
		log.WithError(err).Warn("Can't initiate a state machine")
		sendMsg(bot, message.Chat.ID, "Sorry, internal error occurred, please trt again later")
		return
	}

	if err := sm.CreateNewBookmark(message.Chat.ID); err != nil {
		log.WithError(err).Warn("Can't create a new bookmark")
		sendMsg(bot, message.Chat.ID, "Sorry, internal error occurred, please trt again later")
		return
	}

	resp, _ := sendMsg(bot, message.Chat.ID, "Ok, let's add a location where you want to monitor a weather. "+
		"Start typing name following by the bot name and suggestions will appear. \n"+
		"Example: @WeatherObserverBot London \n\n Or, click the button below")

	renderButtonThatOpensInlineQuery(bot, message.Chat.ID, resp.MessageID)
}

// Process a general text. The context should be retrieved from state machine
func ProcessPlainText(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	defer db.Close()

	if err != nil {
		sendMsg(bot, message.Chat.ID, "Ouch, this is internal error, sorry")
		log.WithError(err).Warn("Can't open database")
		return
	}

	stateMachine, err := LoadStateMachineFor(message.From.ID, db)
	if err != nil {
		sendMsg(bot, message.Chat.ID, "Ouch, this is internal error, sorry")
		log.WithError(err).Warn("Can't create state machine")
		return
	}

	msg := stateMachine.ProcessNextState(message.Text)
	sendMsg(bot, message.Chat.ID, msg)
}

func ProcessButtonCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, opts *structs.Opts) {

	// notify the telegram that we processed the button, it will turn "loading indicator" off
	defer bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
	})

	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	if err != nil {
		log.WithError(err).Warn("Can't open database")
		sendMsg(bot, callbackQuery.Message.Chat.ID, "Whoops... error :(")
		return
	}
	defer db.Close()

	// expected data is "location id # date", for example
	parts := strings.Split(callbackQuery.Data, Separator)

	if parts[0] == ButtonDaysPrefix {

		// render 3 hour charts for temp and wind, for one location within one day
		renderOneDayDetailedWeatherForecast(bot, callbackQuery, db, opts, parts[1], parts[2], parts[3])
	} else if parts[0] == ButtonLocationPrefix {

		// render table with 5 days summary for a given location
		renderWeatherForecastForOneLocation(bot, db, callbackQuery.Message.Chat.ID, callbackQuery.From.ID, opts, parts[1])
	} else if parts[0] == ButtonDeleteMsgPrefix {

		// delete message
		deleteMessage(bot, callbackQuery.Message.Chat.ID, parts[1])
	}
}

func deleteMessage(bot *tgbotapi.BotAPI, chatID int64, messageID string) {
	if intMessageID, err := strconv.Atoi(messageID); err == nil {
		msg := tgbotapi.NewDeleteMessage(chatID, intMessageID)
		bot.Send(msg)
	} else {
		log.WithError(err).WithField("message-id", messageID).
			Warn("Can't delete message, because messageID is invalid")
	}
}

func renderWeatherForecastForOneLocation(bot *tgbotapi.BotAPI, db *storm.DB, chatID int64, userID int, opts *structs.Opts, locationID string) {

	var locations []structs.UsersLocationBookmark
	db.Select(q.Eq("LocationID", locationID), q.Eq("UserID", userID)).Limit(1).Find(&locations)

	mapLocations := getMapOfLocations(locations, db)

	loc, _ := getDailyForecastFor(locations[0].LocationID, opts)

	site := mapLocations[locations[0].LocationID]
	str := fmt.Sprintf("%s %s, %s, %s UK", site.NationalPark, site.Name, site.AuthArea, strings.ToUpper(site.Region))
	str = str + drawFiveDaysTable(loc)
	str = str + "\n For detailed daily forecast per 3 hour please use buttons below:"
	resp, _ := sendMsg(bot, chatID, str)

	// render buttons with dates
	renderDetailedDatesButtons(bot, chatID, resp.MessageID, loc)
}

func renderOneDayDetailedWeatherForecast(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *storm.DB, opts *structs.Opts, locationID, selectedDate string, messageIDtoUpdate string) {

	var locations []structs.UsersLocationBookmark
	db.Select(q.Eq("LocationID", locationID), q.Eq("UserID", callbackQuery.From.ID)).Limit(1).Find(&locations)

	// format the data
	dateFormatted := "unknown date"
	if t, err := time.Parse(layoutMetofficeDate, selectedDate); err == nil {
		dateFormatted = t.Format("2 January 2006, Monday")
	}

	site := getMapOfLocations(locations, db)[locationID]

	// Title
	title := "*" + site.NationalPark + " " + site.Name + ", " + site.AuthArea + ", " + site.Region +
		", UK*\n" + dateFormatted + "\n------\n\n"

	// make request to MetOffice
	root, err := get3HoursForecastFor(locationID, opts)
	if err != nil {

		log.WithError(err).WithFields(logrus.Fields{
			"location-id":   locationID,
			"location-name": site.Name,
		}).Warn("Can't get 3 hours forecast from metoffice")

		sendMsg(bot, callbackQuery.Message.Chat.ID, "Error retrieving data from MetOffice. Try again later")
		return
	}

	// render all the data and plots
	for _, day := range root.SiteRep.Dv.Location.Periods {
		if day.Value == selectedDate {

			str := "Temperature: \n"
			str = str + printDetailedPlotsForADay(day.Rep, "T", "˚C")

			str = str + "Wind speed: \n"
			str = str + printDetailedPlotsForADay(day.Rep, "S", "mph")

			// too tall if rain is possible
			//str = str + "Precipitation Probability: \n"
			//str = str + printDetailedPlotsForADay(day.Rep, "Pp", "%ß")

			// update existing message
			if intMessageID, err := strconv.Atoi(messageIDtoUpdate); err == nil {

				msg := tgbotapi.NewEditMessageText(callbackQuery.Message.Chat.ID, intMessageID, title+str)
				msg.ParseMode = "Markdown"
				msg.DisableWebPagePreview = true

				resp, err := bot.Send(msg)
				if err != nil {
					log.Println("bot.Send:", err, resp)
					break
				}

				// render buttons with dates
				renderDetailedDatesButtons(bot, callbackQuery.Message.Chat.ID, resp.MessageID, root)
			}

			break
		}
	}
}

func ProcessInlineQuery(bot *tgbotapi.BotAPI, inlineQuery *tgbotapi.InlineQuery) {
	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	if err != nil {
		log.WithError(err).Warn("Can't open database")
		return
	}
	defer db.Close()

	// firstly, make query to TFL
	searchQuery := inlineQuery.Query
	var locations []structs.SiteLocation

	db.Select(q.Or(
		q.Re("Name", "(?i)(^| )"+searchQuery),
		q.Re("AuthArea", "(?i)(^| )"+searchQuery),
		q.Re("NationalPark", "(?i)(^| )"+searchQuery),
	)).Limit(20).OrderBy("Name").Find(&locations)

	var answers []interface{}

	for _, loc := range locations {

		// Build one line for inline answer (one result)
		strLocID := fmt.Sprint(loc.ID)
		answer := tgbotapi.NewInlineQueryResultArticleHTML(strLocID, loc.Name, LocationIDPrefix+strLocID)
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

		bytes, err2 := resp.Result.MarshalJSON()
		if err2 != nil {
			log.WithError(err2).Warn("I tried to parse JSON that was returned from bot, but couldn't. Empty array is used")
			bytes = []byte{}
		}

		log.WithError(err).WithFields(logrus.Fields{
			"description": resp.Description,
			"raw-message": string(bytes),
		}).Error("bot.answerInlineQuery")
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
		log.WithError(err).
			Error("error when we tried to send a message using bot.Send:")
		return resp, err
	}

	return resp, err
}

func renderKeyboardButtonActivateQuery(message string) *tgbotapi.InlineKeyboardButton {
	emtpyString := ""
	button := tgbotapi.InlineKeyboardButton{
		Text:                         message,
		SwitchInlineQueryCurrentChat: &emtpyString,
	}
	return &button
}

// renders the button "search for location"
func renderButtonThatOpensInlineQuery(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			*renderKeyboardButtonActivateQuery(" 🔍 Search for location"),
		})

	keyboardMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, keyboard)
	bot.Send(keyboardMsg)
}

// renders the buttons for saved locations
func renderLocationsButtons(bot *tgbotapi.BotAPI, chatID int64, messageID int, locations []structs.UsersLocationBookmark, mapLocs map[string]structs.SiteLocation) {

	buttonRows := make([][]tgbotapi.InlineKeyboardButton, len(locations))
	for i, e := range locations {

		// assemble address (label) of a location
		currentLoc := mapLocs[e.LocationID]
		var buffer bytes.Buffer
		buffer.WriteRune('📍')
		if len(currentLoc.NationalPark) > 0 {
			buffer.WriteString(currentLoc.NationalPark)
			buffer.WriteString(", ")
		}
		buffer.WriteString(currentLoc.Name)
		buffer.WriteString(", ")
		buffer.WriteString(currentLoc.Region)
		buffer.WriteString(", UK")

		// add button to the row
		buttonRows[i] = []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(buffer.String(), ButtonLocationPrefix+Separator+e.LocationID),
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
	keyboardMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, keyboard)
	bot.Send(keyboardMsg)
}

// renders the button row with days for detailed forecast
func renderDetailedDatesButtons(bot *tgbotapi.BotAPI, chatID int64, messageID int, root *structs.RootSiteRep) {

	rowDaysButtons := make([]tgbotapi.InlineKeyboardButton, 5)
	strMessageID := strconv.Itoa(messageID)

	for i, period := range root.SiteRep.Dv.Location.Periods {
		text := "NaN"
		if t, err := time.Parse(layoutMetofficeDate, period.Value); err == nil {
			text = t.Format("2/1")
		}

		rowDaysButtons[i] = tgbotapi.NewInlineKeyboardButtonData(text,
			ButtonDaysPrefix+Separator+root.SiteRep.Dv.Location.ID+Separator+period.Value+Separator+strMessageID)
	}

	rowCloseButton := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Close", ButtonDeleteMsgPrefix+Separator+strMessageID),
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rowDaysButtons, rowCloseButton)
	keyboardMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, keyboard)
	bot.Send(keyboardMsg)
}
