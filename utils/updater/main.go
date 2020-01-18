package updater

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"io"
	"math"
	"net/http"
	"os"
	"sync"
	"time"
)

type Config struct {

	Notifier       *Notifier // Holds all notification channels
	File           string
	Archive        string
	Closed         bool // Mark this db as closed.
	UpdateInterval time.Duration
	RetryInterval  time.Duration
	LastUpdated    time.Time    // Last time the db was updated.
	updateUrl      string       // tor project update url
	Mu             sync.RWMutex // Protects all the above.

	cbk          	func() error
}

func NewDefaultConfig(UpdateInterval time.Duration, RetryInterval time.Duration, File string, Archive string,  updateUrl string, cbk func() error) *Config {
	return &Config{
		UpdateInterval: UpdateInterval,
		RetryInterval: RetryInterval,
		File: File,
		Archive: Archive,
		cbk: cbk,
		Notifier: &Notifier{
			Quit:  make(chan struct{}),
			Open:  make(chan string, 1),
			Error: make(chan error, 1),
			Info:  make(chan string, 1),
		},
		updateUrl:   updateUrl,
	}
}

// Open creates and initializes a DB from a local file.
//
// The database file is monitored by fsnotify and automatically
// reloads when the file is updated or overwritten.
func (c *Config) Open() error {
	err := c.openFile()
	if err != nil {
		c.Close()
		return err
	}
	err = c.watchFile()
	if err != nil {
		c.Close()
		return fmt.Errorf("fsnotify failed for %s: %s", c.File, err)
	}
	return nil
}

// OpenURL creates and initializes a DB from a URL.
// It automatically downloads and updates the file in background, and
// keeps a local copy on $TMPDIR.
func (c *Config) OpenURL() (*Config, error) {
	// Optional, might fail.
	if err := c.openFile(); err != nil {}

	go c.autoUpdate()
	if err := c.watchFile(); err != nil {
		c.Close()
		return nil, fmt.Errorf("fsnotify failed for %s: %s", c.File, err)
	}
	return c, nil
}

func (c *Config) autoUpdate() {
	backoff := time.Second
	for {
		c.SendInfo("Checking for updates")
		err := c.runUpdate(c.updateUrl)
		if err != nil {
			bs := backoff.Seconds()
			ms := c.RetryInterval.Seconds()
			backoff = time.Duration(math.Min(bs*math.E, ms)) * time.Second
			c.SendError(fmt.Errorf("download failed (will retry in %s): %s", backoff, err))
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
		c.SendInfo("DB is up to date")
		return nil
	}
	c.SendInfo("starting update")
	tmpFile, err := c.download(url)
	if err != nil {
		return err
	}
	err = RenameFile(tmpFile, c.Archive)
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
	if LastModified != "" {
		t, err := time.Parse(http.TimeFormat, LastModified)
		if err != nil {
			return false, err
		}
		if t.After(stat.ModTime()) {
			return true, nil
		}
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
	tmpFile = fmt.Sprintf(c.File + "%d", time.Now().UnixNano())
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
	dbdir, err := MakeDir(c.Archive)
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
			if ev.Name == c.Archive && (ev.IsCreate() || ev.IsModify()) {
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
	err := c.cbk()
	if err != nil {
		return err
	}
	stat, err := os.Stat(c.File)
	if err != nil {
		return err
	}
	c.LastUpdated = stat.ModTime()

	return nil
}

// Close closes the database.
func (c *Config)  Close() {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if !c.Closed {
		c.Closed = true
		close(c.Notifier.Quit)
		close(c.Notifier.Open)
		close(c.Notifier.Error)
		close(c.Notifier.Info)
	}
}

// Date returns the UTC date the database file was last modified.
// If no database file has been opened the behaviour of Date is undefined.
func (c *Config) Date() time.Time {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	return c.LastUpdated
}