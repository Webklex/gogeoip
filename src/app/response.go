package app

import (
	"github.com/pariz/gountries"
	"github.com/webklex/gogeoip/src/models"
)

type CountryResponse struct {
	Name struct {
		gountries.BaseLang
		Native map[string]gountries.BaseLang `json:"native"`
	} `json:"name"`

	EuMember    bool   `json:"eu_member"`
	LandLocked  bool   `json:"land_locked"`
	Nationality string `json:"nationality"`

	TLDs []string `json:"tlds"`

	Languages    map[string]string             `json:"languages"`
	Translations map[string]gountries.BaseLang `json:"translations"`
	Currencies   []string                      `json:"currency"`
	Borders      []string                      `json:"borders"`

	// Grouped meta
	Codes
	Geo
	Coordinates
}

// Geo contains geographical information
type Geo struct {
	Region    string  `json:"region"`
	SubRegion string  `json:"subregion"`
	Continent string  `json:"continent"`
	Capital   string  `json:"capital"`
	Area      float64 `json:"area"`
}

// Coordinates contains the coordinates for both Country and SubDivision
type Coordinates struct {
	MinLongitude float64 `json:"min_longitude"`
	MinLatitude  float64 `json:"min_latitude"`
	MaxLongitude float64 `json:"max_longitude"`
	MaxLatitude  float64 `json:"max_latitude"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

// Codes contains various code representations
type Codes struct {
	Alpha2              string   `json:"cca2"`
	Alpha3              string   `json:"cca3"`
	CIOC                string   `json:"cioc"`
	CCN3                string   `json:"ccn3"`
	CallingCodes        []string `json:"calling_codes"`
	InternationalPrefix string   `json:"international_prefix"`
}

type UserAgentResponse struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	OS        string `json:"os"`
	OSVersion string `json:"os_version"`
	Device    string `json:"device"`
	Mobile    bool   `json:"mobile"`
	Tablet    bool   `json:"tablet"`
	Desktop   bool   `json:"desktop"`
	Bot       bool   `json:"bot"`
}

type LanguageResponse struct {
	Language string `json:"language"`
	Region   string `json:"region"`
	Tag      string `json:"tag"`
}

type MeResponse struct {
	Ip        *models.IP         `json:"ip"`
	Language  *LanguageResponse  `json:"language"`
	UserAgent *UserAgentResponse `json:"user_agent"`
}

type StatisticResponse struct {
	Ips       int64 `json:"ips"`
	Cities    int64 `json:"cities"`
	Countries int64 `json:"countries"`
	Domains   int64 `json:"domains"`
	Isps      int64 `json:"isps"`
	Asns      int64 `json:"asns"`
	Networks  int64 `json:"networks"`
}
