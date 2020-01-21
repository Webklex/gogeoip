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

		MMUserID:    		"",
		MMLicenseKey:    	"",
		MMProductID:        "GeoLite2-City",
		MMASNProductID:     "GeoLite2-ASN",
		MMUpdatesHost:      "download.maxmind.com",
		MMUpdateInterval:   4 * time.Hour,
		MMRetryInterval:    2 * time.Hour,

		I2LToken:    		"",
		I2LProductID:       "PX8LITEBIN",
		I2LUpdatesHost:     "www.ip2location.com",
		I2LUpdateInterval:  4 * time.Hour,
		I2LRetryInterval:   2 * time.Hour,

		TorUpdatesHost:     "check.torproject.org",
		TorUpdateInterval:  30 * time.Minute,
		TorRetryInterval:   2 * time.Hour,
		TorExitCheck:		"8.8.8.8",

		APIPrefix:           "/",
		CORSOrigin:          "*",
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        15 * time.Second,
		RateLimitLimit: 	 1,
		RateLimitBurst: 	 3,
		LogTimestamp:        true,
		RedisAddr:           "localhost:6379",
		RedisTimeout:        time.Second,
		MemcacheAddr:        "localhost:11211",
		MemcacheTimeout:     time.Second,
		RateLimitBackend:    "redis",
		RateLimitInterval:   3 * time.Minute,

		RootDir: dir,
		File: path.Join(dir, "conf", "settings.config"),
		SaveConfigFlag: false,
		Silent: true,
	}
}
// AddFlags adds configuration flags to the given FlagSet.
func (c *Config) AddFlags(fs *flag.FlagSet) {
	defer envconfig.Process("geoip", c)

	fs.StringVar(&c.ServerAddr, 	"http", 	c.ServerAddr, 		"Address in form of ip:port to listen")
	fs.StringVar(&c.TLSServerAddr, 	"https", 	c.TLSServerAddr, 	"Address in form of ip:port to listen")

	fs.BoolVar(&c.FastOpen, "tcp-fast-open", 	c.FastOpen, "Enable TCP fast open")
	fs.BoolVar(&c.Naggle, 	"tcp-naggle", 	c.Naggle, 	"Enable TCP Nagle's algorithm (disables NO_DELAY)")
	fs.BoolVar(&c.HTTP2, 	"http2", 			c.HTTP2, 	"Enable HTTP/2 when TLS is enabled")

	fs.StringVar(&c.HSTS, 			"hsts", 	c.HSTS, 		"Set HSTS to the value provided on all responses")
	fs.StringVar(&c.TLSCertFile, 	"cert", 	c.TLSCertFile, 	"X.509 certificate file for HTTPS server")
	fs.StringVar(&c.TLSKeyFile, 	"key", 	c.TLSKeyFile, 	"X.509 key file for HTTPS server")

	fs.BoolVar(&c.LetsEncrypt, 				"letsencrypt", 			c.LetsEncrypt, 			"Enable automatic TLS using letsencrypt.org")
	fs.StringVar(&c.LetsEncryptEmail, 		"letsencrypt-email", 		c.LetsEncryptEmail, 	"Optional email to register with letsencrypt (default is anonymous)")
	fs.StringVar(&c.LetsEncryptHosts, 		"letsencrypt-hosts", 		c.LetsEncryptHosts, 	"Comma separated list of hosts for the certificate (required)")
	fs.StringVar(&c.LetsEncryptCacheDir, 	"letsencrypt-cert-dir", 	c.LetsEncryptCacheDir, 	"Letsencrypt cert dir")

	fs.StringVar(&c.APIPrefix, 		"api-prefix", 			c.APIPrefix, 		"API endpoint prefix")
	fs.StringVar(&c.CORSOrigin, 	"cors-origin", 			c.CORSOrigin, 		"Comma separated list of CORS origins endpoints")
	fs.BoolVar(&c.UseXForwardedFor, "use-x-forwarded-for", 	c.UseXForwardedFor, "Use the X-Forwarded-For header when available (e.g. behind proxy)")

	fs.StringVar(&c.GuiDir, "gui", c.GuiDir, "Web gui directory")

	fs.StringVar(&c.MMLicenseKey, 		"mm-license-key",		c.MMLicenseKey,		"MaxMind License Key")
	fs.StringVar(&c.MMUserID, 			"mm-user-id",			c.MMUserID,			"MaxMind User ID (requires license-key)")
	fs.StringVar(&c.MMProductID, 		"mm-product-id",		c.MMProductID,		"MaxMind Product ID (e.g GeoLite2-City)")
	fs.DurationVar(&c.MMRetryInterval, 	"mm-retry",			c.MMRetryInterval,	"Max time to wait before retrying to download a MaxMind database")
	fs.DurationVar(&c.MMUpdateInterval, "mm-update",			c.MMUpdateInterval,	"MaxMind database update check interval")
	fs.StringVar(&c.MMUpdatesHost, 		"mm-updates-host",	c.MMUpdatesHost,	"MaxMind Updates Host")

	fs.StringVar(&c.I2LToken, 				"i2l-token",			c.I2LToken,				"ip2location token")
	fs.StringVar(&c.I2LProductID, 			"i2l-product-id",		c.I2LProductID,			"ip2location Product ID (e.g PX8LITEBIN)")
	fs.DurationVar(&c.I2LRetryInterval, 	"i2l-retry",			c.I2LRetryInterval,		"Max time to wait before retrying to download a ip2location database")
	fs.DurationVar(&c.I2LUpdateInterval, 	"i2l-update",			c.I2LUpdateInterval,	"ip2location database update check interval")
	fs.StringVar(&c.I2LUpdatesHost, 		"i2l-updates-host",	c.I2LUpdatesHost,		"ip2location Updates Host")

	fs.StringVar(&c.TorExitCheck, 			"tor-exit-check",		c.TorExitCheck,			"Tor exit check (e.g 8.8.8.8)")
	fs.DurationVar(&c.TorRetryInterval, 	"tor-retry",			c.I2LRetryInterval,		"Max time to wait before retrying to download a tor database")
	fs.DurationVar(&c.TorUpdateInterval, 	"tor-update",			c.I2LUpdateInterval,	"Tor database update check interval")
	fs.StringVar(&c.TorUpdatesHost, 		"tor-updates-host",	c.I2LUpdatesHost,		"Tor Updates Host")

	fs.DurationVar(&c.WriteTimeout, 	"write-timeout", 		c.WriteTimeout, 	"Write timeout for HTTP and HTTPS client connections")
	fs.BoolVar(&c.LogToStdout, 			"logtostdout", 		c.LogToStdout, 		"Log to stdout instead of stderr")
	fs.StringVar(&c.LogOutputFile, 			"log-file", 		c.LogOutputFile, 		"Log output file")
	fs.BoolVar(&c.LogTimestamp, 		"logtimestamp", 		c.LogTimestamp, 	"Prefix non-access logs with timestamp")
	fs.StringVar(&c.MemcacheAddr, 		"memcache", 			c.MemcacheAddr, 	"Memcache address in form of host:port[,host:port] for quota")
	fs.DurationVar(&c.MemcacheTimeout, 	"memcache-timeout", 	c.MemcacheTimeout, 	"Memcache read/write timeout")

	fs.StringVar(&c.RateLimitBackend, 		"quota-backend", 	c.RateLimitBackend, 	"Backend for rate limiter: map, redis, or memcache")
	fs.IntVar(&c.RateLimitBurst, 			"quota-burst", 	c.RateLimitBurst, 		"Max requests per source IP per request burst")
	fs.DurationVar(&c.RateLimitInterval, 	"quota-interval", c.RateLimitInterval, 	"Quota expiration interval, per source IP querying the API")
	fs.IntVar(&c.RateLimitLimit, 		   "quota-max", 		c.RateLimitLimit, 		"Max requests per source IP per interval; set 0 to turn quotas off")

	fs.DurationVar(&c.ReadTimeout, 	"read-timeout",	c.ReadTimeout, 	"Read timeout for HTTP and HTTPS client connections")
	fs.StringVar(&c.RedisAddr, 		"redis", 			c.RedisAddr, 	"Redis address in form of host:port[,host:port] for quota")
	fs.DurationVar(&c.RedisTimeout, "redis-timeout", 	c.RedisTimeout, "Redis read/write timeout")

	fs.BoolVar(&c.Silent, 			"silent", c.Silent, 			"Disable HTTP and HTTPS log request details")
	fs.StringVar(&c.File, 			"config", c.File, 			"Config file")
	fs.BoolVar(&c.SaveConfigFlag, 	"save", 	c.SaveConfigFlag, 	"Save config")
	fs.BoolVar(&c.RunSetupFlag, 	"setup", 	c.RunSetupFlag, 	"Run the setup wizard")
}


func NewConfigFromFile(configFile string) *Config {
	config := DefaultConfig()

	if configFile != "" {
		config.Load(configFile)
		config.File = configFile
	}

	return config
}

func CreateDirectory(dirName string) bool {
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
	CreateDirectory("conf")
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

			if !c.Silent {
				log.Printf("[error] Config file failed to load: %s", err.Error())
			}
			return false
		}

		err = json.Unmarshal(content, c)
		if err != nil {
			if !c.Silent {
				log.Printf("[error] Config file failed to load: %s", err.Error())
			}
			return false
		}

		if !c.Silent {
			log.Printf("[info] Config file loaded successfully")
		}

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
		if !c.Silent {
			fmt.Println(err)
		}
		return false, err
	}

	err = ioutil.WriteFile(c.File, file, 0644)
	if err != nil {
		panic(err)
		return false, err
	}

	if !c.Silent {
		log.Printf("[info] Config file saved under: %s", c.File)
	}

	return true, nil
}

func (c *Config) logWriter() io.Writer {
	return c.LogOutput
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