package mmdb

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/oschwald/maxminddb-golang"

	"../config"
	"../updater"
)

func NewDefaultConfig(c *config.Config, productID string) *DB {
	conf := &DB{
		Config: c,
		ErrUnavailable: errors.New("no database available"),
	}
	conf.Updater = updater.NewDefaultConfig(c.MMUpdateInterval, c.MMRetryInterval,
		filepath.Join(c.RootDir, "cache", productID + ".mmdb"),
		filepath.Join(c.RootDir, "cache", productID + ".tar.gz"),
		conf.GenerateUpdateURL(productID), conf.newReader)

	return conf
}

// Generate the update url for the current product database.
func (d *DB) GenerateUpdateURL(productID string) string {
	u := "https://" + d.Config.MMUpdatesHost + "/app/" + "geoip_download?edition_id=" + productID +
		"&date=&license_key=" + d.Config.MMLicenseKey + "&suffix=tar.gz"
	return u
}

func (d *DB) Start() (*updater.Config, error){
	return d.Updater.OpenURL()
}

func (d *DB) newReader() error {
	err, dbFileName := d.Updater.ProcessFile()
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
	stat, err := os.Stat(d.Updater.Archive)
	if err != nil {
		return err
	}
	d.setReader(reader, stat.ModTime())
	return nil
}

func (d *DB) setReader(reader *maxminddb.Reader, modtime time.Time) {
	d.Updater.Mu.Lock()
	defer d.Updater.Mu.Unlock()
	if d.Updater.Closed {
		if err := reader.Close(); err != nil {}
		return
	}
	if d.reader != nil {
		if err := d.reader.Close(); err != nil {}
	}
	d.reader = reader
	d.Updater.LastUpdated = modtime.UTC()
	select {
	case d.Updater.Notifier.Open <- d.Updater.File:
	default:
	}
}

// Lookup performs a database lookup of the given IP address, and stores
// the response into the result value. The result value must be a struct
// with specific fields and tags as described here:
// https://godoc.org/github.com/oschwald/maxminddb-golang#Reader.Lookup
//
// See the DefaultQuery for an example of the result struct.
func (d *DB) Lookup(addr net.IP, result interface{}) error {
	d.Updater.Mu.RLock()
	defer d.Updater.Mu.RUnlock()
	if d.reader != nil {
		return d.reader.Lookup(addr, result)
	}
	return d.ErrUnavailable
}

// Close closes the database.
func (d *DB) Close() {
	d.Updater.Close()
	if d.reader != nil {
		if err := d.reader.Close(); err != nil {}
		d.reader = nil
	}
}
