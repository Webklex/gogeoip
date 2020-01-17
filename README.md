# GeoIP Web API
[![Software License][ico-license]](LICENSE.md)

GoGeoIP - a lightweight geoip api written in GO. [Live Demo](https://www.gogeoip.com/)

![geo_ip_web_gui](https://raw.githubusercontent.com/webklex/gogeoip-gui/master/geo_ip_web_gui.jpg)

## Table of Contents
- [Features](#features)
- [Installation](#installation)
  - [GUI](#gui)
- [Configuration](#server-options)
- [Database](#database)
- [Api](#api)
  - [CSV](#csv)
  - [XML](#xml)
  - [JSON](#json)
  - [JSONP](#jsonp)
- [Build](#build)
- [Support](#support)
- [Security](#security)
- [Credits](#credits)
- [License](#license)

### Features
* Serving over HTTPS (TLS) using your own certificates, or provisioned automatically using [LetsEncrypt.org](https://letsencrypt.org)
* [HSTS ready](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) to restrict your browser clients to always use HTTPS
* Configurable read and write timeouts to avoid stale clients consuming server resources
* Reverse proxy ready
* Configurable [CORS](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) to restrict access to specific domains
* Configurable api prefix to serve the API alongside other APIs on the same host
* Optional round trip optimization by enabling [TCP Fast Open](https://en.wikipedia.org/wiki/TCP_Fast_Open)
* Integrated rate limit (quota) for your clients (per client IP) based on requests per time interval; several backends such as in-memory map (for single instance), or redis or memcache for distributed deployments are supported
* Serve the default [GeoLite2 City](https://dev.maxmind.com/geoip/geoip2/geolite2/) free database that is downloaded and updated automatically in background on a configurable schedule, or
* Serve the commercial [GeoIP2 City](https://www.maxmind.com/en/geoip2-city) database from MaxMind, either as a local file that you provide and update periodically (so the server can reload it), or configured to be downloaded periodically using your API key
* Multiple languages are supported (en, ru, es, jp, fr, de)
* Supports Linux, OS X, FreeBSD, and Windows

### Requirements
A Free MaxMind License is required and can be easily obtained:
1. [Sign up for a MaxMind account](https://www.maxmind.com/en/geolite2/signup) (no purchase required)
2. Set your password and [create a license key](https://www.maxmind.com/en/accounts/current/license-key)

### Installation
Download and unpack a fitting [pre-compiled binary](https://github.com/webklex/gogeoip/releases) or build a binary 
yourself by by following the [build](#build) instructions.

Continue by configuring your application:
```bash
geoip \
    -user-id 100000 \
    -license-key 0AAaAaaAa0A0AAaA \
    -http=:8080 \
    -gui gui \
    -save
```

Open a browser and navigate to `http://localhost:8080/` to verify everything is working.

Please take a look at the available [options](#server-options) for further details.

#### GUI
An example GUI can be found under [webklex/gogeoip-gui](https://github.com/Webklex/gogeoip-gui). It is already included 
in the pre-compiled packages found under [releases](https://github.com/webklex/gogeoip/releases).

### Server Options
To see all the available options, use the `-help` option:
```bash
geoip -help
```
You can configure the web server via command line flags or the config file `conf/settings.config`.

| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -api-prefix            | API_PREFIX           | string | /                    | API endpoint prefix                                         |
| -cert                  | CERT                 | string | cert.pem             | X.509 certificate file for HTTPS server                     |
| -cors-origin           | CORS_ORIGIN          | string | *                    | Comma separated list of CORS origins endpoints              |
| -gui                   | GUI                  | string |                      | Web gui directory                                           |
| -hsts                  | HSTS                 | string |                      |                                                             |
| -http                  | HTTP                 | string | localhost:8080       | Address in form of ip:port to listen                        |
| -http2                 | HTTP2                | bool   | true                 | Enable HTTP/2 when TLS is enabled                           |
| -https                 | HTTPS                | string |                      | Address in form of ip:port to listen                        |
| -key                   | KEY                  | string | key.pem              | X.509 key file for HTTPS server                             |
| -letsencrypt           | LETSENCRYPT          | bool   | false                | Enable automatic TLS using letsencrypt.org                  |
| -letsencrypt-email     | LETSENCRYPT_EMAIL    | string |                      | Optional email to register with letsencrypt                 |
| -letsencrypt-hosts     | LETSENCRYPT_HOSTS    | string |                      | Comma separated list of hosts for the certificate           |
| -letsencrypt-cert-dir  | LETSENCRYPT_CERT_DIR | string |                      | Letsencrypt cert dir                                        |
| -license-key           | LICENSE_KEY          | string |                      | MaxMind License Key                                         |
| -logtostdout           | LOGTOSTDOUT          | bool   | false                | Log to stdout instead of stderr                             |
| -logtimestamp          | LOGTIMESTAMP         | bool   | true                 | Prefix non-access logs with timestamp                       |
| -memcache              | MEMCACHE             | string | localhost:11211      | Memcache address in form of host:port[,host:port] for quota |
| -memcache-timeout      | MEMCACHE_TIMEOUT     | int    | 1000000000           | Memcache read/write timeout in nanoseconds                  |
| -product-id            | PRODUCT_ID           | string | GeoLite2-City        | MaxMind Product ID                                          |
| -quota-backend         | QUOTA_BACKEND        | string | redis                | Backend for rate limiter: map, redis, or memcache           |
| -quota-burst           | QUOTA_BURST          | int    | 3                    | Max requests per source IP per request burst                |
| -quota-interval        | QUOTA_INTERVAL       | int    | 3600000000000        | Quota expiration interval, per source IP querying the API in nanoseconds |
| -quota-max             | QUOTA_MAX            | int    | 1                    | "Max requests per source IP per interval; set 0 to turn quotas off |
| -read-timeout          | READ_TIMEOUT         | int    | 30000000000          | Read timeout in nanoseconds for HTTP and HTTPS client connections |
| -redis                 | REDIS                | string | localhost:6379       | Redis address in form of host:port[,host:port] for quota    |
| -redis-timeout         | REDIS_TIMEOUT        | int    | 1000000000           | Redis read/write timeout in nanoseconds                     |
| -retry                 | RETRY_INTERVAL       | int    | 7200000000000        | Max time in nanoseconds to wait before retrying to download database |
| -save                  |                      | bool   | false                | Save config                                                 |
| -silent                | SILENT               | bool   | false                | Disable HTTP and HTTPS log request details                  |
| -tcp-fast-open         | TCP_FAST_OPEN        | bool   | false                | Enable TCP fast open                                        |
| -tcp-naggle            | TCP_NAGGLE           | bool   | false                | Enable TCP Nagle's algorithm                                |
| -update                | UPDATE_INTERVAL      | int    | 86400000000000       | Database update check interval in nanoseconds               |
| -updates-host          | UPDATES_HOST         | string | download.maxmind.com | MaxMind Updates Host                                        |
| -use-x-forwarded-for   | USE_X_FORWARDED_FOR  | bool   | false                | Use the X-Forwarded-For header when available (e.g. behind proxy) |
| -user-id               | USER_ID              | string |                      | MaxMind User ID                                             |
| -version               |                      | bool   | false                | Show version and exit                                       |
| -write-timeout         | WRITE_TIMEOUT        | int    | 15000000000          | Write timeout in nanoseconds for HTTP and HTTPS client connections |

If you're using LetsEncrypt.org to provision your TLS certificates, you have to listen for HTTPS on port 443. Following is an example of the server listening on 2 different ports: http (80) and https (443):
```bash
geoip \
    -user-id 100000 \
    -license-key 0AAaAaaAa0A0AAaA \
    -http=:8080 \
    -https=:8443 \
    -hsts=max-age=31536000 \
    -letsencrypt \
    -letsencrypt-hosts=example.com \
    -gui gui \
    -save
```

```bash
$ cat conf/settings.config
{
    "USER_ID": "100000",
    "LICENSE_KEY": "0AAaAaaAa0A0AAaA",
    "HTTP": ":8080",
    "HTTPS": ":8443",
    "HSTS": "max-age=31536000",
    "LETSENCRYPT": true,
    "LETSENCRYPT_HOSTS": "example.com",
    ...
```

By default, HTTP/2 is enabled over HTTPS. You can disable by passing the `-http2=false` flag.

If the web server is running behind a reverse proxy or load balancer, you have to run it passing the `-use-x-forwarded-for` 
parameter and provide the `X-Forwarded-For` HTTP header in all requests. This is for the geoip web server be able to log the 
client IP, and to perform geolocation lookups when an IP is not provided to the API, e.g. `/json/` (uses client IP) vs `/json/1.2.3.4`.

## Database
The current implementation uses the free [GeoLite2 City](http://dev.maxmind.com/geoip/geoip2/geolite2/) database from MaxMind.
If you have purchased the commercial database from MaxMind, you can point the geoip web server or (Go API, for dev) to the URL 
containing the file, or local file, and the server will use it.
In case of files on disk, you can replace the file with a newer version and the geoip web server will reload it automatically 
in background. If instead of a file you use a URL (the default), we periodically check the URL in background to see if 
there's a new database version available, then download the reload it automatically.

All responses from the geoip API contain the date that the database was downloaded in the X-Database-Date HTTP header.

## API
The API is served by endpoints that encode the response in different formats.
You can pass a different IP or hostname. For example, to lookup the geolocation of `github.com` the server 
resolves the name first, then uses the first IP address available, which might be IPv4 or IPv6:

```bash
curl :8080/json/{ip or hostname}?lang={language}[&user]
```
Same semantics are available for the `/xml/{ip}` and `/csv/{ip}` endpoints.
JSON responses can be encoded as JSONP, by adding the `callback` parameter:

The used default language depends on the present `Accept-Language` header. You can define the used language by 
providing a `lang` parameter containing the two digit country code (en, ru, es, fr, de, jp).

Add the `user` parameter to the end to receive user device specific information. Please see the [JSON example](#json)
for output details.


### CSV
```bash
curl :8080/csv/
```
```
000.000.000.000,Nordamerika,US,USA,NV,,Las Vegas,89129,America/Los_Angeles,36.2473,-115.2821,839,0,20
```

### XML
```bash
curl :8080/xml/
```
```xml
<Response>
    <IP>000.000.000.000</IP>
    <IsInEuropeanUnion>false</IsInEuropeanUnion>
    <ContinentCode>Nordamerika</ContinentCode>
    <CountryCode>US</CountryCode>
    <CountryName>USA</CountryName>
    <RegionCode>NV</RegionCode>
    <RegionName></RegionName>
    <City>Las Vegas</City>
    <ZipCode>89129</ZipCode>
    <TimeZone>America/Los_Angeles</TimeZone>
    <Latitude>36.2473</Latitude>
    <Longitude>-115.2821</Longitude>
    <AccuracyRadius>20</AccuracyRadius>
    <MetroCode>839</MetroCode>
</Response>
```

### JSON
```bash
curl :8080/json/
```
```json
{
  "ip": "000.000.000.000",
  "is_in_european_union": false,
  "continent_code": "Nordamerika",
  "country_code": "US",
  "country_name": "USA",
  "region_code": "NV",
  "region_name": "",
  "city": "Las Vegas",
  "zip_code": "89129",
  "time_zone": "America/Los_Angeles",
  "latitude": 36.2473,
  "longitude": -115.2821,
  "accuracy_radius": 20,
  "metro_code": 839
}
```
```bash
curl :8080/json/?user
```
```json
{
  "ip": "000.000.000.000",
  "is_in_european_union": false,
  "continent_code": "Nordamerika",
  "country_code": "US",
  "country_name": "USA",
  "region_code": "NV",
  "region_name": "",
  "city": "Las Vegas",
  "zip_code": "89129",
  "time_zone": "America/Los_Angeles",
  "latitude": 36.2473,
  "longitude": -115.2821,
  "accuracy_radius": 20,
  "metro_code": 839,
  "user": {
    "language": {
      "language": "en",
      "region": "US",
      "tag": "en-US"
    },
    "system": {
      "os": "Linux",
      "browser": "Ubuntu Chromium",
      "version": "79.0.3945.79",
      "os_version": "x86_64",
      "device": "",
      "mobile": false,
      "tablet": false,
      "desktop": true,
      "bot": false
    }
  }
}
```

### JSONP
```bash
curl :8080/json/?callback=foobar
```
```javascript
foobar({
  "ip": "000.000.000.000",
  "is_in_european_union": false,
  "continent_code": "Nordamerika",
  "country_code": "US",
  "country_name": "USA",
  "region_code": "NV",
  "region_name": "",
  "city": "Las Vegas",
  "zip_code": "89129",
  "time_zone": "America/Los_Angeles",
  "latitude": 36.2473,
  "longitude": -115.2821,
  "accuracy_radius": 20,
  "metro_code": 839
});
```
The callback parameter is ignored on all other endpoints.

### Build
You can build your own binaries by calling `build.sh`
```bash
build.sh build_dir
```

### Features & pull requests
Everyone can contribute to this project. Every pull request will be considered but it can also happen to be declined. 
To prevent unnecessary work, please consider to create a [feature issue](https://github.com/webklex/gogeoip/issues/new?template=feature_request.md) 
first, if you're planning to do bigger changes. Of course you can also create a new [feature issue](https://github.com/webklex/gogeoip/issues/new?template=feature_request.md)
if you're just wishing a feature ;)

>Off topic, rude or abusive issues will be deleted without any notice.


## Support
If you encounter any problems or if you find a bug, please don't hesitate to create a new [issue](https://github.com/webklex/gogeoip/issues).
However please be aware that it might take some time to get an answer.

If you need **immediate** or **commercial** support, feel free to send me a mail at github@webklex.com. 

## Change log

Please see [CHANGELOG](CHANGELOG.md) for more information what has changed recently.

## Security

If you discover any security related issues, please email github@webklex.com instead of using the issue tracker.

## Credits
- [Webklex][link-author]
- [All Contributors][link-contributors]

## License
The MIT License (MIT). Please see [License File](LICENSE.md) for more information.

[ico-license]: https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square

[link-author]: https://github.com/webklex
[link-contributors]: https://github.com/webklex/gogeoip/graphs/contributors
