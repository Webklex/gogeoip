package api

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Updater struct {
	file      string
	archive   string
	updateUrl string

	interval      time.Duration
	retryInterval time.Duration
	lastUpdated   time.Time
	blockedUntil  time.Time

	cbk   func() error
	mx    sync.RWMutex
	close chan bool
}

func NewUpdater() *Updater {
	return &Updater{
		blockedUntil: time.Now(),
	}
}

func (u *Updater) Start() {
	u.mx.Lock()
	defer u.mx.Unlock()

	u.blockedUntil = time.Now()

	if _, err := MakeDir(u.file); err != nil {
		fmt.Printf("[error] %s\n", err.Error())
		return
	}
	if _, err := MakeDir(u.archive); err != nil {
		fmt.Printf("[error] %s\n", err.Error())
		return
	}

	if updateRequired, _ := u.updateRequired(); updateRequired == false {
		if err := u.cbk(); err != nil {
			fmt.Printf("[error] %s\n", err.Error())
			return
		}
	}

	ticker := time.NewTicker(5 * time.Second)
	u.close = make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				// do stuff
				if err := u.Update(); err != nil {
					fmt.Printf("[error] %s\n", err.Error())
					u.blockedUntil = time.Now().Add(u.retryInterval)
				}
			case <-u.close:
				ticker.Stop()
			}
		}
	}()
}

func (u *Updater) Stop() {
	if u.close != nil {
		u.close <- true
	}
	return
}

func (u *Updater) SetUpdated(t time.Time) {
	u.mx.Lock()
	defer u.mx.Unlock()

	u.lastUpdated = t
}

func (u *Updater) Update() error {
	u.mx.Lock()
	defer u.mx.Unlock()

	yes, err := u.updateRequired()
	if err != nil {
		return err
	}
	if !yes {
		return nil
	}
	u.blockedUntil = time.Now().Add(u.interval)

	fmt.Printf("Downloading: %s\n", u.updateUrl)
	tmpFile, err := u.download()
	if err != nil {
		return err
	}
	err = RenameFile(tmpFile, u.archive)
	if err != nil {
		// Cleanup the temp file if renaming failed.
		_ = os.RemoveAll(tmpFile)
	} else if err = u.cbk(); err != nil {
		return err
	}
	return err
}

func (u *Updater) updateRequired() (bool, error) {
	if time.Now().Before(u.blockedUntil) {
		return false, nil
	}
	stat, err := os.Stat(u.file)
	if err != nil {
		return true, nil // Local db is missing, must be downloaded.
	}

	if time.Now().Sub(stat.ModTime()) < u.interval/12 {
		return false, nil
	}

	resp, err := http.Head(u.updateUrl)
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

	err = os.Chtimes(u.file, time.Now(), time.Now())

	return false, nil
}

func (u *Updater) download() (tmpFile string, err error) {
	resp, err := http.Get(u.updateUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tmpFile = fmt.Sprintf(u.file+"%d", time.Now().UnixNano())
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

func (u *Updater) ProcessFile() (error, string) {
	f, err := os.Open(u.archive)
	if err != nil {
		return err, ""
	}
	defer f.Close()

	err, _ = u.ExtractTarGz(f)
	if err != nil {
		err, _ = u.ExtractZip()
	}
	return err, u.file
}

func (u *Updater) ExtractZip() (error, string) {
	r, err := zip.OpenReader(u.archive)
	if err != nil {
		return err, ""
	}
	defer func() {
		if err := r.Close(); err != nil {
		}
	}()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		if f.FileInfo().IsDir() == false {
			if strings.Contains(f.Name, "mmdb") || strings.Contains(f.Name, "BIN") {
				outFile, err := os.Create(u.file)
				if err != nil {
					return err
				}
				defer outFile.Close()
				if _, err := io.Copy(outFile, rc); err != nil {
					return err
				} else {
					return nil
				}
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			break
		}
	}

	return nil, u.file
}

func (u *Updater) ExtractTarGz(gzipStream io.Reader) (error, string) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err, ""
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err, ""
		}

		switch header.Typeflag {
		case tar.TypeDir:
			break
		case tar.TypeReg:
			if strings.Contains(header.Name, "mmdb") || strings.Contains(header.Name, "BIN") {
				outFile, err := os.Create(u.file)
				if err != nil {
					return err, ""
				}
				if _, err := io.Copy(outFile, tarReader); err != nil {
					_ = outFile.Close()
					return err, ""
				} else {
					_ = outFile.Close()
					return nil, u.file
				}
			}
		default:
			log.Fatalf("ExtractTarGz: uknown type: %b in %s", header.Typeflag, header.Name)
		}
	}

	return nil, ""
}

func (u *Updater) ExtractAllFromZip() error {
	r, err := zip.OpenReader(u.archive)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
		}
	}()

	dbdir := strings.TrimSuffix(u.archive, filepath.Ext(u.archive))
	err = MakeAbsoluteDir(dbdir)
	if err != nil {
		fmt.Printf("[error] failed to create directory: %s", err.Error())
		return err
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		if f.FileInfo().IsDir() == false {
			if strings.Contains(strings.ToLower(f.Name), ".csv") {
				outFile, err := os.Create(path.Join(dbdir, filepath.Base(f.Name)))
				if err != nil {
					return err
				}
				defer outFile.Close()
				if _, err := io.Copy(outFile, rc); err != nil {
					return err
				} else {
					return nil
				}
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func RenameFile(fromName string, toName string) error {
	_ = os.Rename(toName, toName+".bak")
	if _, err := MakeDir(toName); err != nil {
		return err
	}
	return os.Rename(fromName, toName)
}

func MakeDir(filename string) (dbdir string, err error) {
	dbdir = filepath.Dir(filename)
	_, err = os.Stat(dbdir)
	if err != nil {
		err = os.MkdirAll(dbdir, 0755)
		if err != nil {
			return "", err
		}
	}
	return dbdir, nil
}

func MakeAbsoluteDir(dbdir string) error {
	_, err := os.Stat(dbdir)
	if err != nil {
		err = os.MkdirAll(dbdir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
