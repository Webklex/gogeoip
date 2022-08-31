package app

import (
	"context"
	"embed"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ammario/ipisp/v2"
	"github.com/go-web/httpmux"
	"github.com/webklex/gogeoip/src/api"
	"github.com/webklex/gogeoip/src/log"
	"github.com/webklex/gogeoip/src/models"
	"github.com/webklex/gogeoip/src/server"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"io/fs"
	"io/ioutil"
	"math"
	"math/big"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type Build struct {
	Number  string `json:"number"`
	Version string `json:"version"`
}

type Application struct {
	Build *Build `json:"-"`

	Log log.Log `json:"log"`

	DatabaseLocation string `json:"database"`

	MaxMind     api.MaxMind     `json:"max_mind"`
	Ip2Location api.Ip2Location `json:"ip2location"`
	Tor         api.Tor         `json:"tor"`
	Server      server.Server   `json:"server"`

	File    string `json:"-"`
	RootDir string `json:"-"`

	LogOutput *os.File `json:"-"`
	db        *gorm.DB
	ready     bool
	close     chan bool
	mx        sync.RWMutex

	importQueue chan *Record
	workers     []*Worker
	numWorkers  int
	queueSize   int
	assets      embed.FS

	statistic *StatisticResponse
}

func NewApplication(build *Build, fs *flag.FlagSet, assets embed.FS) *Application {
	dir, _ := os.Getwd()

	a := &Application{
		assets: assets,
		Build:  build,
		Log: log.Log{
			Enabled:    false,
			ToStdout:   false,
			Timestamp:  false,
			OutputFile: "",
		},
		MaxMind: api.MaxMind{
			UserID:         "",
			LicenseKey:     "",
			ProductID:      "GeoLite2-City",
			Downstreams:    "download.maxmind.com",
			UpdateInterval: 4 * time.Hour,
			RetryInterval:  2 * time.Hour,
		},
		Ip2Location: api.Ip2Location{
			Token:          "",
			ProductID:      "PX8LITEBIN",
			CsvProductID:   "PX8LITEBINCSV",
			UpdateInterval: 4 * time.Hour,
			RetryInterval:  2 * time.Hour,
			Downstreams:    "www.ip2location.com",
		},
		Tor: api.Tor{
			Downstreams:    "check.torproject.org",
			UpdateInterval: 30 * time.Minute,
			RetryInterval:  2 * time.Hour,
			ExitCheck:      "8.8.8.8",
		},
		LogOutput: nil,

		Server: server.Server{
			ServerAddr: "localhost:8080",
			FastOpen:   false,
			Naggle:     false,
			HTTP2:      true,
			HSTS:       "",
			Tls: server.Tls{
				ServerAddr: "",
				CertFile:   "cert.pem",
				KeyFile:    "key.pem",
			},
			LetsEncrypt: server.LetsEncrypt{
				Enabled:  false,
				CacheDir: ".",
				Email:    "",
				Hosts:    "",
			},
			APIPrefix:        "/",
			CORSOrigin:       "*",
			ReadTimeout:      30 * time.Second,
			WriteTimeout:     15 * time.Second,
			UseXForwardedFor: false,
			RateLimit: server.RateLimit{
				Limit: 1,
				Burst: 4,
			},
		},

		DatabaseLocation: path.Join(dir, "cache", "gogeoip.db"),

		RootDir:    dir,
		File:       path.Join(dir, "config", "settings.json"),
		ready:      false,
		numWorkers: 1,
		queueSize:  64,
	}
	a.addFlags(fs)
	a.load(a.File)
	a.Log.Initialize()

	return a
}

func (a *Application) load(filename string) bool {
	if _, err := os.Stat(filename); err == nil {

		content, err := ioutil.ReadFile(filename)
		if err != nil {

			if !a.Log.Enabled {
				fmt.Printf("[error] Config file failed to load: %s", err.Error())
			}
			return false
		}

		err = json.Unmarshal(content, a)
		if err != nil {
			if !a.Log.Enabled {
				fmt.Printf("[error] Config file failed to load: %s", err.Error())
			}
			return false
		}

		if !a.Log.Enabled {
			fmt.Printf("[info] Config file loaded successfully")
		}

	}
	return true
}

// AddFlags adds configuration flags to the given FlagSet.
func (a *Application) addFlags(fs *flag.FlagSet) {
	fs.StringVar(&a.Server.ServerAddr, "http", a.Server.ServerAddr, "Address in form of ip:port to listen")
	fs.StringVar(&a.Server.Tls.ServerAddr, "https", a.Server.Tls.ServerAddr, "Address in form of ip:port to listen")

	fs.BoolVar(&a.Server.FastOpen, "tcp-fast-open", a.Server.FastOpen, "Enable TCP fast open")
	fs.BoolVar(&a.Server.Naggle, "tcp-naggle", a.Server.Naggle, "Enable TCP Nagle's algorithm (disables NO_DELAY)")
	fs.BoolVar(&a.Server.HTTP2, "http2", a.Server.HTTP2, "Enable HTTP/2 when TLS is enabled")

	fs.StringVar(&a.Server.HSTS, "hsts", a.Server.HSTS, "Set HSTS to the value provided on all responses")
	fs.StringVar(&a.Server.Tls.CertFile, "cert", a.Server.Tls.CertFile, "X.509 certificate file for HTTPS server")
	fs.StringVar(&a.Server.Tls.KeyFile, "key", a.Server.Tls.KeyFile, "X.509 key file for HTTPS server")

	fs.BoolVar(&a.Server.LetsEncrypt.Enabled, "letsencrypt", a.Server.LetsEncrypt.Enabled, "Enable automatic TLS using letsencrypt.org")
	fs.StringVar(&a.Server.LetsEncrypt.Email, "letsencrypt-email", a.Server.LetsEncrypt.Email, "Optional email to register with letsencrypt (default is anonymous)")
	fs.StringVar(&a.Server.LetsEncrypt.Hosts, "letsencrypt-hosts", a.Server.LetsEncrypt.Hosts, "Comma separated list of hosts for the certificate (required)")
	fs.StringVar(&a.Server.LetsEncrypt.CacheDir, "letsencrypt-cert-dir", a.Server.LetsEncrypt.CacheDir, "Letsencrypt cert dir")

	fs.StringVar(&a.Server.APIPrefix, "api-prefix", a.Server.APIPrefix, "API endpoint prefix")
	fs.StringVar(&a.Server.CORSOrigin, "cors-origin", a.Server.CORSOrigin, "Comma separated list of CORS origins endpoints")
	fs.BoolVar(&a.Server.UseXForwardedFor, "use-x-forwarded-for", a.Server.UseXForwardedFor, "Use the X-Forwarded-For header when available (e.g. behind proxy)")

	fs.StringVar(&a.MaxMind.LicenseKey, "mm-license-key", a.MaxMind.LicenseKey, "MaxMind License Key")
	fs.StringVar(&a.MaxMind.UserID, "mm-user-id", a.MaxMind.UserID, "MaxMind User ID (requires license-key)")
	fs.StringVar(&a.MaxMind.ProductID, "mm-product-id", a.MaxMind.ProductID, "MaxMind Product ID (e.g GeoLite2-City)")
	fs.DurationVar(&a.MaxMind.RetryInterval, "mm-retry", a.MaxMind.RetryInterval, "Max time to wait before retrying to download a MaxMind database")
	fs.DurationVar(&a.MaxMind.UpdateInterval, "mm-update", a.MaxMind.UpdateInterval, "MaxMind database update check interval")
	fs.StringVar(&a.MaxMind.Downstreams, "mm-downstreams", a.MaxMind.Downstreams, "MaxMind Update Downstreams")

	fs.StringVar(&a.Ip2Location.Token, "i2l-token", a.Ip2Location.Token, "ip2location token")
	fs.StringVar(&a.Ip2Location.ProductID, "i2l-product-id", a.Ip2Location.ProductID, "ip2location Product ID (e.g PX8LITEBIN)")
	fs.StringVar(&a.Ip2Location.CsvProductID, "i2l-csv-product-id", a.Ip2Location.CsvProductID, "ip2location CSV Product ID (e.g PX8LITEBIN)")
	fs.DurationVar(&a.Ip2Location.RetryInterval, "i2l-retry", a.Ip2Location.RetryInterval, "Max time to wait before retrying to download a ip2location database")
	fs.DurationVar(&a.Ip2Location.UpdateInterval, "i2l-update", a.Ip2Location.UpdateInterval, "ip2location database update check interval")
	fs.StringVar(&a.Ip2Location.Downstreams, "i2l-updates-host", a.Ip2Location.Downstreams, "ip2location Update Downstreams")

	fs.StringVar(&a.Tor.ExitCheck, "tor-exit-check", a.Tor.ExitCheck, "Tor exit check (e.g 8.8.8.8)")
	fs.DurationVar(&a.Tor.RetryInterval, "tor-retry", a.Tor.RetryInterval, "Max time to wait before retrying to download a tor database")
	fs.DurationVar(&a.Tor.UpdateInterval, "tor-update", a.Tor.UpdateInterval, "Tor database update check interval")
	fs.StringVar(&a.Tor.Downstreams, "tor-updates-host", a.Tor.Downstreams, "Tor Update Downstreams")

	fs.DurationVar(&a.Server.WriteTimeout, "write-timeout", a.Server.WriteTimeout, "Write timeout for HTTP and HTTPS client connections")
	fs.BoolVar(&a.Log.ToStdout, "logtostdout", a.Log.ToStdout, "Log to stdout instead of stderr")
	fs.StringVar(&a.Log.OutputFile, "log-file", a.Log.OutputFile, "Log output file")
	fs.BoolVar(&a.Log.Timestamp, "log-timestamp", a.Log.Timestamp, "Prefix non-access logs with timestamp")

	fs.StringVar(&a.DatabaseLocation, "db", a.DatabaseLocation, "Database file location")

	fs.IntVar(&a.Server.RateLimit.Burst, "quota-burst", a.Server.RateLimit.Burst, "Max requests per source IP per request burst")
	fs.Float64Var((*float64)(&a.Server.RateLimit.Limit), "quota-max", float64(a.Server.RateLimit.Limit), "Max requests per source IP per interval; set 0 to turn quotas off")

	fs.DurationVar(&a.Server.ReadTimeout, "read-timeout", a.Server.ReadTimeout, "Read timeout for HTTP and HTTPS client connections")

	fs.BoolVar(&a.Log.Enabled, "silent", a.Log.Enabled, "Disable HTTP and HTTPS log request details")
	fs.StringVar(&a.File, "config", a.File, "Config file")

	sv := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *sv {
		fmt.Printf("geoIP version: %s\n", a.Build.Version)
		fmt.Printf("geoIP build number: %s\n", a.Build.Number)
		os.Exit(0)
	}
}

func (a *Application) Start() error {
	a.ready = false
	if err := a.connect(); err != nil {
		return err
	}

	a.importQueue = make(chan *Record, a.queueSize)

	a.workers = make([]*Worker, a.numWorkers)
	for i := 0; i < a.numWorkers; i++ {
		a.workers[i] = NewWorker(a)
	}

	a.MaxMind.Start(a.RootDir, func(resp *api.MaxMindCsvRecord) {
		_, ipnetA, _ := net.ParseCIDR(resp.Network)
		mask := binary.BigEndian.Uint32(ipnetA.Mask)
		start := binary.BigEndian.Uint32(ipnetA.IP)

		record := &Record{
			ip: ipnetA.IP,
		}
		if _, b := ipnetA.Mask.Size(); b == 32 {
			// IP v4
			// find the final address
			finish := (start & mask) | (mask ^ 0xffffffff)

			// loop through addresses as uint32
			for i := start; i <= finish; i++ {
				// convert back to net.IP
				ip := make(net.IP, 4)
				binary.BigEndian.PutUint32(ip, i)
				record.ip = ip
				a.QueueRecord(record)
			}
		} else {
			// IP v6
			a.QueueRecord(record)
		}
	})
	a.Ip2Location.Start(a.RootDir, func(resp *api.Ip2LocationCsvRecord) {
		for i := resp.IpFrom; i <= resp.IpTo; i++ {
			ip := api.IntToIp(big.NewInt(int64(i)))
			a.QueueRecord(&Record{ip: ip})
		}
	})
	a.Tor.Start(a.RootDir)

	ticker := time.NewTicker(1 * time.Second)
	a.close = make(chan bool)

	go func() {

		for {
			a.mx.RLock()
			if a.ready == false {
				a.mx.RUnlock()
				time.Sleep(time.Second * 1)
			} else {
				a.mx.RUnlock()
				break
			}
		}

		for {
			select {
			case record := <-a.importQueue:
				a.dispatchRecord(record)
			case <-a.close:
				return
			}
		}
	}()
	go func() {
		t := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-t.C:
				a.GenerateStatistic()
			case <-a.close:
				t.Stop()
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			if a.ready == false {
				a.mx.Lock()
				a.ready = a.MaxMind.Ready() && a.Ip2Location.Ready() && a.Tor.Ready()
				if a.ready == true {
					if err := a.Server.Start(a.db, &a.Log, a.routerFunc); err != nil {
						a.mx.Unlock()
						return err
					}
					a.statistic = a.generateStatistic()
					ticker.Stop()
				}
				a.mx.Unlock()
			}
		case <-a.close:
			ticker.Stop()
			return nil
		}
	}
}

func (a *Application) Stop() error {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.MaxMind.Stop()
	a.Ip2Location.Stop()
	a.Tor.Stop()
	a.Server.Stop()
	a.close <- true
	a.ready = false
	return nil
}

func (a *Application) GenerateStatistic() {
	if a.Ready() {
		result := a.generateStatistic()

		a.mx.Lock()
		a.statistic = result
		a.mx.Unlock()
	}
}

func (a *Application) generateStatistic() *StatisticResponse {
	result := &StatisticResponse{
		Ips:       0,
		Cities:    0,
		Countries: 0,
		Domains:   0,
		Isps:      0,
		Asns:      0,
		Networks:  0,
	}

	a.db.Model(&models.IP{}).Count(&result.Ips)
	a.db.Model(&models.City{}).Count(&result.Cities)
	a.db.Model(&models.Country{}).Count(&result.Countries)
	a.db.Model(&models.Domain{}).Count(&result.Domains)
	a.db.Model(&models.ISP{}).Count(&result.Isps)
	a.db.Model(&models.AutonomousSystem{}).Count(&result.Asns)
	a.db.Model(&models.Network{}).Count(&result.Networks)

	return result
}

func (a *Application) connect() error {
	db, err := gorm.Open(sqlite.Open(a.DatabaseLocation), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return err
	}
	a.db = db

	// Migrate the schema
	_ = a.db.AutoMigrate(&models.Region{})
	_ = a.db.AutoMigrate(&models.AutonomousSystem{})
	_ = a.db.AutoMigrate(&models.IP{})

	return nil
}

func (a *Application) dispatchRecord(record *Record) {
	for {
		for _, w := range a.workers {
			if w.IsIdle() {
				w.Work(record)
				return
			}
		}
	}
}

func (a *Application) QueueRecord(record *Record) {
	if a.Ready() == false {
		for {
			a.mx.RLock()
			if a.ready == false {
				a.mx.RUnlock()
				time.Sleep(time.Second * 1)
			} else {
				a.mx.RUnlock()
				break
			}
		}
	}

	record.mm = a.MaxMind.Lookup(record.ip)
	record.mma = a.MaxMind.LookupASN(record.ip)
	record.i2l = a.Ip2Location.Lookup(record.ip)

	if (record.mm.Traits.AutonomousSystemNumber <= 0 && record.mma.AutonomousSystemNumber <= 0 &&
		(record.i2l.ASN == "" || record.i2l.ASN == "0")) || (record.mm.Traits.ISP == "" && record.i2l.ISP == "") {
		record.ispl, _ = ipisp.LookupIP(context.Background(), record.ip)
	}

	a.importQueue <- record
}

func (a *Application) routerFunc(h *httpmux.Handler) {
	h.Handle("GET", "/api/country/*country", a.Server.RegisterHandler(a.apiCountry))
	h.Handle("GET", "/api/useragent", a.Server.RegisterHandler(a.apiUserAgent))
	h.Handle("GET", "/api/language", a.Server.RegisterHandler(a.apiLanguage))
	h.Handle("GET", "/api/me", a.Server.RegisterHandler(a.apiMe))
	h.Handle("GET", "/api/statistic", a.Server.RegisterHandler(a.apiStatistic))
	h.Handle("GET", "/api/detail/*host", a.Server.RegisterHandler(a.apiDetail))

	h.Handle("POST", "/api/search", a.Server.RegisterHandler(a.apiSearch))

	// Deprecated request - supports v1 requests
	h.Handle("GET", "/json/*host", a.Server.RegisterHandler(a.apiV1Detail))

	// Serve static files
	serveAssets(h, a.assets, "/", "static")
}

func (a *Application) Ready() bool {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return a.ready
}

func sendJSON(w http.ResponseWriter, r *http.Request, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

}

func serveAssets(h *httpmux.Handler, assets embed.FS, root, virtual string) {

	hc, err := fs.Sub(fs.FS(assets), "static")
	if err != nil {
		fmt.Printf("[error] failed to load asset: %s", err.Error())
		return
	}

	h.Handle("GET", "/", http.FileServer(http.FS(hc)))

	items, _ := assets.ReadDir(virtual)
	for _, d := range items {
		if d.IsDir() {
			htmlContent, err := fs.Sub(fs.FS(assets), path.Join(virtual, d.Name()))
			if err != nil {
				fmt.Printf("[error] failed to load asset: %s", err.Error())
				continue
			}

			h.HandleFunc("GET", path.Join(root, d.Name())+"/*filepath", func(w http.ResponseWriter, r *http.Request) {
				filepath := httpmux.Params(r).ByName("filepath")
				if filepath != "" {
					c, err := htmlContent.Open(filepath[1:])
					if err != nil {
						fmt.Printf("[error] %s\n", err.Error())
						return
					}
					buf := make([]byte, 1024)

					if strings.HasSuffix(filepath, "css") {
						w.Header().Set("content-type", "text/css")
					} else if strings.HasSuffix(filepath, "js") {
						w.Header().Set("content-type", "text/javascript")
					} else if strings.HasSuffix(filepath, "png") {
						w.Header().Set("content-type", "image/png")
					} else if strings.HasSuffix(filepath, "ico") {
						w.Header().Set("content-type", "image/ico")
					}

					w.WriteHeader(http.StatusOK)

					for {
						// read a chunk
						n, err := c.Read(buf)
						if err != nil && err != io.EOF {
							break
						}
						if n == 0 {
							break
						}

						// write a chunk
						if _, err := w.Write(buf[:n]); err != nil {
							break
						}
					}
				}
			})
		}
	}
}

func getParam(r *http.Request, name string) string {
	host := httpmux.Params(r).ByName(name)
	if len(host) > 0 && host[0] == '/' {
		host = host[1:]
	}
	if strings.Contains(host, "?") {
		host = strings.Split(host, "?")[0]
	}
	if host == "" && name == "host" {
		host, _, _ = net.SplitHostPort(r.RemoteAddr)
		if host == "" {
			host = r.RemoteAddr
		}
	}
	return host
}

type point struct {
	x float64
	y float64
}

func toRadians(p float64) float64 {
	return p * (math.Pi / 180)
}

func toDegrees(p float64) float64 {
	return p * (180 / math.Pi)
}

func calculateDerivedPosition(p *point, _range, bearing float64) *point {
	earthRadius := 3959.0
	latA := toRadians(p.x)
	lonA := toRadians(p.y)
	angularDistance := _range / earthRadius
	trueCourse := toRadians(bearing)

	lat := math.Asin(math.Sin(latA)*math.Cos(angularDistance) + math.Cos(latA)*math.Sin(angularDistance)*math.Cos(trueCourse))
	dlon := math.Atan2(math.Sin(trueCourse)*math.Sin(angularDistance)*math.Cos(latA), math.Cos(angularDistance)-math.Sin(latA)*math.Sin(lat))
	lon := math.Mod(lonA+dlon+math.Pi, math.Pi*2) - math.Pi

	return &point{
		x: toDegrees(lat),
		y: toDegrees(lon),
	}
}
