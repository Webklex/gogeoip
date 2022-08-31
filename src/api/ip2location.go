package api

import (
	"encoding/csv"
	"fmt"
	"github.com/ip2location/ip2proxy-go"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Ip2Location struct {
	Token          string        `json:"token"`
	ProductID      string        `json:"product_id"`
	CsvProductID   string        `json:"csv_product_id"`
	RetryInterval  time.Duration `json:"retry_interval"`
	UpdateInterval time.Duration `json:"update_interval"`
	Downstreams    string        `json:"downstreams"`

	updater    *Updater
	csvUpdater *Updater
	db         *ip2proxy.DB

	mx        sync.RWMutex
	importing int
}

type Ip2LocationResponse struct {
	IsProxy      bool
	ProxyType    string
	CountryShort string
	CountryLong  string
	Region       string
	City         string
	ISP          string
	Domain       string
	UsageType    string
	ASN          string
	AS           string
	LastSeen     int
	Threat       string
}

type Ip2LocationCsvRecord struct {
	IpFrom int // INT (10)† / DECIMAL (39,0)††	First IP address show netblock.
	IpTo   int // INT (10)† / DECIMAL (39,0)††	Last IP address show netblock.

	/*
		VPN	Anonymizing VPN services. These services offer users a publicly accessible VPN for the purpose of hiding their IP address.	High
		TOR	Tor Exit Nodes. The Tor Project is an open network used by those who wish to maintain anonymity.	High
		DCH	Hosting Provider, Data Center or Content Delivery Network. Since hosting providers and data centers can serve to provide anonymity, the Anonymous IP database flags IP addresses associated with them.	Low
		PUB*	Public Proxies. These are services which make connection requests on a user's behalf. Proxy server software can be configured by the administrator to listen on some specified port. These differ from VPNs in that the proxies usually have limited functions compare to VPNs.	High
		WEB	Web Proxies. These are web services which make web requests on a user's behalf. These differ from VPNs or Public Proxies in that they are simple web-based proxies rather than operating at the IP address and other ports level.	High
		SES	Search Engine Robots. These are services which perform crawling or scraping to a website, such as, the search engine spider or bots engine.	Low
		RES	Residential proxies. These services offer users proxy connections through residential ISP with or without consents of peers to share their idle resources.	High
	*/
	ProxyType   string // VARCHAR(3)	Type of proxy
	CountryCode string // CHAR(2)	Two-character country code based on ISO 3166.
	CountryName string // VARCHAR(64)	Country name based on ISO 3166.
	RegionName  string // VARCHAR(128)	Region or state name.
	CityName    string // VARCHAR(128)	City name.
	Isp         string // VARCHAR(256)	Internet Service Provider or company's name.
	Domain      string // VARCHAR(128)	Internet Domain name associated with IP address range.

	// (COM) Commercial
	// (ORG) Organization
	// (GOV) Government
	// (MIL) Military
	// (EDU) University/College/School
	// (LIB) Library
	// (CDN) Content Delivery Network
	// (ISP) Fixed Line ISP
	// (MOB) Mobile ISP
	// (DCH) Data Center/Web Hosting/Transit
	// (SES) Search Engine Spider
	// (RSV) Reserved
	UsageType string // VARCHAR(11)	Usage type classification of ISP or company.

	Asn      int    // INT(10)	Autonomous system number (ASN).
	As       string // VARCHAR(256)	Autonomous system (AS) name.
	LastSeen int    // INT(10)	Proxy last seen in days.
}

func (i2l *Ip2Location) Start(rootDir string, updateCallback func(resp *Ip2LocationCsvRecord)) {
	i2l.updater = &Updater{
		file:          filepath.Join(rootDir, "cache", i2l.ProductID+".bin"),
		archive:       filepath.Join(rootDir, "cache", i2l.ProductID+".zip"),
		interval:      i2l.UpdateInterval,
		retryInterval: i2l.RetryInterval,
		updateUrl:     i2l.UpdateURL(i2l.ProductID),
		cbk:           i2l.NewReader,
	}
	i2l.csvUpdater = &Updater{
		file:          filepath.Join(rootDir, "cache", i2l.CsvProductID+".zip"),
		archive:       filepath.Join(rootDir, "cache", i2l.CsvProductID+".zip"),
		interval:      i2l.UpdateInterval,
		retryInterval: i2l.RetryInterval,
		updateUrl:     i2l.UpdateURL(i2l.CsvProductID),
		cbk: func() error {
			if i2l.IsImporting() {
				return nil
			}
			i2l.SetImporting(1)
			defer i2l.SetImporting(0)

			fmt.Printf("Unpacking: %s\n", i2l.csvUpdater.archive)
			if err := i2l.csvUpdater.ExtractAllFromZip(); err != nil {
				return err
			}

			dbdir := strings.TrimSuffix(i2l.csvUpdater.archive, filepath.Ext(i2l.csvUpdater.archive))
			files, err := ioutil.ReadDir(dbdir)
			if err != nil {
				return err
			} else if len(files) == 0 {
				return fmt.Errorf("no csv files found in %s", dbdir)
			}
			filename := path.Join(dbdir, files[0].Name())

			return i2l.importRecords(filename, updateCallback)
		},
	}
	go i2l.updater.Start()
	go i2l.csvUpdater.Start()
}

func (i2l *Ip2Location) Stop() {
	i2l.updater.Stop()
	i2l.csvUpdater.Stop()
	_ = i2l.db.Close()
}

func (i2l *Ip2Location) UpdateURL(productId string) string {
	return fmt.Sprintf("https://%s/download/?token=%s&file=%s", i2l.Downstreams, i2l.Token, productId)
}

func (i2l *Ip2Location) importRecords(filename string, updateCallback func(resp *Ip2LocationCsvRecord)) error {
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

		if len(data) == 13 {
			resp := Ip2LocationResponseFromArray(data)
			updateCallback(resp)
		}

	}
	return nil
}

func (i2l *Ip2Location) NewReader() error {
	i2l.mx.Lock()
	defer i2l.mx.Unlock()
	fmt.Printf("Loading: %s\n", i2l.updater.file)

	stat, err := os.Stat(i2l.updater.archive)
	if err != nil {
		return err
	}

	if stat.Size() < 1200 {
		return fmt.Errorf("DB File not available")
	}

	err, _ = i2l.updater.ProcessFile()
	if err != nil {
		return err
	}

	_, err = os.Open(i2l.updater.file)
	if err != nil {
		return err
	}

	i2l.db, err = ip2proxy.OpenDB(i2l.updater.file)
	if err != nil {
		return err
	}

	i2l.updater.lastUpdated = stat.ModTime()
	return nil
}

func (i2l *Ip2Location) Lookup(addr net.IP) *Ip2LocationResponse {
	i2l.mx.RLock()
	defer i2l.mx.RUnlock()

	if i2l.db == nil {
		fmt.Printf("[error] Ip2Location not loaded\n")
	}

	response := &Ip2LocationResponse{}
	result, err := i2l.db.GetAll(addr.String())
	if err != nil {
		fmt.Printf("[error] Ip2Location lookup failed: %s\n", err.Error())
		return response
	}

	if i, _ := strconv.Atoi(result["isProxy"]); i > 0 {
		response.IsProxy = true
	}
	response.LastSeen, _ = strconv.Atoi(result["LastSeen"])
	if result["ISP"] != "-" {
		response.ISP = result["ISP"]
	}
	if result["City"] != "-" {
		response.City = result["City"]
	}
	if result["Region"] != "-" {
		response.Region = result["Region"]
	}
	if result["CountryShort"] != "-" {
		response.CountryShort = result["CountryShort"]
	}
	if result["CountryLong"] != "-" {
		response.CountryLong = result["CountryLong"]
	}
	if result["Threat"] != "-" {
		response.Threat = result["Threat"]
	}
	if result["ProxyType"] != "-" {
		response.ProxyType = result["ProxyType"]
	}
	if result["Domain"] != "-" {
		response.Domain = result["Domain"]
	}
	if result["UsageType"] != "-" {
		response.UsageType = result["UsageType"]
	}
	if result["Asn"] != "-" {
		response.ASN = result["Asn"]
	}
	if result["As"] != "-" {
		response.AS = result["As"]
	}

	return response
}

func (i2l *Ip2Location) Ready() bool {
	i2l.mx.RLock()
	defer i2l.mx.RUnlock()

	return i2l.db != nil
}

func (i2l *Ip2Location) IsImporting() bool {
	i2l.mx.RLock()
	defer i2l.mx.RUnlock()

	return i2l.importing > 0
}

func (i2l *Ip2Location) AddImport() {
	i2l.mx.Lock()
	defer i2l.mx.Unlock()

	i2l.importing++
}

func (i2l *Ip2Location) SubImport() {
	i2l.mx.Lock()
	defer i2l.mx.Unlock()

	i2l.importing--
}

func (i2l *Ip2Location) SetImporting(state int) {
	i2l.mx.Lock()
	defer i2l.mx.Unlock()

	i2l.importing = state
}

func Ip2LocationResponseFromArray(data []string) *Ip2LocationCsvRecord {
	response := &Ip2LocationCsvRecord{}
	response.IpFrom, _ = strconv.Atoi(data[0])
	response.IpTo, _ = strconv.Atoi(data[1])
	response.ProxyType = data[2]
	response.CountryCode = data[3]
	response.CountryName = data[4]
	response.RegionName = data[5]
	response.CityName = data[6]
	response.Isp = data[7]
	response.Domain = data[8]
	response.UsageType = data[9]
	response.Asn, _ = strconv.Atoi(data[10])
	response.As = data[11]
	response.LastSeen, _ = strconv.Atoi(data[12])

	return response
}

func IntToIp(nn *big.Int) net.IP {
	ip := nn.Bytes()
	var a net.IP = ip
	return a
}
