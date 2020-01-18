package db

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (d *DB) autoUpdate() {
	backoff := time.Second
	for {
		d.sendInfo("Checking for updates")
		err := d.runUpdate(d.updateUrl)
		if err != nil {
			bs := backoff.Seconds()
			ms := d.Config.RetryInterval.Seconds()
			backoff = time.Duration(math.Min(bs*math.E, ms)) * time.Second
			d.sendError(fmt.Errorf("download failed (will retry in %s): %s", backoff, err))
		} else {
			backoff = d.Config.UpdateInterval
		}
		select {
		case <-d.Notifier.Quit:
			return
		case <-time.After(backoff):
			// Sleep till time for the next update attempt.
		}
	}
}

func (d *DB) runUpdate(url string) error {
	yes, err := d.needUpdate(url)
	if err != nil {
		return err
	}
	if !yes {
		d.sendInfo("DB is up to date")
		return nil
	}
	d.sendInfo("starting update")
	tmpFile, err := d.download(url)
	if err != nil {
		return err
	}
	err = RenameFile(tmpFile, d.dbArchive)
	if err != nil {
		// Cleanup the temp file if renaming failed.
		if err := os.RemoveAll(tmpFile); err != nil {}
	}
	return err
}

func (d *DB) needUpdate(url string) (bool, error) {
	stat, err := os.Stat(d.File)
	if err != nil {
		return true, nil // Local db is missing, must be downloaded.
	}

	if time.Now().Sub(stat.ModTime()) < d.Config.UpdateInterval / 12 {
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
	err = os.Chtimes(d.File, time.Now(), time.Now())

	return false, nil
}

func (d *DB) download(url string) (tmpFile string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tmpFile = filepath.Join(d.Config.RootDir, "cache", fmt.Sprintf("_geoip.%d.db.gz", time.Now().UnixNano()))
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

// Date returns the UTC date the database file was last modified.
// If no database file has been opened the behaviour of Date is undefined.
func (d *DB) Date() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastUpdated
}