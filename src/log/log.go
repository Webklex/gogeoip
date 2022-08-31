package log

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Log struct {
	Enabled    bool   `json:"enabled"`
	ToStdout   bool   `json:"stdout"`
	Timestamp  bool   `json:"timestamp"`
	OutputFile string `json:"output_file"`

	Output *os.File `json:"-"`
}

func (l *Log) Initialize() {
	l.Output = os.Stdout

	if l.ToStdout {
		log.SetOutput(l.Output)
	} else if l.OutputFile != "" {
		_, _ = MakeDir(l.OutputFile)
		_ = ioutil.WriteFile(l.OutputFile, []byte(""), 0644)
		l.Output, _ = os.OpenFile(l.OutputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		log.SetOutput(l.Output)
	}
	if !l.Timestamp {
		log.SetFlags(0)
	}
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
