# GoGeoIP Web API
GoGeoIP - a lightweight web api providing ip intelligence written in GO. This software provides an api to get as many 
information as possible for a given IP address or the current visitor. This includes network, system, location and 
user information. 
A [Live Demo](https://www.gogeoip.com/) is available under [gogeoip.com](https://www.gogeoip.com/).

[![Releases][ico-release]](https://github.com/Webklex/gogeoip/releases)
[![Downloads][ico-downloads]](https://github.com/Webklex/gogeoip/releases)
[![Demo][ico-website-status]](https://www.gogeoip.com/)
[![License][ico-license]](LICENSE.md)
[![Hits][ico-hits]][link-hits]



![geo_ip_web_gui](https://raw.githubusercontent.com/webklex/gogeoip-gui/master/geo_ip_web_gui.jpg)

## Table of Contents
- [Features](#features)
- [Installation](#installation)
  - [GUI](#gui)
- [Configuration](#server-options)
  - [HTTP & HTTPS](#http--https)
  - [Letsencrypt](#letsencrypt)
  - [Middlewares & Extensions](#middlewares--extensions)
  - [Rate limiting & Quota management](#rate-limiting--quota-management)
  - [MaxMind](#maxmind)
  - [ip2location](#ip2location)
  - [Tor Project](#tor-project)
  - [Logging](#logging)
  - [Memcache](#memcache)
  - [Redis](#redis)
  - [Additional](#additional)
- [Database](#database)
- [Api](#api)
  - [Output](#output)
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
* Serve the default [PX8LITEBIN](https://lite.ip2location.com/database) free database that is downloaded and updated automatically in background on a configurable schedule, or
* Serve the commercial [PX8BIN](https://lite.ip2location.com/database) database from ip2location, either as a local file that you provide and update periodically (so the server can reload it), or configured to be downloaded periodically using your API token
* Multiple languages are supported (en, ru, es, jp, fr, de)
* Detect VPN anonymizer, open proxies, web proxies, Tor exits, data center, web hosting (DCH) range and search engine robots (SES).
* Supports Linux, OS X, FreeBSD, and Windows
* Setup wizard

### Requirements
A Free MaxMind and / or ip2location License will be required and can be easily obtained:
1. [Sign up for a MaxMind account](https://www.maxmind.com/en/geolite2/signup) (no purchase required)
2. Set your password and [create a license key](https://www.maxmind.com/en/accounts/current/license-key)
3. [Sign up for a IP2Location account](https://lite.ip2location.com/sign-up) (no purchase required)
4. [Create access token](https://lite.ip2location.com/file-download)

### Installation
Download and unpack a fitting [pre-compiled binary](https://github.com/webklex/gogeoip/releases) or build a binary 
yourself by by following the [build](#build) instructions.

Continue by configuring your application:
```bash
geoip \
    -mm-user-id 100000 \
    -mm-license-key 0AAaAaaAa0A0AAaA \
    -i2l-token 0BBbBbbBb0B0BBbB \
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
You can configure the web server via command line flags, the config file `conf/settings.config` or by using the `-setup` flag:
```bash
geoip -setup
```

#### HTTP & HTTPS
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -http                  | HTTP                 | string | localhost:8080       | Address in form of ip:port to listen                        |
| -https                 | HTTPS                | string |                      | Address in form of ip:port to listen                        |
| -write-timeout         | WRITE_TIMEOUT        | int    | 15000000000          | Write timeout in nanoseconds for HTTP and HTTPS client connections |
| -read-timeout          | READ_TIMEOUT         | int    | 30000000000          | Read timeout in nanoseconds for HTTP and HTTPS client connections |
| -tcp-fast-open         | TCP_FAST_OPEN        | bool   | false                | Enable TCP fast open                                        |
| -tcp-naggle            | TCP_NAGGLE           | bool   | false                | Enable TCP Nagle's algorithm                                |
| -http2                 | HTTP2                | bool   | true                 | Enable HTTP/2 when TLS is enabled                           |
| -hsts                  | HSTS                 | string |                      |                                                             |
| -key                   | KEY                  | string | key.pem              | X.509 key file for HTTPS server                             |
| -cert                  | CERT                 | string | cert.pem             | X.509 certificate file for HTTPS server                     |

#### Letsencrypt
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -letsencrypt           | LETSENCRYPT          | bool   | false                | Enable automatic TLS using letsencrypt.org                  |
| -letsencrypt-email     | LETSENCRYPT_EMAIL    | string |                      | Optional email to register with letsencrypt                 |
| -letsencrypt-hosts     | LETSENCRYPT_HOSTS    | string |                      | Comma separated list of hosts for the certificate           |
| -letsencrypt-cert-dir  | LETSENCRYPT_CERT_DIR | string |                      | Letsencrypt cert dir                                        |

#### Middlewares & Extensions
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -use-x-forwarded-for   | USE_X_FORWARDED_FOR  | bool   | false                | Use the X-Forwarded-For header when available (e.g. behind proxy) |
| -cors-origin           | CORS_ORIGIN          | string | *                    | Comma separated list of CORS origins endpoints              |
| -api-prefix            | API_PREFIX           | string | /                    | API endpoint prefix                                         |
| -gui                   | GUI                  | string |                      | Web gui directory                                           |

##### Rate limiting & Quota management
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -quota-backend         | QUOTA_BACKEND        | string | redis                | Backend for rate limiter: map, redis, or memcache           |
| -quota-burst           | QUOTA_BURST          | int    | 3                    | Max requests per source IP per request burst                |
| -quota-interval        | QUOTA_INTERVAL       | int    | 3600000000000        | Quota expiration interval, per source IP querying the API in nanoseconds |
| -quota-max             | QUOTA_MAX            | int    | 1                    | "Max requests per source IP per interval; set 0 to turn quotas off |

#### MaxMind
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -mm-license-key           | MM_LICENSE_KEY          | string |                      | MaxMind License Key                                         |
| -mm-user-id               | MM_USER_ID              | string |                      | MaxMind User ID                                             |
| -mm-product-id            | MM_PRODUCT_ID           | string | GeoLite2-City        | MaxMind Product ID                                          |
| -mm-retry                 | MM_RETRY_INTERVAL       | int    | 7200000000000        | Max time to wait before retrying to download a MaxMind database |
| -mm-update                | MM_UPDATE_INTERVAL      | int    | 86400000000000       | MaxMind database update check interval in nanoseconds               |
| -mm-updates-host          | MM_UPDATES_HOST         | string | download.maxmind.com | MaxMind Updates Host                                        |

#### ip2location
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -i2l-token             | I2L_TOKEN            | string |                      | ip2location access token                                         |
| -i2l-product-id        | I2L_PRODUCT_ID       | string | PX8LITEBIN           | ip2location Product ID                                          |
| -i2l-retry             | I2L_RETRY_INTERVAL   | int    | 7200000000000        | Max time to wait before retrying to download a ip2location database |
| -i2l-update            | I2L_UPDATE_INTERVAL  | int    | 86400000000000       | ip2location database update check interval in nanoseconds               |
| -i2l-updates-host      | I2L_UPDATES_HOST     | string | www.ip2location.com  | ip2location Updates Host                                        |

#### Tor Project
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -tor-exit-check        | TOR_EXIT             | string | 8.8.8.8              | MaxMind Product ID                                          |
| -tor-retry             | TOR_RETRY_INTERVAL   | int    | 7200000000000        | Max time in nanoseconds to wait before retrying to download database |
| -tor-update            | TOR_UPDATE_INTERVAL  | int    | 86400000000000       | Database update check interval in nanoseconds               |
| -tor-updates-host      | TOR_UPDATES_HOST     | string | check.torproject.org | MaxMind Updates Host                                        |

#### Logging
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -logtostdout           | LOGTOSTDOUT          | bool   | false                | Log to stdout instead of stderr                             |
| -log-file              | LOG_FILE             | string |                      | Log file location                             |
| -logtimestamp          | LOGTIMESTAMP         | bool   | true                 | Prefix non-access logs with timestamp                       |

#### Memcache
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -memcache              | MEMCACHE             | string | localhost:11211      | Memcache address in form of host:port[,host:port] for quota |
| -memcache-timeout      | MEMCACHE_TIMEOUT     | int    | 1000000000           | Memcache read/write timeout in nanoseconds                  |

#### Redis
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -redis                 | REDIS                | string | localhost:6379       | Redis address in form of host:port[,host:port] for quota    |
| -redis-timeout         | REDIS_TIMEOUT        | int    | 1000000000           | Redis read/write timeout in nanoseconds                     |

#### Additional
| CLI                    | Config               | Type   | Default              | Description                                                 |
| :--------------------- | :------------------- | :----- | :------------------- | :---------------------------------------------------------- |
| -silent                | SILENT               | bool   | false                | Disable HTTP and HTTPS log request details                  |
| -config                |                      | string | conf/settings.config | Config file path                                            |
| -setup                 |                      | bool   | false                | Run the setup wizard                                        |
| -save                  |                      | bool   | false                | Save config                                                 |
| -version               |                      | bool   | false                | Show version and exit                                       |
| -help                  |                      | bool   | false                | Show help and exit                                          |

If you're using LetsEncrypt.org to provision your TLS certificates, you have to listen for HTTPS on port 443. Following is an example of the server listening on 2 different ports: http (80) and https (443):
```bash
geoip \
    -mm-user-id 100000 \
    -mm-license-key 0AAaAaaAa0A0AAaA \
    -i2l-token 0BBbBbbBb0B0BBbB \
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
    "MM_USER_ID": "100000",
    "MM_LICENSE_KEY": "0AAaAaaAa0A0AAaA",
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

## Databases
The current implementation uses the free [GeoLite2 City](http://dev.maxmind.com/geoip/geoip2/geolite2/) database from 
MaxMind as well as the free [IP2Proxy](https://lite.ip2location.com/database/px8-ip-proxytype-country-region-city-isp-domain-usagetype-asn-lastseen) 
database from ip2location and the generic tor [exit node list](https://check.torproject.org/cgi-bin/TorBulkExitList.py?ip=8.8.8.8) 
provided by the [TorProject](https://www.torproject.org/).
If you have purchased the commercial database from MaxMind or ip2location, you can point the geoip web server or 
(Go API, for dev) to the URL containing the file, or local file, and the server will use it.
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


### Output
#### Network
| Name                  | Value type    | JSON                      | XML                   | CSV   | Comment   |
| :-------------------- | :------------ | :------------------------ | :-------------------- | :---- | :-------- |
| IP address            | string        | ip                        | IP                    | 0     |           |
| Number (ASN)          | integer       | as.number                 | AS.Number             | 1     |           |
| Organization          | string        | as.name                   | AS.Name               | 2     |           |
| ISP name              | string        | isp                       | Isp                   | 3     |           |
| Domain                | string        | domain                    | Domain                | 4     |           |
| TLDs                  | []string      | tld                       | Tld                   | 5     |           |
| Is bot                | bool          | bot                       | Bot                   | 6     |           |
| Is tor user           | bool          | tor                       | Tor                   | 7     |           |
| Is proxy user         | bool          | proxy                     | Proxy                 | 8     |           |
| Proxy type            | string        | proxy_type                | ProxyType             | 9     | [Available proxy types](https://lite.ip2location.com/database/px8-ip-proxytype-country-region-city-isp-domain-usagetype-asn-lastseen) |
| Last seen in days     | integer       | last_seen                 | LastSeen              | 10    |           |
| Usage type            | string        | usage_type                | UsageType             | 11    | [Available usage types](https://lite.ip2location.com/database/px8-ip-proxytype-country-region-city-isp-domain-usagetype-asn-lastseen) |

#### Location
| Name                  | Value type    | JSON                      | XML                   | CSV   | Comment   |
| :-------------------- | :------------ | :------------------------ | :-------------------- | :---- | :-------- |
| Region code           | string        | region_code               | RegionCode            | 12    |           |
| Region name           | string        | region_name               | RegionName            | 13    |           |
| City name             | string        | city                      | City                  | 14    |           |
| Zip code              | string        | zip_code                  | ZipCode               | 15    |           |
| Time zone             | string        | time_zone                 | TimeZone              | 16    |           |
| Latitude              | float         | latitude                  | Latitude              | 17    |           |
| Longitude             | float         | longitude                 | Longitude             | 18    |           |
| Accuracy radius       | integer       | accuracy_radius           | AccuracyRadius        | 19    |           |
| Metro code            | integer       | metro_code                | MetroCode             | 20    |           |
| Country code              | string        | country.code                  | Country.Code                  | 21    |           |
| CIOC                      | string        | country.cioc                  | Country.CIOC                  | 22    |           |
| CCN3                      | string        | country.ccn3                  | Country.CCN3                  | 23    |           |
| Call codes                | []string      | country.call_code             | Country.CallCode              | 24    |           |
| International call prefix | string        | country.international_prefix  | Country.InternationalPrefix   | 25    |           |
| Country capital           | string        | country.capital               | Country.Capital               | 26    |           |
| Country name              | string        | country.name                  | Country.Name                  | 27    |           |
| Full country name         | string        | country.full_name             | Country.FullName              | 28    |           |
| Country Area km²          | integer       | country.area                  | Country.Area                  | 29    |           |
| Country borders           | []string      | country.borders               | Country.Borders               | 30    |           |
| Latitude                  | float         | country.latitude              | Country.Latitude              | 31    |           |
| Longitude                 | float         | country.longitude             | Country.Longitude             | 32    |           |
| Max. Latitude             | float         | country.max_latitude          | Country.MaxLatitude           | 33    |           |
| Max. Longitude            | float         | country.max_longitude         | Country.MaxLongitude          | 34    |           |
| Min. Latitude             | float         | country.min_latitude          | Country.MinLatitude           | 35    |           |
| Min. Longitude            | float         | country.min_longitude         | Country.MinLongitude          | 36    |           |
| Currencies                | []{code,name} | country.currency              | Country.Currency              | 37    |           |
| Continent code            | string        | country.content.code          | Country.Continent.Code        | 38    |           |
| Continent name            | string        | country.content.name          | Country.Continent.Name        | 39    |           |
| Continent sub region      | string        | country.content.sub_region    | Country.Continent.SubRegion   | 40    |           |

#### System
| Name                  | Value type    | JSON          | XML       | CSV   | Comment   |
| :-------------------- | :------------ | :------------ | :-------- | :---- | :-------- |
| Operating System      | string        | os            | OS        | 41    |           |
| System architecture   | string        | os_version    | OSVersion | 42    |           |
| Browser               | string        | browser       | Browser   | 43    |           |
| Browser Version       | string        | version       | Version   | 44    |           |
| Device name           | string        | device        | Device    | 45    |           |
| Is mobile user        | bool          | mobile        | Mobile    | 46    |           |
| Is tablet user        | bool          | tablet        | Tablet    | 47    |           |
| Is desktop user       | bool          | desktop       | Desktop   | 48    |           |

#### User
| Name                  | Value type    | JSON                 | XML               | CSV   | Comment   |
| :-------------------- | :------------ | :------------------- | :---------------- | :---- | :-------- |
| Language              | string        | language.language    | Language.Language | 49    |           |
| Language region       | string        | language.region      | Language.Region   | 50    |           |
| Language tag          | string        | language.tag         | Language.Tag      | 51    |           |

#### CSV
```bash
curl :8080/csv/208.13.138.36
```
```
208.13.138.36,209,"CenturyLink Communications, LLC",,,.us,0,0,0,,0,,NV,,Las Vegas,839,89129,America/Los_Angeles,-115.2821,36.2473,20,US,USA,840,1,011,Washington D.C.,United States,United States of America,9372610.0000,CAN/MEX,39.4433,-98.9573,71.4411,-66.8854,17.8315,-179.2311,USD/USN/USS,,
```
```bash
curl :8080/csv/208.13.138.36?user
```
```
208.13.138.36,209,"CenturyLink Communications, LLC",,,.us,0,0,0,,0,,NV,,Las Vegas,839,89129,America/Los_Angeles,-115.2821,36.2473,20,US,USA,840,1,011,Washington D.C.,United States,United States of America,9372610.0000,CAN/MEX,39.4433,-98.9573,71.4411,-66.8854,17.8315,-179.2311,USD/USN/USS,,,Linux,Ubuntu Chromium,79.0.3945.79,x86_64,,0,0,1,en,US,en-US
```

#### XML
```bash
curl :8080/xml/208.13.138.36
```
```xml
<Response>
    <Network>
        <IP>208.13.138.36</IP>
        <AS>
            <Number>209</Number>
            <Name>CenturyLink Communications, LLC</Name>
        </AS>
        <Isp/>
        <Domain/>
        <Tld>.us</Tld>
        <Bot>false</Bot>
        <Tor>false</Tor>
        <Proxy>false</Proxy>
        <ProxyType/>
        <LastSeen>0</LastSeen>
        <UsageType/>
    </Network>
    <Location>
        <RegionCode>NV</RegionCode>
        <RegionName/>
        <City>Las Vegas</City>
        <ZipCode>89129</ZipCode>
        <TimeZone>America/Los_Angeles</TimeZone>
        <Longitude>-115.2821</Longitude>
        <Latitude>36.2473</Latitude>
        <AccuracyRadius>20</AccuracyRadius>
        <MetroCode>839</MetroCode>
        <Country>
            <Code>US</Code>
            <CIOC>USA</CIOC>
            <CCN3>840</CCN3>
            <CallCode>1</CallCode>
            <InternationalPrefix>011</InternationalPrefix>
            <Capital>Washington D.C.</Capital>
            <Name>United States</Name>
            <FullName>United States of America</FullName>
            <Area>9.37261e+06</Area>
            <Borders>CAN</Borders>
            <Borders>MEX</Borders>
            <Latitude>39.443256</Latitude>
            <Longitude>-98.95734</Longitude>
            <MaxLatitude>71.441055</MaxLatitude>
            <MaxLongitude>-66.885414</MaxLongitude>
            <MinLatitude>17.831509</MinLatitude>
            <MinLongitude>-179.23108</MinLongitude>
            <Currency>
                <Code>USD</Code>
                <Name/>
            </Currency>
            <Currency>
                <Code>USN</Code>
                <Name/>
            </Currency>
            <Currency>
                <Code>USS</Code>
                <Name/>
            </Currency>
            <Continent>
                <Code/>
                <Name>North America</Name>
                <SubRegion/>
            </Continent>
        </Country>
    </Location>
</Response>
```
```bash
curl :8080/xml/208.13.138.36?user
```
```xml
<Response>
    <Network>
        <IP>208.13.138.36</IP>
        <AS>
            <Number>209</Number>
            <Name>CenturyLink Communications, LLC</Name>
        </AS>
        <Isp/>
        <Domain/>
        <Tld>.us</Tld>
        <Bot>false</Bot>
        <Tor>false</Tor>
        <Proxy>false</Proxy>
        <ProxyType/>
        <LastSeen>0</LastSeen>
        <UsageType/>
    </Network>
    <Location>
        <RegionCode>NV</RegionCode>
        <RegionName/>
        <City>Las Vegas</City>
        <ZipCode>89129</ZipCode>
        <TimeZone>America/Los_Angeles</TimeZone>
        <Longitude>-115.2821</Longitude>
        <Latitude>36.2473</Latitude>
        <AccuracyRadius>20</AccuracyRadius>
        <MetroCode>839</MetroCode>
        <Country>
            <Code>US</Code>
            <CIOC>USA</CIOC>
            <CCN3>840</CCN3>
            <CallCode>1</CallCode>
            <InternationalPrefix>011</InternationalPrefix>
            <Capital>Washington D.C.</Capital>
            <Name>United States</Name>
            <FullName>United States of America</FullName>
            <Area>9.37261e+06</Area>
            <Borders>CAN</Borders>
            <Borders>MEX</Borders>
            <Latitude>39.443256</Latitude>
            <Longitude>-98.95734</Longitude>
            <MaxLatitude>71.441055</MaxLatitude>
            <MaxLongitude>-66.885414</MaxLongitude>
            <MinLatitude>17.831509</MinLatitude>
            <MinLongitude>-179.23108</MinLongitude>
            <Currency>
                <Code>USD</Code>
                <Name/>
            </Currency>
            <Currency>
                <Code>USN</Code>
                <Name/>
            </Currency>
            <Currency>
                <Code>USS</Code>
                <Name/>
            </Currency>
            <Continent>
                <Code/>
                <Name>North America</Name>
                <SubRegion/>
            </Continent>
        </Country>
    </Location>
    <System>
        <OS>Linux</OS>
        <Browser>Ubuntu Chromium</Browser>
        <Version>79.0.3945.79</Version>
        <OSVersion>x86_64</OSVersion>
        <Device/>
        <Mobile>false</Mobile>
        <Tablet>false</Tablet>
        <Desktop>true</Desktop>
    </System>
    <User>
        <Language>
            <Language>en</Language>
            <Region>US</Region>
            <Tag>en-US</Tag>
        </Language>
    </User>
</Response>
```

#### JSON
```bash
curl :8080/json/208.13.138.36
```
```json
{
  "network": {
    "ip": "208.13.138.36",
    "as": {
      "number": 209,
      "name": "CenturyLink Communications, LLC"
    },
    "isp": "",
    "domain": "",
    "tld": [".us"],
    "bot": false,
    "tor": false,
    "proxy": false,
    "proxy_type": "",
    "last_seen": 0,
    "usage_type": ""
  },
  "location": {
    "region_code": "NV",
    "region_name": "",
    "city": "Las Vegas",
    "zip_code": "89129",
    "time_zone": "America/Los_Angeles",
    "longitude": -115.2821,
    "latitude": 36.2473,
    "accuracy_radius": 20,
    "metro_code": 839,
    "country": {
      "code": "US",
      "cioc": "USA",
      "ccn3": "840",
      "call_code": ["1"],
      "international_prefix": "011",
      "capital": "Washington D.C.",
      "name": "United States",
      "full_name": "United States of America",
      "area": 9372610,
      "borders": ["CAN", "MEX"],
      "latitude": 39.443256,
      "longitude": -98.95734,
      "max_latitude": 71.441055,
      "max_longitude": -66.885414,
      "min_latitude": 17.831509,
      "min_longitude": -179.23108,
      "currency": [{
          "code": "USD",
          "name": ""
       }, {
          "code": "USN",
          "name": ""
       }, {
          "code": "USS",
          "name": ""
      }],
      "continent": {
        "code": "",
        "name": "North America",
        "sub_region": ""
      }
    }
  }
}
```
```bash
curl :8080/json/208.13.138.36?user
```
```json
{
  "network": {
    "ip": "208.13.138.36",
    "as": {
      "number": 209,
      "name": "CenturyLink Communications, LLC"
    },
    "isp": "",
    "domain": "",
    "tld": [".us"],
    "bot": false,
    "tor": false,
    "proxy": false,
    "proxy_type": "",
    "last_seen": 0,
    "usage_type": ""
  },
  "location": {
    "region_code": "NV",
    "region_name": "",
    "city": "Las Vegas",
    "zip_code": "89129",
    "time_zone": "America/Los_Angeles",
    "longitude": -115.2821,
    "latitude": 36.2473,
    "accuracy_radius": 20,
    "metro_code": 839,
    "country": {
      "code": "US",
      "cioc": "USA",
      "ccn3": "840",
      "call_code": ["1"],
      "international_prefix": "011",
      "capital": "Washington D.C.",
      "name": "United States",
      "full_name": "United States of America",
      "area": 9372610,
      "borders": ["CAN", "MEX"],
      "latitude": 39.443256,
      "longitude": -98.95734,
      "max_latitude": 71.441055,
      "max_longitude": -66.885414,
      "min_latitude": 17.831509,
      "min_longitude": -179.23108,
      "currency": [{
          "code": "USD",
          "name": ""
       }, {
          "code": "USN",
          "name": ""
       }, {
          "code": "USS",
          "name": ""
      }],
      "continent": {
        "code": "",
        "name": "North America",
        "sub_region": ""
      }
    }
  },
  "system": {
    "os": "Linux",
    "browser": "Ubuntu Chromium",
    "version": "79.0.3945.79",
    "os_version": "x86_64",
    "device": "",
    "mobile": false,
    "tablet": false,
    "desktop": true
  },
  "user": {
    "language": {
      "language": "en",
      "region": "US",
      "tag": "en-US"
    }
  }
}
```

#### JSONP
```bash
curl :8080/json/208.13.138.36?callback=foobar
```
```javascript
foobar({
 "network": {
   "ip": "208.13.138.36",
   "as": {
     "number": 209,
     "name": "CenturyLink Communications, LLC"
   },
   "isp": "",
   "domain": "",
   "tld": [".us"],
   "bot": false,
   "tor": false,
   "proxy": false,
   "proxy_type": "",
   "last_seen": 0,
   "usage_type": ""
 },
 "location": {
   "region_code": "NV",
   "region_name": "",
   "city": "Las Vegas",
   "zip_code": "89129",
   "time_zone": "America/Los_Angeles",
   "longitude": -115.2821,
   "latitude": 36.2473,
   "accuracy_radius": 20,
   "metro_code": 839,
   "country": {
     "code": "US",
     "cioc": "USA",
     "ccn3": "840",
     "call_code": ["1"],
     "international_prefix": "011",
     "capital": "Washington D.C.",
     "name": "United States",
     "full_name": "United States of America",
     "area": 9372610,
     "borders": ["CAN", "MEX"],
     "latitude": 39.443256,
     "longitude": -98.95734,
     "max_latitude": 71.441055,
     "max_longitude": -66.885414,
     "min_latitude": 17.831509,
     "min_longitude": -179.23108,
     "currency": [{
         "code": "USD",
         "name": ""
      }, {
         "code": "USN",
         "name": ""
      }, {
         "code": "USS",
         "name": ""
     }],
     "continent": {
       "code": "",
       "name": "North America",
       "sub_region": ""
     }
   }
 }
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
[ico-release]: https://img.shields.io/github/v/release/webklex/gogeoip?style=flat-square
[ico-downloads]: https://img.shields.io/github/downloads/webklex/gogeoip/total?style=flat-square
[ico-website-status]: https://img.shields.io/website?down_message=Offline&label=Demo&style=flat-square&up_message=Online&url=https%3A%2F%2Fwww.gogeoip.com%2F
[ico-hits]: https://hits.webklex.com/svg/webklex/gogeoip?1

[link-hits]: https://hits.webklex.com
[link-author]: https://github.com/webklex
[link-contributors]: https://github.com/webklex/gogeoip/graphs/contributors
