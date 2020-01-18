package config

import (
	"time"
)

type Build struct {
	Number  string	 `json:"number"`
	Version string	 `json:"version"`
}

type Config struct {
	Build    			Build  		  `json:"build"`

	FastOpen            bool          `json:"TCP_FAST_OPEN"`
	Naggle              bool          `json:"TCP_NAGGLE"`
	ServerAddr          string        `json:"HTTP"`
	HTTP2               bool          `json:"HTTP2"`
	HSTS                string        `json:"HSTS"`
	TLSServerAddr       string        `json:"HTTPS"`
	TLSCertFile         string        `json:"CERT"`
	TLSKeyFile          string        `json:"KEY"`
	LetsEncrypt         bool          `json:"LETSENCRYPT"`
	LetsEncryptCacheDir string        `json:"LETSENCRYPT_CERT_DIR"`
	LetsEncryptEmail    string        `json:"LETSENCRYPT_EMAIL"`
	LetsEncryptHosts    string        `json:"LETSENCRYPT_HOSTS"`
	APIPrefix           string        `json:"API_PREFIX"`
	CORSOrigin          string        `json:"CORS_ORIGIN"`
	ReadTimeout         time.Duration `json:"READ_TIMEOUT"`
	WriteTimeout        time.Duration `json:"WRITE_TIMEOUT"`
	UpdateInterval      time.Duration `json:"UPDATE_INTERVAL"`
	RetryInterval       time.Duration `json:"RETRY_INTERVAL"`
	UseXForwardedFor    bool          `json:"USE_X_FORWARDED_FOR"`
	Silent              bool          `json:"SILENT"`
	LogToStdout         bool          `json:"LOGTOSTDOUT"`
	LogTimestamp        bool          `json:"LOGTIMESTAMP"`
	RedisAddr           string        `json:"REDIS"`
	RedisTimeout        time.Duration `json:"REDIS_TIMEOUT"`
	MemcacheAddr        string        `json:"MEMCACHE"`
	MemcacheTimeout     time.Duration `json:"MEMCACHE_TIMEOUT"`
	RateLimitBackend    string        `json:"QUOTA_BACKEND"`
	RateLimitInterval   time.Duration `json:"QUOTA_INTERVAL"`
	RateLimitLimit      uint64        `json:"QUOTA_MAX"`
	RateLimitBurst      int           `json:"QUOTA_BURST"`
	UpdatesHost         string        `json:"UPDATES_HOST"`
	LicenseKey          string        `json:"LICENSE_KEY"`
	UserID              string        `json:"USER_ID"`
	ProductID           string        `json:"PRODUCT_ID"`
	ASNProductID        string        `json:"ASN_PRODUCT_ID"`
	GuiDir              string        `json:"GUI"`

	File     			string 		  `json:"-"`
	RootDir     		string 		  `json:"-"`
	SaveConfigFlag     	bool 		  `json:"-"`
}