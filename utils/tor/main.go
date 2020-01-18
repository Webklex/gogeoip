package tor

import (
	"../config"
	"../updater"
	"bufio"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {

	Updater 		*updater.Config			// Holds all notification channels
	Config 			*config.Config		// Shared default configuration

	DB          	map[string]bool
}

func NewDefaultConfig(c *config.Config) *Config {
	conf := &Config{
		Config: c,
	}
	dbFile := filepath.Join(c.RootDir, "cache", "tor.db")
	conf.Updater = updater.NewDefaultConfig(c.TorUpdateInterval, c.TorRetryInterval,
		dbFile, dbFile,
		"https://" + c.TorUpdatesHost +"/cgi-bin/TorBulkExitList.py?ip=" + c.TorExitCheck,
		conf.newReader)

	return conf
}

func (c *Config) Start() (*updater.Config, error){
	return c.Updater.OpenURL()
}

func (c *Config) Lookup(addr net.IP) bool {
	return c.LookupString(addr.String())
}

func (c *Config) LookupString(lookup string) bool {
	_, exists := c.DB[lookup]
	return exists
}

func (c *Config) newReader() error {
	f, err := os.Open(c.Updater.File)
	defer f.Close()
	if err != nil {
		return err
	}
	// Start reading from the file with a reader.
	reader := bufio.NewReader(f)

	c.DB = make(map[string]bool)

	var line string
	for {
		line, err = reader.ReadString('\n')
		if line != "#" {
			addr := strings.Replace(line, "\n", "", 1)
			c.DB[addr] = true
		}

		if err != nil {
			break
		}
	}
	select {
	case c.Updater.Notifier.Open <- c.Updater.File:
	default:
	}
	return nil
}