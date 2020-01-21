package server

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type Visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimit struct {
	Limit rate.Limit
	Burst int
	Interval   time.Duration
	Mutex  *sync.RWMutex
	Visitors map[string]*Visitor
}

func NewRateLimit(limit int, burst int, interval time.Duration) *RateLimit {
	conf := &RateLimit{
		Limit: rate.Limit(limit),
		Burst: burst,
		Interval: interval,
		Visitors:  make(map[string]*Visitor),
		Mutex: &sync.RWMutex{},
	}
	go conf.cleanupVisitors()

	return conf
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise calls AddIP to add IP address to the map
func (i *RateLimit) GetLimiter(ip string) *rate.Limiter {
	i.Mutex.Lock()
	visitor, exists := i.Visitors[ip]

	if !exists {
		i.Mutex.Unlock()
		return i.AddIP(ip)
	}

	i.Mutex.Unlock()

	return visitor.limiter
}

// AddIP creates a new rate limiter and adds it to the ips map,
// using the IP address as the key
func (i *RateLimit) AddIP(ip string) *rate.Limiter {
	i.Mutex.Lock()
	defer i.Mutex.Unlock()

	limiter := rate.NewLimiter(i.Limit, i.Burst)

	i.Visitors[ip] = &Visitor{limiter, time.Now()}

	return limiter
}

func (i *RateLimit) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		i.Mutex.Lock()
		for ip, v := range i.Visitors {
			if time.Now().Sub(v.lastSeen) > i.Interval {
				delete(i.Visitors, ip)
			}
		}
		i.Mutex.Unlock()
	}
}