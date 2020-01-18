package tor

import (
	"../config"
	"../db"
	"bufio"
	"fmt"
	"github.com/howeyc/fsnotify"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {

	Notifier 		*db.Notifier // Holds all notification channels
	File 			string
	RootDir 		string
	UpdateInterval 	time.Duration
	RetryInterval 	time.Duration
	lastUpdated 	time.Time    // Last time the db was updated.
	updateUrl		string		 // tor project update url

	DB          	[]string
}

func NewDefaultConfig(c *config.Config) *Config {
	return &Config{
		UpdateInterval: c.TorUpdateInterval,
		RetryInterval: c.RetryInterval,
		RootDir: c.RootDir,
		File: filepath.Join(c.RootDir, "cache", "tor.db"),
		Notifier: &db.Notifier{
			Quit:  make(chan struct{}),
			Open:  make(chan string, 1),
			Error: make(chan error, 1),
			Info:  make(chan string, 1),
		},
		updateUrl:   "https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=" + c.TorExitCheck,
	}
}

// OpenURL creates and initializes a DB from a URL.
// It automatically downloads and updates the file in background, and
// keeps a local copy on $TMPDIR.
func (c *Config) OpenURL() (*Config, error) {
	// Optional, might fail.
	if err := c.openFile(); err != nil {}

	go c.autoUpdate()
	if err := c.watchFile(); err != nil {
		return nil, fmt.Errorf("fsnotify failed for %s: %s", c.File, err)
	}
	return c, nil
}

func (c *Config) Lookup(addr net.IP) bool {
	return c.LookupString(addr.String())
}
func (c *Config) LookupString(lookup string) bool {
	for _, val := range c.DB {
		if val == lookup {
			return true
		}
	}
	return false
}

func (c *Config) autoUpdate() {
	backoff := time.Second
	for {
		c.sendInfo("Checking for updates")
		err := c.runUpdate(c.updateUrl)
		if err != nil {
			bs := backoff.Seconds()
			ms := c.RetryInterval.Seconds()
			backoff = time.Duration(math.Min(bs*math.E, ms)) * time.Second
			c.sendError(fmt.Errorf("download failed (will retry in %s): %s", backoff, err))
		} else {
			backoff = c.UpdateInterval
		}
		select {
		case <-c.Notifier.Quit:
			return
		case <-time.After(backoff):
			// Sleep till time for the next update attempt.
		}
	}
}

func (c *Config) runUpdate(url string) error {
	yes, err := c.needUpdate(url)
	if err != nil {
		return err
	}
	if !yes {
		c.sendInfo("DB is up to date")
		return nil
	}
	c.sendInfo("starting update")
	tmpFile, err := c.download(url)
	if err != nil {
		return err
	}
	err = db.RenameFile(tmpFile, c.File)
	if err != nil {
		// Cleanup the temp file if renaming failed.
		if err := os.RemoveAll(tmpFile); err != nil {}
	}
	return err
}

func (c *Config) needUpdate(url string) (bool, error) {
	stat, err := os.Stat(c.File)
	if err != nil {
		return true, nil // Local db is missing, must be downloaded.
	}

	if time.Now().Sub(stat.ModTime()) < c.UpdateInterval / 12 {
		return false, nil
	}

	resp, err := http.Head(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Check X-Database-MD5 if it exists
	LastModified := resp.Header.Get("Last-Modified")
	t, err := time.Parse(http.TimeFormat, LastModified)
	if err != nil {
		return false, err
	}

	if t.After(stat.ModTime()) {
		return true, nil
	}
	err = os.Chtimes(c.File, time.Now(), time.Now())

	return false, nil
}

func (c *Config) download(url string) (tmpFile string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tmpFile = filepath.Join(c.RootDir, "cache", fmt.Sprintf("tor.%d.db", time.Now().UnixNano()))
	f, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}
	return tmpFile, nil
}


func (c *Config) watchFile() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	dbdir, err := db.MakeDir(c.File)
	if err != nil {
		return err
	}
	go c.watchEvents(watcher)
	return watcher.Watch(dbdir)
}

func (c *Config) watchEvents(watcher *fsnotify.Watcher) {
	for {
		select {
		case ev := <-watcher.Event:
			if ev.Name == c.File && (ev.IsCreate() || ev.IsModify()) {
				if err := c.openFile(); err != nil {}
			}
		case <-watcher.Error:
		case <-c.Notifier.Quit:
			if err := watcher.Close(); err != nil {}
			return
		}
		time.Sleep(time.Second) // Suppress high-rate events.
	}
}

func (c *Config) openFile() error {
	err := c.newReader()
	if err != nil {
		return err
	}
	stat, err := os.Stat(c.File)
	if err != nil {
		return err
	}
	c.lastUpdated = stat.ModTime()

	select {
	case c.Notifier.Open <- c.File:
	default:
	}

	return nil
}

func (c *Config) newReader() error {
	f, err := os.Open(c.File)
	defer f.Close()
	if err != nil {
		return err
	}
	// Start reading from the file with a reader.
	reader := bufio.NewReader(f)

	c.DB = make([]string, 0)

	var line string
	for {
		line, err = reader.ReadString('\n')
		if line != "#" {
			c.DB = append(c.DB, strings.Replace(line, "\n", "", 1))
		}

		if err != nil {
			break
		}
	}
	return nil
}

// NotifyClose returns a channel that is closed when the database is closed.
func (c *Config) NotifyClose() <-chan struct{} {
	return c.Notifier.Quit
}

// NotifyOpen returns a channel that notifies when a new database is
// loaded or reloaded. This can be used to monitor background updates
// when the DB points to a URL.
func (c *Config) NotifyOpen() (filename <-chan string) {
	return c.Notifier.Open
}

// NotifyError returns a channel that notifies when an error occurs
// while downloading or reloading a DB that points to a URL.
func (c *Config) NotifyError() (errChan <-chan error) {
	return c.Notifier.Error
}

// NotifyInfo returns a channel that notifies informational messages
// while downloading or reloading.
func (c *Config) NotifyInfo() <-chan string {
	return c.Notifier.Info
}

func (c *Config) sendError(err error) {
	select {
	case c.Notifier.Error <- err:
	default:
	}
}

func (c *Config) sendInfo(message string) {
	select {
	case c.Notifier.Info <- message:
	default:
	}
}