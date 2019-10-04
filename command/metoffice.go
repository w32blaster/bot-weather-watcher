package command

import (
	"encoding/json"
	"github.com/w32blaster/bot-weather-watcher/structs"
	"net/http"
)

func getDailyForecastFor(locationID string, opts *structs.Opts) (*structs.RootSiteRep, error) {

	resp, err := http.Get("http://datapoint.metoffice.gov.uk/public/data/val/wxfcs/all/json/" + locationID + "?res=daily&key=" + opts.MetofficeAppID)
	if err != nil {
		return nil, err
	}

	var result structs.RootSiteRep
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func get3HoursForecastFor(locationID string, opts *structs.Opts) (*structs.RootSiteRep, error) {

	resp, err := http.Get("http://datapoint.metoffice.gov.uk/public/data/val/wxfcs/all/json/" + locationID + "?res=3hourly&key=" + opts.MetofficeAppID)
	if err != nil {
		return nil, err
	}

	var result structs.RootSiteRep
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
