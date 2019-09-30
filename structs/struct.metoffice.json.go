package structs

import "time"

type (
	RootSiteRep struct {
		SiteRep SiteRep `json:"SiteRep"`
	}

	SiteRep struct {
		Wx Wx `json:"Wx"`
		Dv Dv `json:"DV"`
	}

	Wx struct {
		Params []WxParam `json:"Param"`
	}

	WxParam struct {
		Name    string `json:"name"`
		Units   string `json:"units"`
		Comment string `json:"$"`
	}

	Dv struct {
		Data     time.Time `json:"dataDate"`
		Type     string    `json:"type"`
		Location Location  `json:"Location"`
	}

	Location struct {
		ID        string   `json:"i"`
		Latitude  string   `json:"lat"`
		Longitude string   `json:"lon"`
		Elevation string   `json:"elevation"`
		Name      string   `json:"name"`
		Country   string   `json:"country"`
		Continent string   `json:"continent"`
		Periods   []Period `json:"Period"`
	}

	Period struct {
		Type  string              `json:"type"`
		Value string              `json:"value"`
		Rep   []map[string]string `json:"Rep"`
	}

	RootLocations struct {
		Locations SiteLocations `json:"Locations"`
	}

	SiteLocations struct {
		Location []SiteLocation `json:"Location"`
	}

	SiteLocation struct {
		ID           string `json:"id" storm:"unique"`
		Elevation    string `json:"elevation"`
		Latitude     string `json:"latitude"`
		Longitude    string `json:"longitude"`
		Name         string `json:"name" storm:"index"`
		Region       string `json:"region"`
		AuthArea     string `json:"unitaryAuthArea" storm:"index"`
		NationalPark string `json:"nationalPark" storm:"index"`
		ObsSource    string `json:"obsSource"`
	}
)
