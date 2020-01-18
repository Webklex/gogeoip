package db

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/howeyc/fsnotify"
	"github.com/oschwald/maxminddb-golang"

	"../config"
)

func NewDefaultConfig(c *config.Config, productID string) *DB {
	conf := &DB{
		Config: c,
		Notifier: &Notifier{
			Quit:  make(chan struct{}),
			Open:  make(chan string, 1),
			Error: make(chan error, 1),
			Info:  make(chan string, 1),
		},
		ErrUnavailable: errors.New("no database available"),
		dbArchive: filepath.Join(c.RootDir, "cache", productID + ".tar.gz"),
		File: filepath.Join(c.RootDir, "cache", productID + ".mmdb"),
	}
	conf.updateUrl = conf.GenerateUpdateURL(productID)

	return conf
}

// Open creates and initializes a DB from a local file.
//
// The database file is monitored by fsnotify and automatically
// reloads when the file is updated or overwritten.
func (d *DB) Open() (*DB, error) {
	err := d.openFile()
	if err != nil {
		d.Close()
		return nil, err
	}
	err = d.watchFile()
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("fsnotify failed for %s: %s", d.File, err)
	}
	return d, nil
}

// OpenURL creates and initializes a DB from a URL.
// It automatically downloads and updates the file in background, and
// keeps a local copy on $TMPDIR.
func (d *DB) OpenURL() (*DB, error) {
	 // Optional, might fail.
	if err := d.openFile(); err != nil {}

	go d.autoUpdate()
	if err := d.watchFile(); err != nil {
		d.Close()
		return nil, fmt.Errorf("fsnotify failed for %s: %s", d.File, err)
	}
	return d, nil
}

// Generate the update url for the current product database.
func (d *DB) GenerateUpdateURL(productID string) string {
	u := "https://" + d.Config.UpdatesHost + "/app/" + "geoip_download?edition_id=" + productID +
		"&date=&license_key=" + d.Config.LicenseKey + "&suffix=tar.gz"
	return u
}

func (d *DB) watchFile() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	dbdir, err := d.makeDir(d.dbArchive)
	if err != nil {
		return err
	}
	go d.watchEvents(watcher)
	return watcher.Watch(dbdir)
}

func (d *DB) watchEvents(watcher *fsnotify.Watcher) {
	for {
		select {
		case ev := <-watcher.Event:
			if ev.Name == d.dbArchive && (ev.IsCreate() || ev.IsModify()) {
				if err := d.openFile(); err != nil {}
			}
		case <-watcher.Error:
		case <-d.Notifier.Quit:
			if err := watcher.Close(); err != nil {}
			return
		}
		time.Sleep(time.Second) // Suppress high-rate events.
	}
}

func (d *DB) openFile() error {
	reader, err := d.newReader()
	if err != nil {
		return err
	}
	stat, err := os.Stat(d.dbArchive)
	if err != nil {
		return err
	}
	d.setReader(reader, stat.ModTime())
	return nil
}

func (d *DB) newReader() (*maxminddb.Reader, error) {
	err, dbFileName := d.processFile()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(dbFileName)
	if err != nil {
		return nil, err
	}

	if _, err := ioutil.ReadAll(f); err != nil {
		return nil, err
	}
	mmdb, err := maxminddb.Open(dbFileName)
	return mmdb, err
}

func (d *DB) setReader(reader *maxminddb.Reader, modtime time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		if err := reader.Close(); err != nil {}
		return
	}
	if d.reader != nil {
		if err := d.reader.Close(); err != nil {}
	}
	d.reader = reader
	d.lastUpdated = modtime.UTC()
	select {
	case d.Notifier.Open <- d.File:
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
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.reader != nil {
		return d.reader.Lookup(addr, result)
	}
	return d.ErrUnavailable
}

// Close closes the database.
func (d *DB) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.closed {
		d.closed = true
		close(d.Notifier.Quit)
		close(d.Notifier.Open)
		close(d.Notifier.Error)
		close(d.Notifier.Info)
	}
	if d.reader != nil {
		if err := d.reader.Close(); err != nil {}
		d.reader = nil
	}
}
