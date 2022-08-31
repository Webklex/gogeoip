package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mileusna/useragent"
	"github.com/pariz/gountries"
	"github.com/webklex/gogeoip/src/models"
	"github.com/webklex/gogeoip/src/paginator"
	"golang.org/x/text/language"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (a *Application) apiV1Detail(w http.ResponseWriter, r *http.Request) {
	host := getParam(r, "host")

	ip, err := a.resolveIP(host)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	region := models.Region{}
	if len(ip.City.Regions) > 0 {
		region = ip.City.Regions[0]
	}

	geo := staticGeoInfo(ip.Country.ISOCode)

	currencies := make([]LegacyCurrencyRecord, len(geo.Currencies))
	for i := range currencies {
		currencies[i] = LegacyCurrencyRecord{
			Code: geo.Currencies[i],
		}
	}

	t, qq, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))

	plang := getMostPreferredLanguage(t, qq)
	base, _ := plang.Base()
	_region, _ := plang.Region()

	userAgent := useragent.Parse(r.Header.Get("User-Agent"))

	response := &LegacyResponseRecord{
		Network: LegacyNetworkRecord{
			IP: ip.Address,
			AS: LegacyASRecord{
				Number: ip.AutonomousSystem.Number,
				Name:   ip.AutonomousSystem.Name,
			},
			Isp:       ip.Isp.Name,
			Domain:    ip.Network.Domain,
			Tld:       geo.TLDs,
			Bot:       userAgent.Bot,
			Tor:       ip.IsTorExitNode,
			Proxy:     ip.IsPublicProxy || ip.IsAnonymousProxy,
			ProxyType: ip.ProxyType,
			LastSeen:  uint(ip.LastSeen),
			UsageType: ip.Type,
		},
		Location: LegacyLocationRecord{
			RegionCode:     region.Code,
			RegionName:     region.Name,
			City:           ip.City.Name,
			ZipCode:        ip.Postal.Zip,
			TimeZone:       ip.City.TimeZone,
			Longitude:      float64(ip.Longitude),
			Latitude:       float64(ip.Latitude),
			AccuracyRadius: ip.AccuracyRadius,
			MetroCode:      ip.City.MetroCode,
			Country: LegacyCountryRecord{
				Code:                ip.Country.ISOCode,
				CIOC:                geo.CIOC,
				CCN3:                geo.CCN3,
				CallCode:            geo.CallingCodes,
				InternationalPrefix: geo.InternationalPrefix,
				Capital:             geo.Capital,
				Name:                geo.Name.Common,
				FullName:            geo.Name.Official,
				Area:                geo.Area,
				Borders:             geo.Borders,
				Latitude:            geo.Latitude,
				Longitude:           geo.Longitude,
				MaxLatitude:         geo.MaxLatitude,
				MaxLongitude:        geo.MaxLongitude,
				MinLatitude:         geo.MinLatitude,
				MinLongitude:        geo.MinLongitude,
				Currency:            currencies,
				Continent: LegacyContinentRecord{
					Code:      ip.Country.Continent.Code,
					Name:      ip.Country.Continent.Name,
					SubRegion: geo.SubRegion,
				},
			},
		},
		System: SystemRecord{
			OS:        userAgent.OS,
			Browser:   userAgent.Name,
			Version:   userAgent.Version,
			OSVersion: userAgent.OSVersion,
			Device:    userAgent.Device,
			Mobile:    userAgent.Mobile,
			Tablet:    userAgent.Tablet,
			Desktop:   userAgent.Desktop,
		},
		User: UserRecord{Language: &LanguageRecord{
			Language: base.String(),
			Region:   _region.String(),
			Tag:      plang.String(),
		}},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (a *Application) apiCountry(w http.ResponseWriter, r *http.Request) {
	country := getParam(r, "country")
	if country == "" {
		host := getParam(r, "host")
		ip := &models.IP{
			Address: host,
		}
		a.db.Where(ip).
			Preload("Country").
			First(ip)
		country = ip.Country.ISOCode
	}
	sendJSON(w, r, countryResult(country))
}

func (a *Application) apiUserAgent(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, r, userAgentResult(r))
}

func (a *Application) apiLanguage(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, r, languageResult(r))
}

func (a *Application) apiDetail(w http.ResponseWriter, r *http.Request) {
	host := getParam(r, "host")

	ip, err := a.resolveIP(host)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	sendJSON(w, r, ip)
}

func (a *Application) apiMe(w http.ResponseWriter, r *http.Request) {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	if host == "" {
		host = r.RemoteAddr
	}

	ip, err := a.resolveIP(host)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	sendJSON(w, r, &MeResponse{
		Ip:        ip,
		Language:  languageResult(r),
		UserAgent: userAgentResult(r),
	})
}

func (a *Application) apiStatistic(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, r, a.statistic)
}

func (a *Application) apiSearch(w http.ResponseWriter, r *http.Request) {
	var sir models.IP

	err := json.NewDecoder(r.Body).Decode(&sir)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ip := &models.IP{}

	if sir.Address != "" {
		if strings.Contains(sir.Address, "/") == false {
			_ip, err := a.resolveIP(sir.Address)
			if err == nil {
				sir.Address = _ip.Address
			}
		}
	}

	db := a.db.Model(ip)
	if sir.Address != "" {
		db = db.Where("address LIKE ?", "%"+sir.Address+"%")
	}

	if sir.IsAnonymous {
		db = db.Where("is_anonymous = ?", true)
	}
	if sir.IsAnonymousProxy {
		db = db.Where("is_anonymous_proxy = ?", true)
	}
	if sir.IsAnonymousVPN {
		db = db.Where("is_anonymous_vpn = ?", true)
	}
	if sir.IsHostingProvider {
		db = db.Where("is_hosting_provider = ?", true)
	}
	if sir.IsPublicProxy {
		db = db.Where("is_public_proxy = ?", true)
	}
	if sir.IsSatelliteProvider {
		db = db.Where("is_satellite_provider = ?", true)
	}
	if sir.IsTorExitNode {
		db = db.Where("is_tor_exit_node = ?", true)
	}

	if sir.ProxyType != "" {
		db = db.Where("proxy_type LIKE ?", "%"+sir.ProxyType+"%")
	}
	if sir.Type != "" {
		db = db.Where("type LIKE ?", "%"+sir.Type+"%")
	}
	if sir.StaticIpScore != "" {
		db = db.Where("static_ip_score LIKE ?", "%"+sir.StaticIpScore+"%")
	}
	if sir.Threat != "" {
		db = db.Where("threat LIKE ?", "%"+sir.Threat+"%")
	}
	if sir.UserCount != "" {
		db = db.Where("user_count = ?", "%"+sir.UserCount+"%")
	}

	if sir.Latitude > 0 && sir.Longitude > 0 {
		if sir.AccuracyRadius <= 0 {
			sir.AccuracyRadius = 100
		}

		center := &point{
			x: float64(sir.Latitude),
			y: float64(sir.Longitude),
		}
		mult := 1.0 // mult = 1.1; is more reliable
		radius := float64(sir.AccuracyRadius)

		p1 := calculateDerivedPosition(center, mult*radius, 0)
		p2 := calculateDerivedPosition(center, mult*radius, 90)
		p3 := calculateDerivedPosition(center, mult*radius, 180)
		p4 := calculateDerivedPosition(center, mult*radius, 270)

		// 6371.04 => miles
		db = db.Where(
			"latitude > ? AND latitude < ? AND longitude > ? AND longitude < ?",
			float32(p3.x),
			float32(p1.x),
			float32(p4.y),
			float32(p2.y),
		)
		/*
			This can be enabled if ACOS is supported
			db = db.Where(
				"(3959 * ACOS(COS(RADIANS(?)) * cos(RADIANS(latitude)) * COS(RADIANS(longitude)-RADIANS(?))+SIN(RADIANS(?)) * SIN(RADIANS(latitude)))) < ?",
				sir.Latitude,
				sir.Longitude,
				sir.Latitude,
				sir.AccuracyRadius,
			)
		*/

	} else {
		if sir.Latitude > 0 {
			db = db.Where("latitude = ?", sir.Latitude)
		}
		if sir.Longitude > 0 {
			db = db.Where("longitude = ?", sir.Longitude)
		}
		if sir.AccuracyRadius > 0 {
			db = db.Where("accuracy_radius = ?", sir.AccuracyRadius)
		}
	}

	matrix := map[string]struct {
		valid    bool
		operator string
		value    interface{}
	}{
		"is_in_european_union": {
			operator: "=",
			value:    sir.Country.IsInEuropeanUnion,
			valid:    sir.Country.IsInEuropeanUnion,
		},
		"iso_code": {
			operator: "LIKE",
			value:    "%" + sir.Country.ISOCode + "%",
			valid:    sir.Country.ISOCode != "",
		},
		"name": {
			operator: "LIKE",
			value:    "%" + sir.Country.Name + "%",
			valid:    sir.Country.Name != "",
		},
	}

	q := ""
	values := make([]interface{}, 0)
	for key, m := range matrix {
		if m.valid {
			if q != "" {
				q += " AND"
			}
			q += fmt.Sprintf(" %s %s ?", key, m.operator)
			values = append(values, m.value)
		}
	}
	if q != "" {
		db = db.Where("country_id IN (SELECT id FROM countries WHERE "+q+")", values...)
	}

	matrix = map[string]struct {
		valid    bool
		operator string
		value    interface{}
	}{
		"code": {
			operator: "LIKE",
			value:    "%" + sir.Country.Continent.Code + "%",
			valid:    sir.Country.Continent.Code != "",
		},
		"name": {
			operator: "LIKE",
			value:    "%" + sir.Country.Continent.Name + "%",
			valid:    sir.Country.Continent.Name != "",
		},
	}
	q = ""
	values = make([]interface{}, 0)
	for key, m := range matrix {
		if m.valid {
			if q != "" {
				q += " AND"
			}
			q += fmt.Sprintf(" %s %s ?", key, m.operator)
			values = append(values, m.value)
		}
	}
	if q != "" {
		db = db.Where("country_id IN (SELECT id FROM countries WHERE continent_id IN (SELECT id FROM continents WHERE "+q+"))", values...)
	}

	matrix = map[string]struct {
		valid    bool
		operator string
		value    interface{}
	}{
		"metro_code": {
			operator: "=",
			value:    sir.City.MetroCode,
			valid:    sir.City.MetroCode > 0,
		},
		"population_density": {
			operator: "=",
			value:    sir.City.PopulationDensity,
			valid:    sir.City.PopulationDensity > 0,
		},
		"time_zone": {
			operator: "LIKE",
			value:    "%" + sir.City.TimeZone + "%",
			valid:    sir.City.TimeZone != "",
		},
		"name": {
			operator: "LIKE",
			value:    "%" + sir.City.Name + "%",
			valid:    sir.City.Name != "",
		},
	}
	q = ""
	values = make([]interface{}, 0)
	for key, m := range matrix {
		if m.valid {
			if q != "" {
				q += " AND"
			}
			q += fmt.Sprintf(" %s %s ?", key, m.operator)
			values = append(values, m.value)
		}
	}
	if q != "" {
		db = db.Where("city_id IN (SELECT id FROM cities WHERE "+q+")", values...)
	}

	if ld := len(sir.City.Regions); ld > 0 {
		for _, region := range sir.City.Regions {

			matrix = map[string]struct {
				valid    bool
				operator string
				value    interface{}
			}{
				"name": {
					operator: "LIKE",
					value:    "%" + region.Name + "%",
					valid:    region.Name != "",
				},
				"code": {
					operator: "LIKE",
					value:    "%" + region.Code + "%",
					valid:    region.Code != "",
				},
			}
			values = make([]interface{}, ld)
			q = ""
			for key, m := range matrix {
				if m.valid {
					if q != "" {
						q += " AND"
					}
					q += fmt.Sprintf(" %s %s ?", key, m.operator)
					values = append(values, m.value)
				}
			}
			db = db.Where("city_id IN (SELECT city_id FROM city_regions WHERE region_id IN (SELECT id FROM regions WHERE "+q+"))", values)
		}
	}

	if sir.Postal.Zip != "" {
		db = db.Where("postal_id IN (SELECT id FROM postals WHERE zip LIKE ?)", "%"+sir.Postal.Zip+"%")
	}
	if sir.Isp.Name != "" {
		db = db.Where("isp_id IN (SELECT id FROM isps WHERE name LIKE ?)", "%"+sir.Isp.Name+"%")
	}
	if sir.Organization.Name != "" {
		db = db.Where("organization_id IN (SELECT id FROM organizations WHERE name LIKE ?)", "%"+sir.Organization.Name+"%")
	}

	matrix = map[string]struct {
		valid    bool
		operator string
		value    interface{}
	}{
		"network": {
			operator: "LIKE",
			value:    "%" + sir.Network.Network + "%",
			valid:    sir.Network.Network != "",
		},
		"domain": {
			operator: "LIKE",
			value:    "%" + sir.Network.Domain + "%",
			valid:    sir.Network.Domain != "",
		},
	}
	q = ""
	values = make([]interface{}, 0)
	for key, m := range matrix {
		if m.valid {
			if q != "" {
				q += " AND"
			}
			q += fmt.Sprintf(" %s %s ?", key, m.operator)
			values = append(values, m.value)
		}
	}
	if q != "" {
		db = db.Where("network_id IN (SELECT id FROM networks WHERE "+q+")", values...)
	}

	matrix = map[string]struct {
		valid    bool
		operator string
		value    interface{}
	}{
		"number": {
			operator: "=",
			value:    sir.AutonomousSystem.Number,
			valid:    sir.AutonomousSystem.Number > 0,
		},
		"name": {
			operator: "LIKE",
			value:    "%" + sir.AutonomousSystem.Name + "%",
			valid:    sir.AutonomousSystem.Name != "",
		},
	}
	q = ""
	values = make([]interface{}, 0)
	for key, m := range matrix {
		if m.valid {
			if q != "" {
				q += " AND"
			}
			q += fmt.Sprintf(" %s %s ?", key, m.operator)
			values = append(values, m.value)
		}
	}
	if q != "" {
		db = db.Where("autonomous_system_id IN (SELECT id FROM autonomous_systems WHERE "+q+")", values...)
	}

	if ld := len(sir.Domains); ld > 0 {
		values = make([]interface{}, ld)
		for i, d := range sir.Domains {
			values[i] = d.Name
		}
		db = db.Where("id IN (SELECT ip_id FROM domain_ips WHERE domain_id IN (SELECT id FROM domains WHERE name IN ?))", values)
	}

	db.Preload("Country.Continent").
		Preload("City.Regions").
		Preload("Postal").
		Preload("Isp").
		Preload("Network").
		Preload("Organization").
		Preload("Domains").
		Preload("AutonomousSystem")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit > 1000 {
		limit = 1000
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	// @TODO: validate the sort argument - should not contain an unknown table key
	// r.URL.Query().Get("sort")
	p := paginator.NewGormPaginator(&paginator.DefaultPaginator{
		Limit:         limit,
		Page:          page + 1,
		SortKey:       "id",
		SortDirection: r.URL.Query().Get("direction"),
	})

	var vs []*models.IP
	if err := p.Paginate(db.Preload("Country.Continent").
		Preload("City.Regions").
		Preload("Postal").
		Preload("Isp").
		Preload("Network").
		Preload("Organization").
		Preload("Domains").
		Preload("AutonomousSystem"), &vs); err != nil {
		fmt.Printf("[error] database query failed: %s\n", err.Error())
		http.NotFound(w, r)
		return
	}

	sendJSON(w, r, p)
}

func (a *Application) resolveIP(host string, domains ...string) (*models.IP, error) {
	ip := &models.IP{
		Address: host,
	}

	ipa := net.ParseIP(host)
	if ipa == nil && len(domains) == 0 {
		// Given host is not an IP address. Perhaps a domain name?
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return nil, errors.New("ip not found")
		}
		domain := host
		for _, _ip := range ips {
			host = _ip.String()
			ip, err = a.resolveIP(host, domain)
		}
		return ip, err
	}

	a.db.Where(ip).
		Preload("Country.Continent").
		Preload("City.Regions").
		Preload("Postal").
		Preload("Isp").
		Preload("Network").
		Preload("Organization").
		Preload("Domains").
		Preload("AutonomousSystem").
		Find(ip)

	for fails := 0; fails < 3 && ip.ID == 0; fails++ {
		if ip.ID == 0 {
			IPAddress := net.ParseIP(host)
			a.QueueRecord(&Record{
				ip:      IPAddress,
				fails:   0,
				domains: domains,
			})
			time.Sleep(3 * time.Second)
			a.db.Where(ip).
				Preload("Country.Continent").
				Preload("City.Regions").
				Preload("Postal").
				Preload("Isp").
				Preload("Network").
				Preload("Organization").
				Preload("Domains").
				Preload("AutonomousSystem").
				Find(ip)
		}
	}
	if len(domains) > 0 && ip.ID > 0 {
		for _, name := range domains {
			d := &models.Domain{
				Name: name,
			}
			if ip.HasDomain(d) == false {
				if domain, err := models.UpdateDomain(a.db, d); err != nil {
					return nil, err
				} else {
					ip.Domains = append(ip.Domains, *domain)
				}
			}
		}

		if err := a.db.Model(ip).Association("Domains").Replace(ip.Domains); err != nil {
			return nil, err
		}

		return a.resolveIP(host)
	}

	return ip, nil
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

func staticGeoInfo(code string) gountries.Country {
	query := gountries.New()
	country, err := query.FindCountryByAlpha(code)
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

	return country
}

func languageResult(r *http.Request) *LanguageResponse {
	t, qq, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))

	plang := getMostPreferredLanguage(t, qq)
	base, _ := plang.Base()
	_region, _ := plang.Region()

	return &LanguageResponse{
		Language: base.String(),
		Region:   _region.String(),
		Tag:      plang.String(),
	}
}

func userAgentResult(r *http.Request) *UserAgentResponse {
	userAgent := useragent.Parse(r.Header.Get("User-Agent"))

	return &UserAgentResponse{
		Name:      userAgent.Name,
		Version:   userAgent.Version,
		OS:        userAgent.OS,
		OSVersion: userAgent.OSVersion,
		Device:    userAgent.Device,
		Mobile:    userAgent.Mobile,
		Tablet:    userAgent.Tablet,
		Desktop:   userAgent.Desktop,
		Bot:       userAgent.Bot,
	}
}

func countryResult(country string) *CountryResponse {
	geo := staticGeoInfo(country)

	return &CountryResponse{
		Name: struct {
			gountries.BaseLang
			Native map[string]gountries.BaseLang `json:"native"`
		}{
			BaseLang: geo.Name.BaseLang,
			Native:   geo.Name.Native,
		},
		EuMember:     geo.EuMember,
		LandLocked:   geo.LandLocked,
		Nationality:  geo.Nationality,
		TLDs:         geo.TLDs,
		Languages:    geo.Languages,
		Translations: geo.Translations,
		Currencies:   geo.Currencies,
		Borders:      geo.Borders,
		Codes: Codes{
			Alpha2:              geo.Codes.Alpha2,
			Alpha3:              geo.Codes.Alpha3,
			CIOC:                geo.Codes.CIOC,
			CCN3:                geo.Codes.CCN3,
			CallingCodes:        geo.Codes.CallingCodes,
			InternationalPrefix: geo.Codes.InternationalPrefix,
		},
		Geo: Geo{
			Region:    geo.Geo.Region,
			SubRegion: geo.Geo.SubRegion,
			Continent: geo.Geo.Continent,
			Capital:   geo.Geo.Capital,
			Area:      geo.Geo.Area,
		},
		Coordinates: Coordinates{
			MinLongitude: geo.Coordinates.MinLongitude,
			MinLatitude:  geo.Coordinates.MinLatitude,
			MaxLongitude: geo.Coordinates.MaxLongitude,
			MaxLatitude:  geo.Coordinates.MaxLatitude,
			Latitude:     geo.Coordinates.Latitude,
			Longitude:    geo.Coordinates.Longitude,
		},
	}
}
