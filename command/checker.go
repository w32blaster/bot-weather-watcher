package command

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/w32blaster/bot-weather-watcher/structs"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/msgpack"
	"github.com/sirupsen/logrus"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func CheckWeather(bot *tgbotapi.BotAPI, opts *structs.Opts) {

	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	if err != nil {
		log.WithError(err).Warn("Can't open database")
		return
	}
	defer db.Close()

	var locations []structs.UsersLocationBookmark
	db.All(&locations)

	var buffer bytes.Buffer
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

			if feelsLikeDayTemp > loc.LowestTemp && windNoon < loc.MaxWindSpeed && precProbab < 40 {
				buffer.WriteString(
					fmt.Sprintf(" - in %s at %s (day temp %d˚C, wind is %d mpg and precipitation probability is %d) \n",
						forecast.SiteRep.Dv.Location.Name, day.Value, feelsLikeDayTemp, windNoon, precProbab),
				)
			}
		}

		if buffer.Len() > 0 {
			sendMsg(bot, loc.ChatID, "Hey, good weather will be at: \n\n"+buffer.String())
		}

		buffer.Reset()
	}
}
