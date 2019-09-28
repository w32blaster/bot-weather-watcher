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
		params []WxParam `json:"Param"`
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
		Latitude  float32  `json:"lat"`
		Longitude float32  `json:"lon"`
		Elevation float32  `json:"elevation"`
		Name      string   `json:"name"`
		Country   string   `json:"country"`
		Continent string   `json:"continent"`
		Periods   []Period `json:"Period"`
	}

	Period struct {
		Type  string            `json:"type"`
		Value string            `json:"value"`
		Rep   map[string]string `json:"Rep"`
	}

	RootLocations struct {
		Location []SiteLocation `json:"Location"`
	}

	SiteLocation struct {
		ID           string  `json:"id"`
		Elevation    float32 `json:"elevation"`
		Latitude     float32 `json:"latitude"`
		Longitude    float32 `json:"longitude"`
		Name         string  `json:"name"`
		Region       string  `json:"region"`
		AuthAread    string  `json:"unitaryAuthArea"`
		NationalPart string  `json:"nationalPark"`
		ObsSource    string  `json:"obsSource"`
	}
)
