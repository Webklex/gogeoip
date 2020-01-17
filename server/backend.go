package server

import (
	"github.com/go-web/httplog"
	"github.com/go-web/httpmux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"strings"

	_ "net/http/pprof"

	"github.com/fiorix/go-listener/listener"
)

type writerFunc func(w http.ResponseWriter, r *http.Request, d *responseRecord)

// NewHandler creates an http handler for the geoip server that
// can be embedded in other servers.
func (s *Server) NewHandler() (http.Handler, error) {
	s.initCors()
	mc := httpmux.DefaultConfig
	if err := s.initMiddlewares(&mc); err != nil {
		return nil, err
	}
	mux := httpmux.NewHandler(&mc)
	mux.GET("/csv/*host",  s.registerHandler(csvResponse))
	mux.GET("/xml/*host",  s.registerHandler(xmlResponse))
	mux.GET("/json/*host", s.registerHandler(jsonResponse))
	go s.watchEvents()
	return mux, nil
}

func (s *Server) initCors() {
	s.Api.cors = cors.New(cors.Options{
		AllowedOrigins:   strings.Split(s.Config.CORSOrigin, ","),
		AllowedMethods:   []string{"GET"},
		AllowCredentials: true,
	})
}

// watchEvents logs and collect metrics of database events.
func (s *Server) watchEvents() {
	for {
		select {
		case file := <-s.Api.db.NotifyOpen():
			log.Println("database loaded:", file)
		case err := <-s.Api.db.NotifyError():
			log.Println("database error:", err)
		case msg := <-s.Api.db.NotifyInfo():
			log.Println("database info:", msg)
		case <-s.Api.db.NotifyClose():
			return
		}
	}
}

func (s *Server) registerHandler(writer writerFunc) http.HandlerFunc {
	return s.Api.cors.Handler(s.IpLookUp(writer)).ServeHTTP
}

func (s *Server) listenerOpts() []listener.Option {
	var opts []listener.Option
	if s.Config.FastOpen {
		opts = append(opts, listener.FastOpen())
	}
	if s.Config.Naggle {
		opts = append(opts, listener.Naggle())
	}
	return opts
}

func (s *Server) runServer(f http.Handler) {
	log.Println("geoip http server starting on", s.Config.ServerAddr)
	ln, err := listener.New(s.Config.ServerAddr, s.listenerOpts()...)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Handler:      f,
		ReadTimeout:  s.Config.ReadTimeout,
		WriteTimeout: s.Config.WriteTimeout,
		ErrorLog:     s.Config.ErrorLogger(),
	}
	log.Fatal(srv.Serve(ln))
}

func (s *Server) runTLSServer(f http.Handler) {
	log.Println("geoip https server starting on", s.Config.TLSServerAddr)
	opts := s.listenerOpts()
	if s.Config.HTTP2 {
		opts = append(opts, listener.HTTP2())
	}
	if s.Config.LetsEncrypt {
		if s.Config.LetsEncryptHosts == "" {
			log.Fatal("must set at least one host using --letsencrypt-hosts")
		}
		opts = append(opts, listener.LetsEncrypt(
			s.Config.LetsEncryptCacheDir,
			s.Config.LetsEncryptEmail,
			strings.Split(s.Config.LetsEncryptHosts, ",")...,
		))
	} else {
		opts = append(opts, listener.TLS(s.Config.TLSCertFile, s.Config.TLSKeyFile))
	}
	ln, err := listener.New(s.Config.TLSServerAddr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Addr:         s.Config.TLSServerAddr,
		Handler:      f,
		ReadTimeout:  s.Config.ReadTimeout,
		WriteTimeout: s.Config.WriteTimeout,
		ErrorLog:     s.Config.ErrorLogger(),
		TLSConfig:    ln.TLSConfig(),
	}
	log.Fatal(srv.Serve(ln))
}

func (s *Server) initMiddlewares(mc *httpmux.Config) error {
	mc.Prefix = s.Config.APIPrefix
	mc.NotFound = guiMiddleware(s.Config.GuiDir)
	if s.Config.UseXForwardedFor {
		mc.UseFunc(httplog.UseXForwardedFor)
	}
	if !s.Config.Silent {
		mc.UseFunc(httplog.ApacheCombinedFormat(s.Config.AccessLogger()))
	}
	if s.Config.HSTS != "" {
		mc.UseFunc(hstsMiddleware(s.Config.HSTS))
	}
	if s.Config.RateLimitLimit > 0 {
		mc.Use(s.rateLimitMiddleware)
	}
	return nil
}

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := s.RateLimit.GetLimiter(r.RemoteAddr)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
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

func guiMiddleware(path string) http.Handler {
	handler := http.NotFoundHandler()
	if path != "" {
		handler = http.FileServer(http.Dir(path))
	}

	return handler
}

