package server

import (
	"fmt"
	"github.com/go-web/httpmux"
	"github.com/rs/cors"
	"github.com/webklex/gogeoip/src/log"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	ServerAddr string `json:"server_addr"`
	FastOpen   bool   `json:"fast_open"`
	Naggle     bool   `json:"naggle"`
	HTTP2      bool   `json:"http2"`
	HSTS       string `json:"hsts"`

	Tls Tls `json:"tls"`

	LetsEncrypt LetsEncrypt `json:"lets_encrypt"`

	APIPrefix  string `json:"api_prefix"`
	CORSOrigin string `json:"cors_origin"`

	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`

	UseXForwardedFor bool `json:"use_x_forwarded_for"`

	RateLimit RateLimit `json:"rate_limit"`

	db         *gorm.DB
	cors       *cors.Cors
	routerFunc RouterFunc
	logger     *log.Log
}

type RouterFunc func(mux *httpmux.Handler)

func (s *Server) Start(db *gorm.DB, logger *log.Log, routerFunc RouterFunc) error {
	s.db = db
	s.routerFunc = routerFunc
	s.logger = logger

	parts := strings.Split(s.ServerAddr, ":")
	port, err := strconv.Atoi(parts[1])

	if err != nil || port <= 0 {
		return fmt.Errorf("invalid server socket provided: %s", s.ServerAddr)
	}

	f, err := s.NewHandler()
	if err != nil {
		return err
	}

	if s.Tls.ServerAddr != "" {
		go s.runTLSServer(f)
	} else if s.ServerAddr != "" {
		go s.runServer(f)
	}

	return nil
}

func (s *Server) Stop() {

}
