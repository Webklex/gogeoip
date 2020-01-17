package server

import (
	"../utils/db"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"github.com/go-web/httpmux"
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
func (s *Server) openDB() (*db.DB, error) {
	return s.Api.db.OpenURL()
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
		err = s.Api.db.Lookup(ip, &q.DefaultQuery)
		if err != nil {
			http.Error(w, "Try again later.", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("X-Database-Date", s.Api.db.Date().Format(http.TimeFormat))
		lang := getRequestParam(r, "lang")

		resp := q.Record(ip, lang)
		writer(w, r, resp)
	}
}

func (q *GeoIpQuery) Translate(names map[string]string, lang string) string {
	if val, ok := names[lang]; ok {
		return val
	}
	return names["en"]
}

func (q *GeoIpQuery) Record(ip net.IP, lang string) *responseRecord {
	lang = parseAcceptLanguage(lang, q.Country.Names)

	r := &responseRecord{
		IP:          ip.String(),
		IsInEuropeanUnion: q.Country.IsInEuropeanUnion,
		CountryCode: q.Country.ISOCode,
		ContinentCode: q.Translate(q.Continent.Names, lang),
		CountryName: q.Translate(q.Country.Names, lang),
		City:        q.Translate(q.City.Names, lang),
		ZipCode:     q.Postal.Code,
		TimeZone:    q.Location.TimeZone,
		Latitude:    roundFloat(q.Location.Latitude, .5, 4),
		Longitude:   roundFloat(q.Location.Longitude, .5, 4),
		MetroCode:   q.Location.MetroCode,
		AccuracyRadius:   q.Location.AccuracyRadius,
	}
	if len(q.Region) > 0 {
		r.RegionCode = q.Region[0].ISOCode
		r.RegionName = q.Region[0].Names[lang]
	}
	return r
}

func (rr *responseRecord) String() string {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	w.UseCRLF = true
	var inEU int
	if rr.IsInEuropeanUnion {
		inEU = 1
	}
	err := w.Write([]string{
		rr.IP,
		rr.ContinentCode,
		rr.CountryCode,
		rr.CountryName,
		rr.RegionCode,
		rr.RegionName,
		rr.City,
		rr.ZipCode,
		rr.TimeZone,
		strconv.FormatFloat(rr.Latitude, 'f', 4, 64),
		strconv.FormatFloat(rr.Longitude, 'f', 4, 64),
		strconv.Itoa(int(rr.MetroCode)),
		strconv.Itoa(inEU),
		strconv.Itoa(int(rr.AccuracyRadius)),
	})
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

func csvResponse(w http.ResponseWriter, r *http.Request, d *responseRecord) {
	w.Header().Set("Content-Type", "text/csv")
	if n, err := io.WriteString(w, d.String()); err != nil || n <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func xmlResponse(w http.ResponseWriter, r *http.Request, d *responseRecord) {
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

func jsonResponse(w http.ResponseWriter, r *http.Request, d *responseRecord) {
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
