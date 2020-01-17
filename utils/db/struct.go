package db

import (
	"../config"
	"github.com/oschwald/maxminddb-golang"
	"sync"
	"time"
)

// DB is the IP geolocation database.
type DB struct {
	File        		string            	// Database file name.
	reader      		*maxminddb.Reader 	// Actual db object.
	closed      		bool              	// Mark this db as closed.
	lastUpdated 		time.Time         	// Last time the db was updated.
	mu          		sync.RWMutex      	// Protects all the above.

	Notifier 			*Notifier			// Holds all notification channels
	Config 				*config.Config		// Shared default configuration

	updateUrl			string				// MaxMind update url
	dbArchive			string				// Local cached copy of a database downloaded from a URL.
	ErrUnavailable		error				// ErrUnavailable may be returned by DB.Lookup when the database
	// points to a URL and is not yet available because it's being
	// downloaded in background.
}

// DefaultQuery is the default query used for database lookups.
// https://dev.maxmind.com/geoip/geoip2/geoip2-city-country-csv-databases/
type DefaultQuery struct {
	Continent struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"continent"`
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		ContinentCode string      `maxminddb:"continent_code"`
		IsInEuropeanUnion bool    `maxminddb:"is_in_european_union"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	Region []struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
		AccuracyRadius uint `maxminddb:"accuracy_radius"`
		MetroCode uint    `maxminddb:"metro_code"`
		TimeZone  string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
}
