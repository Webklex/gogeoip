package db

type Notifier struct {

	Quit  chan struct{}     // Stop the db auto-update and all active watch goroutines.
	Open  chan string       // Notify when a new db file is open.
	Error chan error        // Notify when an error occurs.
	Info  chan string       // Notify info actions for logging
}

// NotifyClose returns a channel that is closed when the database is closed.
func (d *DB) NotifyClose() <-chan struct{} {
	return d.Notifier.Quit
}

// NotifyOpen returns a channel that notifies when a new database is
// loaded or reloaded. This can be used to monitor background updates
// when the DB points to a URL.
func (d *DB) NotifyOpen() (filename <-chan string) {
	return d.Notifier.Open
}

// NotifyError returns a channel that notifies when an error occurs
// while downloading or reloading a DB that points to a URL.
func (d *DB) NotifyError() (errChan <-chan error) {
	return d.Notifier.Error
}

// NotifyInfo returns a channel that notifies informational messages
// while downloading or reloading.
func (d *DB) NotifyInfo() <-chan string {
	return d.Notifier.Info
}

func (d *DB) sendError(err error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed {
		return
	}
	select {
	case d.Notifier.Error <- err:
	default:
	}
}

func (d *DB) sendInfo(message string) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed {
		return
	}
	select {
	case d.Notifier.Info <- message:
	default:
	}
}