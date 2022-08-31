package app

import (
	"fmt"
	"sync"
)

type Worker struct {
	idle bool
	app  *Application
	mx   sync.RWMutex
}

func NewWorker(app *Application) *Worker {
	return &Worker{
		idle: true,
		app:  app,
		mx:   sync.RWMutex{},
	}
}

func (w *Worker) Work(record *Record) {
	w.SetIdle(false)
	go func(record *Record) {
		defer w.SetIdle(true)
		if err := w.app.importRecord(record); err != nil {
			fmt.Printf("[error] %s\n", err.Error())
			return
		}
	}(record)
}

func (w *Worker) SetIdle(state bool) {
	w.mx.Lock()
	defer w.mx.Unlock()

	w.idle = state
}

func (w *Worker) IsIdle() bool {
	w.mx.RLock()
	defer w.mx.RUnlock()

	return w.idle
}
