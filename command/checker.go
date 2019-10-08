package command

import (
	"fmt"
	"log"

	"github.com/w32blaster/bot-weather-watcher/structs"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/msgpack"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func CheckWeather(bot *tgbotapi.BotAPI) {

	db, err := storm.Open(DbPath, storm.Codec(msgpack.Codec))
	if err != nil {
		log.Println("Error! " + err.Error())
		return
	}
	defer db.Close()

	var locations []structs.UsersLocationBookmark
	db.All(&locations)

	//for i, loc := range locations {
	//
	//}

	fmt.Printf("CHEDULEEEER!!!!")
}
