package mmdb

import (
	"../config"
	"../updater"
	"github.com/oschwald/maxminddb-golang"
)

// DB is the IP geolocation database.
type DB struct {
	reader      		*maxminddb.Reader 	// Actual db object.

	Updater 			*updater.Config			// Holds all notification channels
	Config 				*config.Config		// Shared default configuration

	ErrUnavailable		error				// ErrUnavailable may be returned by DB.Lookup when the database
	// points to a URL and is not yet available because it's being
	// downloaded in background.
}

// DefaultQuery is the default query used for database lookups.
// https://dev.maxmind.com/geoip/geoip2/geoip2-city-country-csv-databases/
type DefaultQuery struct {
	Continent struct {
		Code string `maxminddb:"code"`
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
		PopulationDensity  uint  `maxminddb:"population_density"`
	} `maxminddb:"location"`
	Traits struct{
		AutonomousSystemNumber uint `maxminddb:"autonomous_system_number"`
		AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
		Domain string `maxminddb:"domain"`
		IsAnonymous bool `maxminddb:"is_anonymous"`
		IsAnonymousProxy bool `maxminddb:"is_anonymous_proxy"`
		IsAnonymousVPN bool `maxminddb:"is_anonymous_vpn"`
		IsHostingProvider bool `maxminddb:"is_hosting_provider"`
		IsPublicProxy bool `maxminddb:"is_public_proxy"`
		IsSatelliteProvider bool `maxminddb:"is_satellite_provider"`
		IsTorExitNode bool `maxminddb:"is_tor_exit_node"`
		ISP string `maxminddb:"isp"`
		IpAddress string `maxminddb:"ip_address"`
		Network string `maxminddb:"network"`
		Organization string `maxminddb:"organization"`
		StaticIpScore string `maxminddb:"static_ip_score"`
		UserCount string `maxminddb:"user_count"`
		UserType string `maxminddb:"user_type"`
	} `maxminddb:"traits"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
}

type ASNDefaultQuery struct {
	AutonomousSystemNumber uint `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
}

type TorDefaultQuery struct {
	IsTorUser bool
}
