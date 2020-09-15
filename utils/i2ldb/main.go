package i2ldb

import (
	"../config"
	"../updater"
	"fmt"
	"github.com/ip2location/ip2proxy-go"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {

	Updater 		*updater.Config			// Holds all notification channels
	Config 			*config.Config		// Shared default configuration

	DB          	map[string]bool
}

type ProxyDefaultQuery struct {
	Isp string
	ProxyType string
	Domain string
	UsageType string
	Asn string
	As string
	LastSeen int
	Proxy bool
}

func NewDefaultConfig(c *config.Config) *Config {
	conf := &Config{
		Config: c,
	}
	dbFile := filepath.Join(c.RootDir, "cache", c.I2LProductID + ".bin")
	dbArchive := filepath.Join(c.RootDir, "cache", c.I2LProductID + ".zip")
	conf.Updater = updater.NewDefaultConfig(c.I2LUpdateInterval, c.I2LRetryInterval,
		dbFile, dbArchive,
		conf.GenerateUpdateURL(),
		conf.newReader)

	return conf
}

// Generate the update url for the current product database.
func (c *Config) GenerateUpdateURL() string {
	u := "https://www.ip2location.com/download/?token="+ c.Config.I2LToken+"&file=" + c.Config.I2LProductID
	return u
}

func (c *Config) Start() (*updater.Config, error){
	return c.Updater.OpenURL()
}

func (c *Config) newReader() error {
	stat, err := os.Stat(c.Updater.Archive)
	if err != nil {
		return err
	}

	if stat.Size() < 800 {
		err := fmt.Errorf("DB File not available")
		c.Updater.SendError(err)
		return err
	}

	err, _ = c.Updater.ProcessFile()
	if err != nil {
		return err
	}

	_, err = os.Open(c.Updater.File)
	if err != nil {
		return err
	}

	if ip2proxy.Open(c.Updater.File) != 0 {
		err := fmt.Errorf("DB failed to load")
		c.Updater.SendError(err)
		return err
	}

	c.Updater.Mu.Lock()
	defer c.Updater.Mu.Unlock()
	if c.Updater.Closed {
		ip2proxy.Close()
		return nil
	}

	c.Updater.LastUpdated = stat.ModTime().UTC()
	select {
	case c.Updater.Notifier.Open <- c.Updater.File:
	default:
	}
	return nil
}

func (c *Config) Lookup(addr net.IP) ProxyDefaultQuery {
	c.Updater.Mu.RLock()
	defer c.Updater.Mu.RUnlock()
	result := ip2proxy.GetAll(addr.String())

	var IsProxy bool;if i, _ := strconv.Atoi(result["isProxy"]); i > 0 {IsProxy = true}
	LastSeen, _ := strconv.Atoi(result["LastSeen"])
	var ISP string;if result["ISP"] != "-" {ISP = result["ISP"]}
	var ProxyType string;if result["ProxyType"] != "-" {ProxyType = result["ProxyType"]}
	var Domain string;if result["Domain"] != "-" {Domain = result["Domain"]}
	var UsageType string;if result["UsageType"] != "-" {UsageType = result["UsageType"]}
	var Asn string;if result["Asn"] != "-" {Asn = result["Asn"]}
	var As string;if result["As"] != "-" {As = result["As"]}

	return ProxyDefaultQuery{
		Isp: ISP,
		ProxyType: ProxyType,
		Domain: Domain,
		UsageType: UsageType,
		Asn: Asn,
		As: As,
		LastSeen: LastSeen,
		Proxy: IsProxy,
	}
}

// Close closes the database.
func (c *Config) Close() {
	c.Updater.Close()
	ip2proxy.Close()
}