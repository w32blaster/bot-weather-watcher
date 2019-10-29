package command

import (
	"bytes"
	"fmt"
	"github.com/asdine/storm/q"
	"github.com/getsentry/sentry-go"
	"strconv"
	"strings"
	"time"

	"github.com/w32blaster/bot-weather-watcher/structs"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/msgpack"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const precipProbRain = 40 // min precipitation probability when we assume that will be rainy day

func CheckWeather(bot *tgbotapi.BotAPI, opts *structs.Opts, userID int) bool {

	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	if err != nil {
		sentry.CaptureException(err)
		return false
	}
	defer db.Close()

	locations, ok := getBookmarksFromDatabase(db, userID)
	if !ok {
		return false
	}

	// load locations, build a map
	mapLocs := getMapOfLocations(locations, db)

	var buffer bytes.Buffer
	wasFoundSomething := false
	for _, loc := range locations {

		sentry.CurrentHub().PushScope()
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       strconv.Itoa(loc.UserID),
				Username: loc.UserName,
			})
			scope.SetTag("action", "check-weather-nightly")
			scope.SetExtra("location-id", loc.LocationID)
			scope.SetExtra("location-name", mapLocs[loc.LocationID].Name)
			scope.SetExtra("location-auth-area", mapLocs[loc.LocationID].AuthArea)
			scope.RemoveExtra("raw-text")
		})

		forecast, err := getDailyForecastFor(loc.LocationID, opts)
		if err != nil {
			sentry.CaptureException(err)
			continue
		}

		// iterate over days
		for _, day := range forecast.SiteRep.Dv.Location.Periods {

			// parse date
			t, err := time.Parse(layoutMetofficeDate, day.Value)
			if err != nil {
				sentry.CaptureException(errors.Wrap(err, "Can't parse date from the bookmark, this day is ignored from checking"))
				continue
			}

			if !shouldBotherForWeekdays(loc.CheckPeriod, t.Weekday()) {
				continue
			}

			feelsLikeDayTemp, windNoon, precProbab, weatherType := parseNumberFigures(day)

			// decide whether current weather is that "good" or "naaah"
			isSuitableWeather := feelsLikeDayTemp > loc.LowestTemp &&
				windNoon < loc.MaxWindSpeed &&
				precProbab < precipProbRain

			logEventToSentry(loc, day, forecast, feelsLikeDayTemp, windNoon, precProbab, isSuitableWeather)

			if isSuitableWeather {
				buffer.WriteString(
					fmt.Sprintf(" - %c in %s at %s (day temp %d˚C, wind is %dmph and precipitation probability is %d%%) \n",
						mapWeatherTypes[weatherType].icon,
						strings.Title(strings.ToLower(forecast.SiteRep.Dv.Location.Name)),
						t.Format("02 Jan 2006, Mon"),
						feelsLikeDayTemp,
						windNoon,
						precProbab),
				)
			}

			// just to avoid any dDOS filters block us :) we are not in a hurry
			time.Sleep(1 * time.Second)
		}

		if buffer.Len() > 0 {
			msg, _ := sendMsg(bot, loc.ChatID, "Hey, good weather will be at: \n\n"+buffer.String())
			renderButtonDeleteBookmark(bot, loc.ChatID, msg.MessageID, loc.ID, mapLocs[loc.LocationID].Name)
			wasFoundSomething = true
		}

		buffer.Reset()
		sentry.CurrentHub().PopScope()
	}

	return wasFoundSomething
}

func logEventToSentry(loc structs.UsersLocationBookmark, day structs.Period, forecast *structs.RootSiteRep, feelsLikeDayTemp, windNoon, precProbab int, isSuitableWeather bool) {
	event := sentry.NewEvent()
	event.Message = "Checker was called the forecast"
	event.Timestamp = time.Now().UTC().Unix()
	event.Tags = map[string]string{
		"action": "check-weather-nightly",
	}
	event.User = sentry.User{
		ID: strconv.Itoa(loc.UserID),
	}
	event.Level = sentry.LevelInfo
	event.Extra = map[string]interface{}{
		"date":                   day.Value,
		"bookmark-owner":         loc.UserID,
		"bookmark-location":      forecast.SiteRep.Dv.Location.Name,
		"temp-feels-like":        feelsLikeDayTemp,
		"temp-min-desired":       loc.LowestTemp,
		"wind-speed":             windNoon,
		"wind-speed-max-desired": loc.MaxWindSpeed,
		"precip-prob":            precProbab,
		"is-suitable":            isSuitableWeather,
	}
	sentry.CaptureEvent(event)
}

func parseNumberFigures(day structs.Period) (int, int, int, int) {

	var feelsLikeDayTemp, windNoon, precProbab int
	weatherType := 4 // default value is "not used"

	if intFrm, err := strconv.Atoi(day.Rep[0]["FDm"]); err == nil {
		feelsLikeDayTemp = intFrm
	}
	if intGn, err := strconv.Atoi(day.Rep[0]["Gn"]); err == nil {
		windNoon = intGn
	}
	if intPpd, err := strconv.Atoi(day.Rep[0]["PPd"]); err == nil {
		precProbab = intPpd
	}
	if intWt, err := strconv.Atoi(day.Rep[0]["W"]); err == nil {
		weatherType = intWt
	}
	return feelsLikeDayTemp, windNoon, precProbab, weatherType
}

func getBookmarksFromDatabase(db *storm.DB, userID int) ([]structs.UsersLocationBookmark, bool) {

	var locations []structs.UsersLocationBookmark
	var err error
	if userID == -1 {

		// find all ready bookmarks
		err = db.Find("IsReady", true, &locations)
	} else {

		// find all the bookrmarks for the given user
		err = db.Select(q.And(
			q.Eq("UserID", userID),
			q.Eq("IsReady", true),
		)).Find(&locations)
	}

	if err != nil {
		sentry.CaptureException(err)
		return nil, false
	}

	return locations, true
}

// shortcut function, checks should we bother a customer in a specific day depending on his/her
// preferences
// Please refer to unit tests
func shouldBotherForWeekdays(dayChoice int, weekday time.Weekday) bool {
	return !(dayChoice == onlyWeekends && weekday != time.Friday && weekday != time.Saturday && weekday != time.Sunday)
}

// renders the button row with days for detailed forecast
func renderButtonDeleteBookmark(bot *tgbotapi.BotAPI, chatID int64, messageID int, bookmarkID int, locationName string) {

	rowCloseButton := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ Stop observing "+locationName, ButtonDeleteBookmark+Separator+strconv.Itoa(bookmarkID)),
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rowCloseButton)
	keyboardMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, keyboard)
	bot.Send(keyboardMsg)
}
