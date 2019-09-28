package structs

import "time"

type (
	Root struct {
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
)
