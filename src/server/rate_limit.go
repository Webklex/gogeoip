package server

import (
	"golang.org/x/time/rate"
)

type RateLimit struct {
	Limit rate.Limit `json:"limit"`
	Burst int        `json:"int"`
}
