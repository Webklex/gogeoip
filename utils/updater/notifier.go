package updater

type Notifier struct {

	Quit  chan struct{}     // Stop the db auto-update and all active watch goroutines.
	Open  chan string       // Notify when a new db file is open.
	Error chan error        // Notify when an error occurs.
	Info  chan string       // Notify info actions for logging
}

// NotifyClose returns a channel that is closed when the database is closed.
func (c *Config) NotifyClose() <-chan struct{} {
	return c.Notifier.Quit
}

// NotifyOpen returns a channel that notifies when a new database is
// loaded or reloaded. This can be used to monitor background updates
// when the Config points to a URL.
func (c *Config) NotifyOpen() (filename <-chan string) {
	return c.Notifier.Open
}

// NotifyError returns a channel that notifies when an error occurs
// while downloading or reloading a Config that points to a URL.
func (c *Config) NotifyError() (errChan <-chan error) {
	return c.Notifier.Error
}

// NotifyInfo returns a channel that notifies informational messages
// while downloading or reloading.
func (c *Config) NotifyInfo() <-chan string {
	return c.Notifier.Info
}

func (c *Config) SendError(err error) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	if c.Closed {
		return
	}
	select {
	case c.Notifier.Error <- err:
	default:
	}
}

func (c *Config) SendInfo(message string) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	if c.Closed {
		return
	}
	select {
	case c.Notifier.Info <- message:
	default:
	}
}