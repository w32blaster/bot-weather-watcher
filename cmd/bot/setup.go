package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/w32blaster/bot-weather-watcher/command"
	"github.com/w32blaster/bot-weather-watcher/structs"

	"github.com/asdine/storm"
)

func main() {
	fmt.Println("Populate database with site locations")
	db, err := storm.Open(command.DbPath, storm.Batch())
	if err != nil {
		fmt.Println("Error opening the database, err " + err.Error())
		os.Exit(1)
	}

	// parse the list of locations
	locations, err := getListOfLocations()
	if err != nil {
		fmt.Println("Error getting of a list of locations, err: " + err.Error())
		os.Exit(1)
	}

	fmt.Printf("Found %d items. Let's insert them to a database\n", len(locations))
	tx, err := db.Begin(true)
	for i, loc := range locations {
		tx.Save(&loc)
		if i%500 == 0 {
			fmt.Printf("%d) records inserted\n", i)
		}
	}
	if err := tx.Commit(); err != nil {
		fmt.Printf("Error! Can't commit transaction. " + err.Error())
	}
	fmt.Println("All done")

	var locs []structs.SiteLocation
	db.All(&locs)

	fmt.Printf("We have %d records\n", len(locs))
}

func getListOfLocations() ([]structs.SiteLocation, error) {
	confFile, err := os.Open("api-examples/site-list.json")
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(confFile)
	if err != nil {
		return nil, err
	}

	var result structs.RootLocations
	if err = json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	return result.Locations.Location, nil
}
