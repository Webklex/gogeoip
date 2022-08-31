package api

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Tor struct {
	ExitCheck      string        `json:"exit_check"`
	RetryInterval  time.Duration `json:"retry_interval"`
	UpdateInterval time.Duration `json:"update_interval"`
	Downstreams    string        `json:"downstreams"`

	updater *Updater
	db      map[string]bool

	mx sync.RWMutex
}

func (t *Tor) Start(rootDir string) {
	t.db = map[string]bool{}
	t.updater = &Updater{
		file:          filepath.Join(rootDir, "cache", "tor.db"),
		archive:       filepath.Join(rootDir, "cache", "tor.db"),
		interval:      t.UpdateInterval,
		retryInterval: t.RetryInterval,
		updateUrl:     t.UpdateURL(),
		cbk:           t.NewReader,
	}
	go t.updater.Start()
}

func (t *Tor) Stop() {
	t.updater.Stop()
}

func (t *Tor) UpdateURL() string {
	return fmt.Sprintf("https://%s/cgi-bin/TorBulkExitList.py?ip=%s", t.Downstreams, t.ExitCheck)
}

func (t *Tor) Lookup(addr net.IP) bool {
	return t.LookupString(addr.String())
}

func (t *Tor) LookupString(lookup string) bool {
	t.mx.RLock()
	defer t.mx.RUnlock()

	_, exists := t.db[lookup]
	return exists
}

func (t *Tor) NewReader() error {
	t.mx.Lock()
	defer t.mx.Unlock()

	f, err := os.Open(t.updater.file)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	t.db = make(map[string]bool)

	var line string
	for {
		line, err = reader.ReadString('\n')
		if line != "#" {
			addr := strings.Replace(line, "\n", "", 1)
			t.db[addr] = true
		}

		if err != nil {
			break
		}
	}
	return nil
}

func (t *Tor) Ready() bool {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.db != nil && len(t.db) > 0
}
