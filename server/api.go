package server

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/go-web/httpmux"
	"github.com/mileusna/useragent"
	"github.com/pariz/gountries"
	"golang.org/x/text/language"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// openDB opens and returns the IP database file or URL.
func (s *Server) openDB() error {
	if _, err := s.Api.i2lDB.Start(); err != nil {return err}
	if _, err := s.Api.torDB.Start(); err != nil {return err}
	if _, err := s.Api.db.Start(); err != nil {return err}
	if _, err := s.Api.asnDB.Start(); err != nil {return err}

	return nil
}

func (s *Server) IpLookUp(writer writerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := getRequestParam(r, "host")
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			http.NotFound(w, r)
			return
		}

		ip, q := ips[rand.Intn(len(ips))], &GeoIpQuery{}
		if err := s.Api.db.Lookup(ip, &q.DefaultQuery); err != nil {
			fmt.Println(err)
			http.Error(w, "Try again later.", http.StatusServiceUnavailable)
			return
		}
		if err := s.Api.asnDB.Lookup(ip, &q.ASNDefaultQuery); err != nil {
			http.Error(w, "Try again later.", http.StatusServiceUnavailable)
			return
		}

		q.ProxyDefaultQuery = s.Api.i2lDB.Lookup(ip)

		w.Header().Set("X-Database-Date", s.Api.db.Updater.Date().Format(http.TimeFormat))
		lang := getRequestParam(r, "lang")

		q.IsTorUser = s.Api.torDB.Lookup(ip)
		resp := q.Record(ip, lang, r)
		writer(w, r, resp)
	}
}

func (q *GeoIpQuery) Translate(names map[string]string, lang string) string {
	if val, ok := names[lang]; ok {
		return val
	}
	return names["en"]
}

func (q *GeoIpQuery) TranslateCountry(t map[string]gountries.BaseLang, lang string) *gountries.BaseLang {
	if val, ok := t["ENG"]; ok {
		return &val
	}
	return &gountries.BaseLang{}
}

func getMostPreferredLanguage(tag []language.Tag, q []float32) language.Tag {
	lang := language.Tag{}
	score := 0.0
	for i, val := range q {
		f := float64(val)
		if score < f {
			score = f
			lang = tag[i]
		}
	}
	return lang
}

func (q *GeoIpQuery) Record(ip net.IP, lang string, request *http.Request) *ResponseRecord {
	lang = parseAcceptLanguage(lang, q.Country.Names)

	//CountryCode: 		q.Country.ISOCode,
	//ContinentCode: 		q.Translate(q.Continent.Names, lang),
	//CountryName: 		q.Translate(q.Country.Names, lang),
	query := gountries.New()
	country, err := query.FindCountryByAlpha(q.Country.ISOCode)
	if err != nil {
		country = gountries.Country{
			Name: struct {
				gountries.BaseLang `yaml:",inline"`
				Native             map[string]gountries.BaseLang
			}{},
			EuMember:     false,
			LandLocked:   false,
			Nationality:  "",
			TLDs:         nil,
			Languages:    nil,
			Translations: make(map[string]gountries.BaseLang),
			Currencies:   nil,
			Borders:      nil,
			Codes:        gountries.Codes{},
			Geo:          gountries.Geo{},
			Coordinates:  gountries.Coordinates{},
		}
	}

	c := make([]*CurrencyRecord, len(country.Currencies))
	for i := range c {
		c[i] = &CurrencyRecord{
			Code: country.Currencies[i],
		}
	}

	r := &ResponseRecord{
		Location: &LocationRecord{
			MetroCode:      q.Location.MetroCode,
			City:           q.Translate(q.City.Names, lang),
			ZipCode:        q.Postal.Code,
			TimeZone:       q.Location.TimeZone,
			Latitude:       roundFloat(q.Location.Latitude, .5, 4),
			Longitude:      roundFloat(q.Location.Longitude, .5, 4),
			AccuracyRadius: q.Location.AccuracyRadius,
			Country:        &CountryRecord{
				Code: q.Country.ISOCode,
				Name: country.Name.Common,
				FullName: country.Name.Official,
				Currency: c,
				Borders: country.Borders,
				CIOC: country.CIOC,
				CCN3: country.CCN3,
				CallCode: country.CallingCodes,
				InternationalPrefix: country.InternationalPrefix,
				Capital: country.Capital,
				Area: country.Area,
				Latitude: country.Latitude,
				Longitude: country.Longitude,
				MaxLatitude: country.MaxLatitude,
				MaxLongitude: country.MaxLongitude,
				MinLatitude: country.MinLatitude,
				MinLongitude: country.MinLongitude,
				Continent:   &ContinentRecord{
					Code: "",
					Name: country.Continent,
				},
			},
		},
		Network:  &NetworkRecord{
			AS:        &ASRecord{
				Number: q.AutonomousSystemNumber,
				Name: 	q.AutonomousSystemOrganization,
			},
			IP:         ip.String(),
			Isp:	    q.Isp,
			Tld:	    country.TLDs,
			Domain:	    q.Domain,
			Tor: 		q.IsTorUser,
			ProxyType: 	q.ProxyType,
			Proxy: 		q.Proxy,
			UsageType: 	q.UsageType,
			LastSeen:	uint(q.LastSeen),
		},
		User: &UserRecord{},
	}

	if len(q.Region) > 0 {
		r.Location.RegionCode = q.Region[0].ISOCode
		r.Location.RegionName = q.Region[0].Names[lang]
	}

	if len(request.URL.Query()["user"]) > 0 {
		t, qq, _ := language.ParseAcceptLanguage(request.Header.Get("Accept-Language"))

		plang := getMostPreferredLanguage(t, qq)
		base, _ := plang.Base()
		region, _ := plang.Region()

		userAgent := ua.Parse(request.Header.Get("User-Agent"))
		r.System = &SystemRecord{
			OS:      	userAgent.OS,
			Browser: 	userAgent.Name,
			Version: 	userAgent.Version,
			OSVersion: 	userAgent.OSVersion,
			Device: 	userAgent.Device,
			Mobile: 	userAgent.Mobile,
			Tablet: 	userAgent.Tablet,
			Desktop: 	userAgent.Desktop,
		}
		r.User = &UserRecord{Language:&LanguageRecord{
			Language: base.String(),
			Region:   region.String(),
			Tag:      plang.String(),
		}}

		r.Network.Bot = userAgent.Bot
	}

	return r
}

func (rr *ResponseRecord) String() string {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	w.UseCRLF = true
	var err error

	currency := make([]string, len(rr.Location.Country.Currency))
	for i := range rr.Location.Country.Currency {
		currency[i] = rr.Location.Country.Currency[i].Code
	}
	var SystemBot int;if rr.Network.Bot {SystemBot = 1}
	var SystemTor int;if rr.Network.Tor {SystemTor = 1}
	var SystemProxy int;if rr.Network.Proxy {SystemProxy = 1}

	if rr.User != nil && rr.System != nil {

		var SystemMobile int;if rr.System.Mobile {SystemMobile = 1}
		var SystemTablet int;if rr.System.Tablet {SystemTablet = 1}
		var SystemDesktop int;if rr.System.Desktop {SystemDesktop = 1}

		err = w.Write([]string{
			rr.Network.IP,
			strconv.Itoa(int(rr.Network.AS.Number)),
			rr.Network.AS.Name,
			rr.Network.Isp,
			rr.Network.Domain,
			strings.Join(rr.Network.Tld, "/"),
			strconv.Itoa(SystemBot),
			strconv.Itoa(SystemTor),
			strconv.Itoa(SystemProxy),
			rr.Network.ProxyType,
			strconv.Itoa(int(rr.Network.LastSeen)),
			rr.Network.UsageType,

			rr.Location.RegionCode,
			rr.Location.RegionName,
			rr.Location.City,
			strconv.Itoa(int(rr.Location.MetroCode)),
			rr.Location.ZipCode,
			rr.Location.TimeZone,
			strconv.FormatFloat(rr.Location.Longitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Latitude, 'f', 4, 64),
			strconv.Itoa(int(rr.Location.AccuracyRadius)),

			rr.Location.Country.Code,
			rr.Location.Country.CIOC,
			rr.Location.Country.CCN3,
			strings.Join(rr.Location.Country.CallCode, "/"),
			rr.Location.Country.InternationalPrefix,
			rr.Location.Country.Capital,
			rr.Location.Country.Name,
			rr.Location.Country.FullName,
			strconv.FormatFloat(rr.Location.Country.Area, 'f', 4, 64),
			strings.Join(rr.Location.Country.Borders, "/"),
			strconv.FormatFloat(rr.Location.Country.Latitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.Longitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.MaxLatitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.MaxLongitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.MinLatitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.MinLongitude, 'f', 4, 64),
			strings.Join(currency, "/"),
			rr.Location.Country.Continent.Code,
			rr.Location.Country.Continent.Name,
			rr.Location.Country.Continent.SubRegion,

			rr.System.OS,
			rr.System.Browser,
			rr.System.Version,
			rr.System.OSVersion,
			rr.System.Device,
			strconv.Itoa(SystemMobile),
			strconv.Itoa(SystemTablet),
			strconv.Itoa(SystemDesktop),

			rr.User.Language.Language,
			rr.User.Language.Region,
			rr.User.Language.Tag,
		})
	}else{
		err = w.Write([]string{
			rr.Network.IP,
			strconv.Itoa(int(rr.Network.AS.Number)),
			rr.Network.AS.Name,
			rr.Network.Isp,
			rr.Network.Domain,
			strings.Join(rr.Network.Tld, "/"),
			strconv.Itoa(SystemBot),
			strconv.Itoa(SystemTor),
			strconv.Itoa(SystemProxy),
			rr.Network.ProxyType,
			strconv.Itoa(int(rr.Network.LastSeen)),
			rr.Network.UsageType,

			rr.Location.RegionCode,
			rr.Location.RegionName,
			rr.Location.City,
			strconv.Itoa(int(rr.Location.MetroCode)),
			rr.Location.ZipCode,
			rr.Location.TimeZone,
			strconv.FormatFloat(rr.Location.Longitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Latitude, 'f', 4, 64),
			strconv.Itoa(int(rr.Location.AccuracyRadius)),

			rr.Location.Country.Code,
			rr.Location.Country.CIOC,
			rr.Location.Country.CCN3,
			strings.Join(rr.Location.Country.CallCode, "/"),
			rr.Location.Country.InternationalPrefix,
			rr.Location.Country.Capital,
			rr.Location.Country.Name,
			rr.Location.Country.FullName,
			strconv.FormatFloat(rr.Location.Country.Area, 'f', 4, 64),
			strings.Join(rr.Location.Country.Borders, "/"),
			strconv.FormatFloat(rr.Location.Country.Latitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.Longitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.MaxLatitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.MaxLongitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.MinLatitude, 'f', 4, 64),
			strconv.FormatFloat(rr.Location.Country.MinLongitude, 'f', 4, 64),
			strings.Join(currency, "/"),
			rr.Location.Country.Continent.Code,
			rr.Location.Country.Continent.Name,
			rr.Location.Country.Continent.SubRegion,
		})
	}
	if err != nil {
		return ""
	}
	w.Flush()
	return b.String()
}

func getRequestParam(r *http.Request, param string) string {
	switch param {
	case "host":
		host := httpmux.Params(r).ByName("host")
		if len(host) > 0 && host[0] == '/' {
			host = host[1:]
		}
		if strings.Contains(host, "?") {
			host = strings.Split(host, "?")[0]
		}
		if host == "" {
			host, _, _ = net.SplitHostPort(r.RemoteAddr)
			if host == "" {
				host = r.RemoteAddr
			}
		}
		return host
	case "lang":
		lang := ""
		if len(r.URL.Query()["lang"]) > 0 {
			lang = r.URL.Query()["lang"][0]
		}
		if lang == "" {
			lang = r.Header.Get("Accept-Language")
		}

		return lang
	}

	return ""
}

func parseAcceptLanguage(header string, dbLangs map[string]string) string {
	// supported languages -- i.e. languages available in the DB
	matchLangs := []language.Tag{
		language.English,
	}

	// parse available DB languages and add to supported
	for name := range dbLangs {
		matchLangs = append(matchLangs, language.Raw.Make(name))
	}

	var matcher = language.NewMatcher(matchLangs)

	// parse header
	t, _, _ := language.ParseAcceptLanguage(header)
	// match most acceptable language
	tag, _, _ := matcher.Match(t...)
	// extract base language
	base, _ := tag.Base()

	return base.String()
}

func roundFloat(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow
}

func csvResponse(w http.ResponseWriter, r *http.Request, d *ResponseRecord) {
	w.Header().Set("Content-Type", "text/csv")
	if n, err := io.WriteString(w, d.String()); err != nil || n <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func xmlResponse(w http.ResponseWriter, r *http.Request, d *ResponseRecord) {
	w.Header().Set("Content-Type", "application/xml")
	x := xml.NewEncoder(w)
	x.Indent("", "\t")
	if err := x.Encode(d); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if n, err := w.Write([]byte{'\n'}); err != nil || n <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func jsonResponse(w http.ResponseWriter, r *http.Request, d *ResponseRecord) {
	if cb := r.FormValue("callback"); cb != "" {
		w.Header().Set("Content-Type", "application/javascript")
		if n, err := io.WriteString(w, cb); err != nil || n <= 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if n, err := w.Write([]byte("(")); err != nil || n <= 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		b, err := json.Marshal(d)
		if err == nil {
			if n, err := w.Write(b); err != nil || n <= 0 {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}
		if n, err := io.WriteString(w, ");"); err != nil || n <= 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(d); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}
