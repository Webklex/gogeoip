package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

func NewConfig() *Config {
	return &Config{}
}

func DefaultConfig() *Config {
	dir, _ := os.Getwd()

	return &Config{
		FastOpen:            false,
		Naggle:              false,
		ServerAddr:          "localhost:8080",
		HTTP2:               true,
		HSTS:                "",
		TLSCertFile:         "cert.pem",
		TLSKeyFile:          "key.pem",
		LetsEncrypt:         false,
		LetsEncryptCacheDir: ".",
		LetsEncryptEmail:    "",
		LetsEncryptHosts:    "",
		UserID:    			 "",
		LicenseKey:    		 "",
		APIPrefix:           "/",
		CORSOrigin:          "*",
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        15 * time.Second,
		UpdateInterval:      24 * time.Hour,
		RetryInterval:       2 * time.Hour,
		RateLimitLimit: 	 1,
		RateLimitBurst: 	 3,
		LogTimestamp:        true,
		RedisAddr:           "localhost:6379",
		RedisTimeout:        time.Second,
		MemcacheAddr:        "localhost:11211",
		MemcacheTimeout:     time.Second,
		RateLimitBackend:    "redis",
		RateLimitInterval:   time.Hour,
		UpdatesHost:         "download.maxmind.com",

		// https://www.maxmind.com/en/accounts/{UserID}/geoip/downloads?direct=1
		ProductID:           "GeoLite2-City",

		RootDir: dir,
		File: path.Join(dir, "conf", "settings.config"),
		SaveConfigFlag: false,
	}
}
// AddFlags adds configuration flags to the given FlagSet.
func (c *Config) AddFlags(fs *flag.FlagSet) {
	defer envconfig.Process("geoip", c)
	fs.StringVar(&c.APIPrefix, "api-prefix", c.APIPrefix, "API endpoint prefix")
	fs.StringVar(&c.TLSCertFile, "cert", c.TLSCertFile, "X.509 certificate file for HTTPS server")
	fs.StringVar(&c.CORSOrigin, "cors-origin", c.CORSOrigin, "Comma separated list of CORS origins endpoints")
	fs.StringVar(&c.GuiDir, "gui", c.GuiDir, "Web gui directory")
	fs.StringVar(&c.HSTS, "hsts", c.HSTS, "Set HSTS to the value provided on all responses")
	fs.StringVar(&c.ServerAddr, "http", c.ServerAddr, "Address in form of ip:port to listen")
	fs.BoolVar(&c.HTTP2, "http2", c.HTTP2, "Enable HTTP/2 when TLS is enabled")
	fs.StringVar(&c.TLSServerAddr, "https", c.TLSServerAddr, "Address in form of ip:port to listen")
	fs.StringVar(&c.TLSKeyFile, "key", c.TLSKeyFile, "X.509 key file for HTTPS server")
	fs.BoolVar(&c.LetsEncrypt, "letsencrypt", c.LetsEncrypt, "Enable automatic TLS using letsencrypt.org")
	fs.StringVar(&c.LetsEncryptEmail, "letsencrypt-email", c.LetsEncryptEmail, "Optional email to register with letsencrypt (default is anonymous)")
	fs.StringVar(&c.LetsEncryptHosts, "letsencrypt-hosts", c.LetsEncryptHosts, "Comma separated list of hosts for the certificate (required)")
	fs.StringVar(&c.LetsEncryptCacheDir, "letsencrypt-cert-dir", c.LetsEncryptCacheDir, "Letsencrypt cert dir")
	fs.StringVar(&c.LicenseKey, "license-key", c.LicenseKey, "MaxMind License Key")
	fs.BoolVar(&c.LogToStdout, "logtostdout", c.LogToStdout, "Log to stdout instead of stderr")
	fs.BoolVar(&c.LogTimestamp, "logtimestamp", c.LogTimestamp, "Prefix non-access logs with timestamp")
	fs.StringVar(&c.MemcacheAddr, "memcache", c.MemcacheAddr, "Memcache address in form of host:port[,host:port] for quota")
	fs.DurationVar(&c.MemcacheTimeout, "memcache-timeout", c.MemcacheTimeout, "Memcache read/write timeout")
	fs.StringVar(&c.ProductID, "product-id", c.ProductID, "MaxMind Product ID (e.g GeoLite2-City)")
	fs.StringVar(&c.RateLimitBackend, "quota-backend", c.RateLimitBackend, "Backend for rate limiter: map, redis, or memcache")
	fs.IntVar(&c.RateLimitBurst, "quota-burst", c.RateLimitBurst, "Max requests per source IP per request burst")
	fs.DurationVar(&c.RateLimitInterval, "quota-interval", c.RateLimitInterval, "Quota expiration interval, per source IP querying the API")
	fs.Uint64Var(&c.RateLimitLimit, "quota-max", c.RateLimitLimit, "Max requests per source IP per interval; set 0 to turn quotas off")
	fs.DurationVar(&c.ReadTimeout, "read-timeout", c.ReadTimeout, "Read timeout for HTTP and HTTPS client connections")
	fs.StringVar(&c.RedisAddr, "redis", c.RedisAddr, "Redis address in form of host:port[,host:port] for quota")
	fs.DurationVar(&c.RedisTimeout, "redis-timeout", c.RedisTimeout, "Redis read/write timeout")
	fs.DurationVar(&c.RetryInterval, "retry", c.RetryInterval, "Max time to wait before retrying to download database")
	fs.BoolVar(&c.SaveConfigFlag, "save", c.SaveConfigFlag, "Save config")
	fs.BoolVar(&c.Silent, "silent", c.Silent, "Disable HTTP and HTTPS log request details")
	fs.BoolVar(&c.FastOpen, "tcp-fast-open", c.FastOpen, "Enable TCP fast open")
	fs.BoolVar(&c.Naggle, "tcp-naggle", c.Naggle, "Enable TCP Nagle's algorithm (disables NO_DELAY)")
	fs.DurationVar(&c.UpdateInterval, "update", c.UpdateInterval, "Database update check interval")
	fs.StringVar(&c.UpdatesHost, "updates-host", c.UpdatesHost, "MaxMind Updates Host")
	fs.BoolVar(&c.UseXForwardedFor, "use-x-forwarded-for", c.UseXForwardedFor, "Use the X-Forwarded-For header when available (e.g. behind proxy)")
	fs.StringVar(&c.UserID, "user-id", c.UserID, "MaxMind User ID (requires license-key)")
	fs.DurationVar(&c.WriteTimeout, "write-timeout", c.WriteTimeout, "Write timeout for HTTP and HTTPS client connections")

}


func NewConfigFromFile(configFile string) *Config {
	config := DefaultConfig()

	if configFile != "" {
		config.Load(configFile)
		config.File = configFile
	}

	return config
}

func createDirectory(dirName string) bool {
	src, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirName, 0755)
		if errDir != nil {
			panic(err)
		}
		return true
	}

	if src.Mode().IsRegular() {
		return false
	}

	return false
}

func (c *Config) initFile(filename string) {
	createDirectory("conf")
	if len(filename) == 0{
		dir, _ := os.Getwd()
		filename = path.Join(dir, "conf", "settings.config")

		c.Load(filename)
	}
	c.File = filename
}

func (c *Config) Load(filename string) bool {
	c.initFile(filename)

	if _, err := os.Stat(filename); err == nil {

		content, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("[error] Config file failed to load: %s", err.Error())
			return false
		}

		err = json.Unmarshal(content, c)
		if err != nil {
			log.Printf("[error] Config file failed to load: %s", err.Error())
			return false
		}

		log.Printf("[info] Config file loaded successfully")

	} else {
		_, _ = c.Save()
	}
	return true
}

func (c *Config) Save() (bool, error) {
	if len(c.File) == 0 {
		c.initFile("")
	}

	file, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	err = ioutil.WriteFile(c.File, file, 0644)
	if err != nil {
		panic(err)
		return false, err
	}
	log.Printf("[info] Config file saved under: %s", c.File)

	return true, nil
}

func (c *Config) logWriter() io.Writer {
	if c.LogToStdout {
		return os.Stdout
	}
	return os.Stderr
}

func (c *Config) ErrorLogger() *log.Logger {
	if c.LogTimestamp {
		return log.New(c.logWriter(), "[error] ", log.LstdFlags)
	}
	return log.New(c.logWriter(), "[error] ", 0)
}

func (c *Config) AccessLogger() *log.Logger {
	return log.New(c.logWriter(), "[access] ", 0)
}