package command

import (
	"bytes"
	"fmt"
	"github.com/asdine/storm/q"
	"strconv"
	"time"

	"github.com/w32blaster/bot-weather-watcher/structs"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/msgpack"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

func CheckWeather(bot *tgbotapi.BotAPI, opts *structs.Opts, userID int) bool {

	locations, ok := getBookmarksFromDatabase(userID)
	if !ok {
		return false
	}

	var buffer bytes.Buffer
	wasFoundSomething := false
	for _, loc := range locations {

		forecast, err := getDailyForecastFor(loc.LocationID, opts)
		if err != nil {
			log.WithError(err).WithFields(logrus.Fields{
				"location-id": loc.LocationID,
				"id":          loc.ID,
			}).Error("Can't get daily forecast from metoffice")

			continue
		}

		// iterate over days
		for _, day := range forecast.SiteRep.Dv.Location.Periods {

			// parse date
			t, err := time.Parse(layoutMetofficeDate, day.Value)
			if err != nil {
				log.WithError(err).Error("Can't parse date from the bookmark, this day is ignored from checking")
				continue
			}

			if !shouldBotherForWeekdays(loc.CheckPeriod, t.Weekday()) {
				continue
			}

			// REFINE THIS LOGIC!
			var feelsLikeDayTemp, windNoon, precProbab int

			if intFrm, err := strconv.Atoi(day.Rep[0]["FDm"]); err == nil {
				feelsLikeDayTemp = intFrm
			}
			if intGn, err := strconv.Atoi(day.Rep[0]["Gn"]); err == nil {
				windNoon = intGn
			}
			if intPpd, err := strconv.Atoi(day.Rep[0]["PPd"]); err == nil {
				precProbab = intPpd
			}

			isSuitableWeather := feelsLikeDayTemp > loc.LowestTemp && windNoon < loc.MaxWindSpeed && precProbab < 40

			log.WithFields(logrus.Fields{
				"date":                   day.Value,
				"for-user":               loc.UserID,
				"location":               forecast.SiteRep.Dv.Location.Name,
				"temp-feels-like":        feelsLikeDayTemp,
				"temp-min-desired":       loc.LowestTemp,
				"wind-speed":             windNoon,
				"wind-speed-max-desired": loc.MaxWindSpeed,
				"precip-prob":            precProbab,
				"is-suitable":            isSuitableWeather,
			}).Info("Checker was called the forecast")

			if isSuitableWeather {
				buffer.WriteString(
					fmt.Sprintf(" - in %s at %s (day temp %d˚C, wind is %d mpg and precipitation probability is %d%%) \n",
						forecast.SiteRep.Dv.Location.Name, day.Value, feelsLikeDayTemp, windNoon, precProbab),
				)
			}
		}

		if buffer.Len() > 0 {
			sendMsg(bot, loc.ChatID, "Hey, good weather will be at: \n\n"+buffer.String())
			wasFoundSomething = true
		}

		buffer.Reset()
	}

	return wasFoundSomething
}

func getBookmarksFromDatabase(userID int) ([]structs.UsersLocationBookmark, bool) {

	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	if err != nil {
		log.WithError(err).Warn("Can't open database")
		return nil, false
	}
	defer db.Close()

	var locations []structs.UsersLocationBookmark
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
		log.WithError(err).Error("Can't read bookmarks form DB")
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
