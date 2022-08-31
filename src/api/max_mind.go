package api

import (
	"encoding/csv"
	"fmt"
	"github.com/oschwald/maxminddb-golang"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MaxMind struct {
	UserID         string        `json:"user_id"`
	LicenseKey     string        `json:"license_key"`
	ProductID      string        `json:"product_id"`
	Downstreams    string        `json:"downstreams"`
	RetryInterval  time.Duration `json:"retry_interval"`
	UpdateInterval time.Duration `json:"update_interval"`

	updater    *Updater
	asnUpdater *Updater
	csvUpdater *Updater

	reader    *maxminddb.Reader // Actual db object.
	asnReader *maxminddb.Reader // Actual db object.

	mx        sync.RWMutex
	importing int
}

type MaxMindCsvRecord struct {
	Network                     string
	GeonameId                   int
	RegisteredCountryGeonameId  int
	RepresentedCountryGeonameId int
	IsAnonymousProxy            int
	IsSatelliteProvider         int
	PostalCode                  string
	Latitude                    float64
	Longitude                   float64
	AccuracyRadius              int
}

// MaxMindDefaultQuery is the default query used for database lookups.
// https://dev.maxmind.com/geoip/geoip2/geoip2-city-country-csv-databases/
type MaxMindDefaultQuery struct {
	Continent struct {
		Code  string            `maxminddb:"code"`
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"continent"`
	Country struct {
		ISOCode           string            `maxminddb:"iso_code"`
		ContinentCode     string            `maxminddb:"continent_code"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union"`
		Names             map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	Region []struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Location struct {
		Latitude          float64 `maxminddb:"latitude"`
		Longitude         float64 `maxminddb:"longitude"`
		AccuracyRadius    uint    `maxminddb:"accuracy_radius"`
		MetroCode         uint    `maxminddb:"metro_code"`
		TimeZone          string  `maxminddb:"time_zone"`
		PopulationDensity uint    `maxminddb:"population_density"`
	} `maxminddb:"location"`
	Traits struct {
		AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
		AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
		Domain                       string `maxminddb:"domain"`
		IsAnonymous                  bool   `maxminddb:"is_anonymous"`
		IsAnonymousProxy             bool   `maxminddb:"is_anonymous_proxy"`
		IsAnonymousVPN               bool   `maxminddb:"is_anonymous_vpn"`
		IsHostingProvider            bool   `maxminddb:"is_hosting_provider"`
		IsPublicProxy                bool   `maxminddb:"is_public_proxy"`
		IsSatelliteProvider          bool   `maxminddb:"is_satellite_provider"`
		IsTorExitNode                bool   `maxminddb:"is_tor_exit_node"`
		ISP                          string `maxminddb:"isp"`
		IpAddress                    string `maxminddb:"ip_address"`
		Network                      string `maxminddb:"network"`
		Organization                 string `maxminddb:"organization"`
		StaticIpScore                string `maxminddb:"static_ip_score"`
		UserCount                    string `maxminddb:"user_count"`
		UserType                     string `maxminddb:"user_type"`
	} `maxminddb:"traits"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
}

type MaxMindASNDefaultQuery struct {
	AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
}

func (mm *MaxMind) Start(rootDir string, updateCallback func(resp *MaxMindCsvRecord)) {
	mm.updater = &Updater{
		file:          filepath.Join(rootDir, "cache", mm.ProductID+"-City.mmdb"),
		archive:       filepath.Join(rootDir, "cache", mm.ProductID+"-City.tar.gz"),
		interval:      mm.UpdateInterval,
		retryInterval: mm.RetryInterval,
		updateUrl:     mm.UpdateURL(mm.ProductID+"-City", "tar.gz"),
		cbk: func() error {
			mm.mx.Lock()
			defer mm.mx.Unlock()
			if mm.reader != nil {
				_ = mm.reader.Close()
			}

			return mm.updateCallback(mm.updater, func(reader *maxminddb.Reader) {
				mm.reader = reader
			})()
		},
	}
	mm.asnUpdater = &Updater{
		file:          filepath.Join(rootDir, "cache", mm.ProductID+"-ASN.mmdb"),
		archive:       filepath.Join(rootDir, "cache", mm.ProductID+"-ASN.tar.gz"),
		interval:      mm.UpdateInterval,
		retryInterval: mm.RetryInterval,
		updateUrl:     mm.UpdateURL(mm.ProductID+"-ASN", "tar.gz"),
		cbk: func() error {
			mm.mx.Lock()
			defer mm.mx.Unlock()
			if mm.asnReader != nil {
				_ = mm.asnReader.Close()
			}

			return mm.updateCallback(mm.asnUpdater, func(reader *maxminddb.Reader) {
				mm.asnReader = reader
			})()
		},
	}

	mm.csvUpdater = &Updater{
		file:          filepath.Join(rootDir, "cache", mm.ProductID+"-City-CSV.zip"),
		archive:       filepath.Join(rootDir, "cache", mm.ProductID+"-City-CSV.zip"),
		interval:      mm.UpdateInterval,
		retryInterval: mm.RetryInterval,
		updateUrl:     mm.UpdateURL(mm.ProductID+"-City-CSV", "zip"),
		cbk: func() error {
			if mm.IsImporting() {
				return nil
			}
			mm.SetImporting(2)
			fmt.Printf("Unpacking: %s\n", mm.csvUpdater.file)
			if err := mm.csvUpdater.ExtractAllFromZip(); err != nil {
				mm.SetImporting(0)
				return err
			}

			dbdir := strings.TrimSuffix(mm.csvUpdater.archive, filepath.Ext(mm.csvUpdater.archive))

			go func() {
				filename := path.Join(dbdir, mm.ProductID+"-City-Blocks-IPv6.csv")
				if err := mm.importRecords(filename, updateCallback); err != nil {
					fmt.Printf("[error] %s\n", err.Error())
				}
				mm.SubImport()
			}()

			go func() {
				filename := path.Join(dbdir, mm.ProductID+"-City-Blocks-IPv4.csv")
				if err := mm.importRecords(filename, updateCallback); err != nil {
					fmt.Printf("[error] %s\n", err.Error())
				}
				mm.SubImport()
			}()

			return nil
		},
	}

	go mm.updater.Start()
	go mm.asnUpdater.Start()
	go mm.csvUpdater.Start()
}

func (mm *MaxMind) Stop() {
	mm.updater.Stop()
	mm.asnUpdater.Stop()
	mm.csvUpdater.Stop()
}

func (mm *MaxMind) UpdateURL(productID, suffix string) string {
	return fmt.Sprintf("https://%s/app/geoip_download?edition_id=%s&date=&license_key=%s&suffix=%s", mm.Downstreams, productID, mm.LicenseKey, suffix)
}

func (mm *MaxMind) importRecords(filename string, updateCallback func(d *MaxMindCsvRecord)) error {
	fmt.Printf("Loading: %s\n", filename)
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)

	for {
		data, err := csvReader.Read()
		if err != nil {
			return err
		}
		if data == nil {
			break
		}

		if len(data) == 10 {
			resp := MaxMindResponseFromArray(data)
			if resp.Network == "network" {
				continue
			}

			updateCallback(resp)
		}
	}
	return nil
}

func (mm *MaxMind) updateCallback(u *Updater, readerSetter func(reader *maxminddb.Reader)) func() error {
	return func() error {
		fmt.Printf("Loading: %s\n", u.file)
		err, dbFileName := u.ProcessFile()
		if err != nil {
			return err
		}

		f, err := os.Open(dbFileName)
		if err != nil {
			return err
		}

		if _, err := ioutil.ReadAll(f); err != nil {
			return err
		}
		reader, err := maxminddb.Open(dbFileName)
		if err != nil {
			return err
		}
		stat, err := os.Stat(u.archive)
		if err != nil {
			return err
		}
		u.lastUpdated = stat.ModTime()
		readerSetter(reader)
		return nil
	}
}

// Lookup performs a database lookup of the given IP address, and stores
// the response into the result value. The result value must be a struct
// with specific fields and tags as described here:
// https://godoc.org/github.com/oschwald/maxminddb-golang#Reader.Lookup
//
// See the DefaultQuery for an example of the result struct.
func (mm *MaxMind) Lookup(addr net.IP) *MaxMindDefaultQuery {
	mm.mx.RLock()
	defer mm.mx.RUnlock()

	result := &MaxMindDefaultQuery{}
	if mm.reader != nil {
		if err := mm.reader.Lookup(addr, result); err != nil {
			fmt.Printf("[error] %s\n", err.Error())
		}
	}
	return result
}

func (mm *MaxMind) LookupASN(addr net.IP) *MaxMindASNDefaultQuery {
	mm.mx.RLock()
	defer mm.mx.RUnlock()

	result := &MaxMindASNDefaultQuery{}
	if mm.reader != nil {
		if err := mm.asnReader.Lookup(addr, result); err != nil {
			fmt.Printf("[error] %s\n", err.Error())
		}
	}
	return result
}

func MaxMindResponseFromArray(data []string) *MaxMindCsvRecord {
	response := &MaxMindCsvRecord{}

	response.Network = data[0]
	response.GeonameId, _ = strconv.Atoi(data[1])
	response.RegisteredCountryGeonameId, _ = strconv.Atoi(data[2])
	response.RepresentedCountryGeonameId, _ = strconv.Atoi(data[3])
	response.IsAnonymousProxy, _ = strconv.Atoi(data[4])
	response.IsSatelliteProvider, _ = strconv.Atoi(data[5])
	response.PostalCode = data[6]
	response.Latitude, _ = strconv.ParseFloat(data[7], 64)
	response.Longitude, _ = strconv.ParseFloat(data[8], 64)
	response.AccuracyRadius, _ = strconv.Atoi(data[9])

	return response
}

func (mm *MaxMind) Ready() bool {
	mm.mx.RLock()
	defer mm.mx.RUnlock()

	return mm.reader != nil && mm.asnReader != nil
}

func (mm *MaxMind) IsImporting() bool {
	mm.mx.RLock()
	defer mm.mx.RUnlock()

	return mm.importing > 0
}

func (mm *MaxMind) AddImport() {
	mm.mx.Lock()
	defer mm.mx.Unlock()

	mm.importing++
}

func (mm *MaxMind) SubImport() {
	mm.mx.Lock()
	defer mm.mx.Unlock()

	mm.importing--
}

func (mm *MaxMind) SetImporting(state int) {
	mm.mx.Lock()
	defer mm.mx.Unlock()

	mm.importing = state
}
