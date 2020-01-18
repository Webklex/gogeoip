package server

import (
	"../utils/config"
	"../utils/i2ldb"
	"../utils/mmdb"
	"../utils/tor"
	"encoding/xml"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)


// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.

type Server struct {
	Config *config.Config

	Host string
	Port int

	RateLimit *RateLimit
	Visitors map[string]*Visitor
	Mutex sync.Mutex

	Api *ApiHandler
}

type GeoIpQuery struct {
	i2ldb.ProxyDefaultQuery
	mmdb.DefaultQuery
	mmdb.ASNDefaultQuery
	mmdb.TorDefaultQuery
}

type responseRecord struct {
	XMLName     		xml.Name	`xml:"Response" json:"-"`
	IP          		string  	`json:"ip"`
	Isp 				string		`json:"isp"`
	Domain 				string		`json:"domain"`
	IsInEuropeanUnion 	bool  		`json:"is_in_european_union"`
	ContinentCode 		string  	`json:"continent_code"`
	CountryCode 		string  	`json:"country_code"`
	CountryName 		string  	`json:"country_name"`
	RegionCode  		string  	`json:"region_code"`
	RegionName  		string  	`json:"region_name"`
	City        		string  	`json:"city"`
	ZipCode     		string  	`json:"zip_code"`
	TimeZone    		string  	`json:"time_zone"`
	Latitude    		float64 	`json:"latitude"`
	Longitude   		float64 	`json:"longitude"`
	PopulationDensity   uint     	`json:"population_density,omitempty"`
	AccuracyRadius   	uint  		`json:"accuracy_radius"`
	MetroCode   		uint    	`json:"metro_code"`
	ASN					*ASNRecord  `json:"asn,omitempty"`
	User				*UserRecord `json:"user,omitempty"`
}

type ASNRecord struct {
	AutonomousSystemNumber 			uint   `json:"asn"`
	AutonomousSystemOrganization 	string `json:"aso"`
}

type UserRecord struct {
	Language	*LanguageRecord  `json:"language"`
	System		*SystemRecord    `json:"system"`
}

type LanguageRecord struct {
	Language string 	`json:"language"`
	Region   string 	`json:"region"`
	Tag      string 	`json:"tag"`
}

type SystemRecord struct {
	OS 			string 	`json:"os"`
	Browser   	string 	`json:"browser"`
	Version   	string 	`json:"version"`
	OSVersion   string 	`json:"os_version"`
	Device   	string 	`json:"device"`
	Mobile   	bool 	`json:"mobile"`
	Tablet   	bool 	`json:"tablet"`
	Desktop   	bool 	`json:"desktop"`
	Bot   		bool 	`json:"bot"`
	Tor   		bool 	`json:"tor"`
	LastSeen 	uint	`json:"last_seen"`
	UsageType 	string	`json:"usage_type"`
	ProxyType 	string	`json:"proxy_type"`
	Proxy 		bool	`json:"proxy"`
}

type ApiHandler struct {
	db    *mmdb.DB
	asnDB *mmdb.DB
	torDB *tor.Config
	i2lDB *i2ldb.Config
	cors  *cors.Cors
}

func NewServerConfig(c *config.Config) *Server {
	parts := strings.Split(c.ServerAddr, ":")
	host := parts[0]
	port, err := strconv.Atoi(parts[1])

	if err != nil || port <= 0 {
		print("Invalid Socket provided")
		os.Exit(1)
	}

	conf := &Server{
		Config: c,

		Host: host,
		Port: port,

		RateLimit: NewRateLimit(rate.Limit(c.RateLimitLimit), c.RateLimitBurst),

		Api: &ApiHandler{
			db:    mmdb.NewDefaultConfig(c, c.MMProductID),
			asnDB: mmdb.NewDefaultConfig(c, c.MMASNProductID),
			torDB: tor.NewDefaultConfig(c),
			i2lDB: i2ldb.NewDefaultConfig(c),
		},
	}

	return conf
}

func (s *Server) Start() {
	go s.watchEvents(s.Api.torDB.Updater)
	go s.watchEvents(s.Api.db.Updater)
	go s.watchEvents(s.Api.asnDB.Updater)
	go s.watchEvents(s.Api.i2lDB.Updater)

	_ = s.openDB()

	if s.Config.LogToStdout {
		log.SetOutput(os.Stdout)
	}
	if !s.Config.LogTimestamp {
		log.SetFlags(0)
	}
	f, err := s.NewHandler()
	if err != nil {
		log.Fatal(err)
	}
	if s.Config.ServerAddr != "" {
		go s.runServer(f)
	}
	if s.Config.TLSServerAddr != "" {
		go s.runTLSServer(f)
	}
	select {}
}