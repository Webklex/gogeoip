package config

import (
	"os"
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
	UseXForwardedFor    bool          `json:"USE_X_FORWARDED_FOR"`
	Silent              bool          `json:"SILENT"`
	LogToStdout         bool          `json:"LOG_STDOUT"`
	LogOutputFile       string        `json:"LOG_FILE"`
	LogTimestamp        bool          `json:"LOG_TIMESTAMP"`
	RedisAddr           string        `json:"REDIS"`
	RedisTimeout        time.Duration `json:"REDIS_TIMEOUT"`
	MemcacheAddr        string        `json:"MEMCACHE"`
	MemcacheTimeout     time.Duration `json:"MEMCACHE_TIMEOUT"`

	RateLimitBackend    string        `json:"QUOTA_BACKEND"`
	RateLimitInterval   time.Duration `json:"QUOTA_INTERVAL"`
	RateLimitLimit      int           `json:"QUOTA_MAX"`
	RateLimitBurst      int           `json:"QUOTA_BURST"`

	MMUserID            string        `json:"MM_USER_ID"`
	MMLicenseKey        string        `json:"MM_LICENSE_KEY"`
	MMProductID         string        `json:"MM_PRODUCT_ID"`
	MMASNProductID      string        `json:"MM_ASN_PRODUCT_ID"`
	MMUpdatesHost       string        `json:"MM_UPDATES_HOST"`
	MMRetryInterval     time.Duration `json:"MM_RETRY_INTERVAL"`
	MMUpdateInterval    time.Duration `json:"MM_UPDATE_INTERVAL"`

	I2LToken			string		  `json:"I2L_TOKEN"`
	I2LProductID		string		  `json:"I2L_PRODUCT_ID"`
	I2LRetryInterval    time.Duration `json:"I2L_RETRY_INTERVAL"`
	I2LUpdateInterval   time.Duration `json:"I2L_UPDATE_INTERVAL"`
	I2LUpdatesHost      string        `json:"I2L_UPDATES_HOST"`

	TorExitCheck      	string 		  `json:"TOR_EXIT"`
	TorRetryInterval    time.Duration `json:"TOR_RETRY_INTERVAL"`
	TorUpdateInterval   time.Duration `json:"TOR_UPDATE_INTERVAL"`
	TorUpdatesHost      string        `json:"TOR_UPDATES_HOST"`

	GuiDir              string        `json:"GUI"`

	File     			string 		  `json:"-"`
	RootDir     		string 		  `json:"-"`
	SaveConfigFlag     	bool 		  `json:"-"`
	RunSetupFlag     	bool 		  `json:"-"`

	LogOutput        	*os.File 	  `json:"-"`
}