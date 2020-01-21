package setup

import (
	"bufio"
	"fmt"
	"../utils/config"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Setup struct {
	reader *bufio.Reader
}

func (s *Setup) askForInput(question string, defaultString string, defaultResponse string) string {
	fmt.Printf("%s? [%s]\n", question, defaultString)
	response, _ := s.reader.ReadString('\n')
	response = strings.Replace(response, "\n", "", -1)
	if len(response) == 0 {
		return defaultResponse
	}
	return response
}

func Stob(s string) bool {
	var response bool
	if strings.ToLower(s) == "y" {response = true}
	return response
}

func Stohd(s string) time.Duration {
	hours, _ := strconv.Atoi(s)
	return time.Duration(hours) * time.Hour
}

func Stomd(s string) time.Duration {
	minutes, _ := strconv.Atoi(s)
	return time.Duration(minutes) * time.Minute
}

func Stosd(s string) time.Duration {
	seconds, _ := strconv.Atoi(s)
	return time.Duration(seconds) * time.Second
}

func RunSetup(c *config.Config){
	fmt.Println("")
	fmt.Println("Go GeoIP Webapi Setup")
	fmt.Println("Version: " + c.Build.Version)
	fmt.Println("")
	fmt.Println("Starting setup..")

	s := &Setup{
		reader: bufio.NewReader(os.Stdin),
	}

	p := strings.Split(c.ServerAddr, ":")
	hostAddr := s.askForInput("Host address", "Default: " + p[0], p[0])
	hostPort := s.askForInput("Host port", "Default: " + p[1], p[1])
	if Stob(s.askForInput("Enable HTTPS", "y/N", "n")) {
		c.TLSServerAddr = hostAddr + ":" + hostPort

		enableHttp2 := s.askForInput("Enable HTTP2", "Y/n", "y")
		c.HTTP2 = Stob(enableHttp2)

		c.HSTS = s.askForInput("HSTS option", "Default: " + c.HSTS, c.HSTS)

		c.LetsEncrypt = Stob(s.askForInput("Enable letsencrypt", "Y/n", "y"))
		if c.LetsEncrypt {
			c.LetsEncryptEmail = s.askForInput("Email to register with letsencrypt", "Default: " + c.LetsEncryptEmail, c.LetsEncryptEmail)
			c.LetsEncryptHosts = s.askForInput("Comma separated list of hosts for the certificate", "Default: " + c.LetsEncryptHosts, c.LetsEncryptHosts)
			c.LetsEncryptCacheDir = s.askForInput("Letsencrypt cert dir", "Default: " + c.LetsEncryptCacheDir, c.LetsEncryptCacheDir)
		}else{
			c.TLSCertFile = s.askForInput("X.509 certificate file", "Default: " + c.TLSCertFile, c.TLSCertFile)
			c.TLSKeyFile = s.askForInput("X.509 key file", "Default: " + c.TLSCertFile, c.TLSCertFile)
		}
	}else{
		c.ServerAddr = hostAddr + ":" + hostPort
	}

	c.FastOpen = Stob(s.askForInput("Enable TCP fast open", "y/N", "n"))
	c.Naggle = Stob(s.askForInput("Enable TCP Nagle's algorithm (disables NO_DELAY)", "y/N", "n"))

	c.APIPrefix = s.askForInput("API endpoint prefix", "Default: " + c.APIPrefix, c.APIPrefix)
	c.CORSOrigin = s.askForInput("Comma separated list of CORS origins endpoints", "Default: " + c.CORSOrigin, c.CORSOrigin)
	c.UseXForwardedFor = Stob(s.askForInput("Use the X-Forwarded-For header when available (e.g. behind proxy)", "y/N", "n"))
	c.GuiDir = s.askForInput("Web gui directory", "leave empty to disable", "")

	availableRateLimitBackends := []string{"map", "redis", "memcache"}
	for {
		c.RateLimitBackend = s.askForInput("Backend for rate limiter: map, redis, or memcache", "Default: " + c.RateLimitBackend, c.RateLimitBackend)
		if i := sort.SearchStrings(availableRateLimitBackends, c.RateLimitBackend); len(c.RateLimitBackend) > 0 && i >= 0 {break}
	}
	switch c.RateLimitBackend {
	case "memcache":
		p := strings.Split(c.MemcacheAddr, ":")
		hostAddr := s.askForInput("Memcache host address", "Default: " + p[0], p[0])
		hostPort := s.askForInput("Memcache host port", "Default: " + p[1], p[1])
		c.MemcacheAddr = hostAddr + ":" + hostPort

		memcacheTimeout := strconv.Itoa(int(c.MemcacheTimeout.Seconds()))
		c.MemcacheTimeout = Stosd(s.askForInput("Memcache read/write timeout in seconds", "Default: " + memcacheTimeout, memcacheTimeout))
		break
	case "redis":
		p := strings.Split(c.RedisAddr, ":")
		hostAddr := s.askForInput("Redis host address", "Default: " + p[0], p[0])
		hostPort := s.askForInput("Redis host port", "Default: " + p[1], p[1])
		c.RedisAddr = hostAddr + ":" + hostPort

		redisTimeout := strconv.Itoa(int(c.RedisTimeout.Seconds()))
		c.RedisTimeout = Stosd(s.askForInput("Redis read/write timeout in seconds", "Default: " + redisTimeout, redisTimeout))
		break
	}

	rateLimitBurst := strconv.Itoa(c.RateLimitBurst)
	rateLimitLimit := strconv.Itoa(int(c.RateLimitLimit))
	rateLimitInterval := strconv.Itoa(int(c.RateLimitInterval.Minutes()))

	c.RateLimitBurst, _ = strconv.Atoi(s.askForInput("Max requests per source IP per request burst", "Default: " + rateLimitBurst, rateLimitBurst))
	c.RateLimitLimit, _ = strconv.Atoi(s.askForInput("Max requests per source IP per interval; set 0 to turn quotas off", "Default: " + rateLimitLimit, rateLimitLimit))
	c.RateLimitInterval = Stomd(s.askForInput("Quota expiration interval in minutes, per source IP querying the API", "Default: " + rateLimitInterval, rateLimitInterval))

	mmRetryInterval := strconv.Itoa(int(c.MMRetryInterval.Hours()))
	mmUpdateInterval := strconv.Itoa(int(c.MMUpdateInterval.Hours()))
	c.MMLicenseKey = s.askForInput("MaxMind License Key", "Default: " + c.MMLicenseKey, c.MMLicenseKey)
	c.MMUserID = s.askForInput("MaxMind User ID", "Default: " + c.MMUserID, c.MMUserID)
	c.MMProductID = s.askForInput("MaxMind Product ID (e.g GeoLite2-City)", "Default: " + c.MMProductID, c.MMProductID)
	c.MMUpdatesHost = s.askForInput("MaxMind Updates Host", "Default: " + c.MMUpdatesHost, c.MMUpdatesHost)
	c.MMRetryInterval = Stohd(s.askForInput("Max time in hours to wait before retrying to download a MaxMind database", "Default: " + mmRetryInterval, mmRetryInterval))
	c.MMUpdateInterval = Stohd(s.askForInput("MaxMind database update in hours check interval", "Default: " + mmUpdateInterval, mmUpdateInterval))

	ip2RetryInterval := strconv.Itoa(int(c.MMRetryInterval.Hours()))
	ip2UpdateInterval := strconv.Itoa(int(c.MMUpdateInterval.Hours()))
	c.I2LToken = s.askForInput("ip2location token", "Default: " + c.I2LToken, c.I2LToken)
	c.I2LProductID = s.askForInput("ip2location Product ID", "Default: " + c.I2LProductID, c.I2LProductID)
	c.I2LUpdatesHost = s.askForInput("ip2location Updates Host", "Default: " + c.I2LUpdatesHost, c.I2LUpdatesHost)
	c.I2LRetryInterval = Stohd(s.askForInput("Max time to wait before retrying to download a ip2location database", "Default: " + ip2RetryInterval, ip2RetryInterval))
	c.I2LUpdateInterval = Stohd(s.askForInput("ip2location database update check interval", "Default: " + ip2UpdateInterval, ip2UpdateInterval))

	torRetryInterval := strconv.Itoa(int(c.MMRetryInterval.Hours()))
	torUpdateInterval := strconv.Itoa(int(c.MMUpdateInterval.Hours()))
	c.TorExitCheck = s.askForInput("Tor exit check", "Default: " + c.TorExitCheck, c.TorExitCheck)
	c.TorUpdatesHost = s.askForInput("Tor Updates Host", "Default: " + c.TorUpdatesHost, c.TorUpdatesHost)
	c.TorRetryInterval = Stohd(s.askForInput("Max time to wait before retrying to download the tor database", "Default: " + torRetryInterval, torRetryInterval))
	c.TorUpdateInterval = Stomd(s.askForInput("ip2location database update check interval", "Default: " + torUpdateInterval, torUpdateInterval))

	writeTimeout := strconv.Itoa(int(c.WriteTimeout.Seconds()))
	readTimeout := strconv.Itoa(int(c.ReadTimeout.Seconds()))
	c.WriteTimeout = Stosd(s.askForInput("Write timeout for HTTP and HTTPS client connections", "Default: " + writeTimeout, writeTimeout))
	c.ReadTimeout = Stosd(s.askForInput("Read timeout for HTTP and HTTPS client connections", "Default: " + readTimeout, readTimeout))

	c.LogToStdout = Stob(s.askForInput("Log to stdout instead of stderr", "y/N", "n"))
	c.LogTimestamp = Stob(s.askForInput("Prefix non-access logs with timestamp", "Y/n", "y"))

	fmt.Println("")
	if ok, err := c.Save(); err != nil || !ok {
		fmt.Println("ERROR! Failed to save the config file")
		os.Exit(1)
	}
	fmt.Println("Setup completed")

}
