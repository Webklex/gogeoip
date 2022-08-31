package server

import (
	"encoding/xml"
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/fiorix/go-listener/listener"
	"github.com/go-web/httplog"
	"github.com/go-web/httpmux"
	"github.com/rs/cors"
	"github.com/webklex/gogeoip/src/models"
	"io"
	"log"
	"net/http"
	"strings"
)

type writerFunc func(w http.ResponseWriter, r *http.Request, d *ResponseRecord)

type ResponseRecord struct {
	XMLName    xml.Name `xml:"Response" json:"-"`
	*models.IP `xml:"IP" json:"ip"`
}

func (s *Server) NewHandler() (http.Handler, error) {
	s.cors = cors.New(cors.Options{
		AllowedOrigins:     strings.Split(s.CORSOrigin, ","),
		AllowedMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:     []string{"*"},
		OptionsPassthrough: true,
		AllowCredentials:   true,
		Debug:              false,
	})

	mc := httpmux.DefaultConfig
	if err := s.initMiddlewares(&mc); err != nil {
		return nil, err
	}
	mux := httpmux.NewHandler(&mc)
	s.routerFunc(mux)
	return mux, nil
}

func (s *Server) RegisterHandler(handle http.HandlerFunc) http.Handler {
	return s.cors.Handler(handle)
}

func (s *Server) listenerOpts() []listener.Option {
	var opts []listener.Option
	if s.FastOpen {
		opts = append(opts, listener.FastOpen())
	}
	if s.Naggle {
		opts = append(opts, listener.Naggle())
	}
	return opts
}

func (s *Server) runServer(f http.Handler) {
	fmt.Printf("http server listening on: http://%s\n", s.ServerAddr)
	ln, err := listener.New(s.ServerAddr, s.listenerOpts()...)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Handler:      f,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		ErrorLog:     s.ErrorLogger(),
	}
	log.Fatal(srv.Serve(ln))
}

func (s *Server) runTLSServer(f http.Handler) {
	fmt.Printf("https server listening on: https://%s\n", s.Tls.ServerAddr)
	opts := s.listenerOpts()
	if s.HTTP2 {
		opts = append(opts, listener.HTTP2())
	}
	if s.LetsEncrypt.Enabled {
		if s.LetsEncrypt.Hosts == "" {
			log.Fatal("must set at least one host using --letsencrypt-hosts")
		}
		opts = append(opts, listener.LetsEncrypt(
			s.LetsEncrypt.CacheDir,
			s.LetsEncrypt.Email,
			strings.Split(s.LetsEncrypt.Hosts, ",")...,
		))
	} else {
		opts = append(opts, listener.TLS(s.Tls.CertFile, s.Tls.KeyFile))
	}
	ln, err := listener.New(s.Tls.ServerAddr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Addr:         s.Tls.ServerAddr,
		Handler:      f,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		ErrorLog:     s.ErrorLogger(),
		TLSConfig:    ln.TLSConfig(),
	}
	log.Fatal(srv.Serve(ln))
}

func (s *Server) initMiddlewares(mc *httpmux.Config) error {
	mc.Prefix = s.APIPrefix
	// mc.NotFound = guiMiddleware(s.GuiDir)
	if s.UseXForwardedFor {
		mc.UseFunc(httplog.UseXForwardedFor)
	}
	mc.UseFunc(httplog.ApacheCombinedFormat(s.AccessLogger()))
	if s.HSTS != "" {
		mc.UseFunc(hstsMiddleware(s.HSTS))
	}
	if s.RateLimit.Limit > 0 {
		mc.Use(s.rateLimitMiddleware)
	}
	return nil
}

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	lmt := tollbooth.NewLimiter(float64(s.RateLimit.Limit)/60.0, nil)
	if s.RateLimit.Burst > 0 {
		lmt.SetBurst(s.RateLimit.Burst)
	}

	return tollbooth.LimitHandler(lmt, next)
}

func hstsMiddleware(policy string) httpmux.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil {
				return
			}
			w.Header().Set("Strict-Transport-Security", policy)
			next(w, r)
		}
	}
}

func (s *Server) logWriter() io.Writer {
	return s.logger.Output
}

func (s *Server) ErrorLogger() *log.Logger {
	if s.logger.Timestamp {
		return log.New(s.logWriter(), "[error] ", log.LstdFlags)
	}
	return log.New(s.logWriter(), "[error] ", 0)
}

func (s *Server) AccessLogger() *log.Logger {
	return log.New(s.logWriter(), "[access] ", 0)
}
