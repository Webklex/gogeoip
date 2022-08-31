package app

type LegacyResponseRecord struct {
	Network  LegacyNetworkRecord  `json:"network"`
	Location LegacyLocationRecord `json:"location"`
	System   SystemRecord         `json:"system,omitempty"`
	User     UserRecord           `json:"user,omitempty"`
}

type LegacyLocationRecord struct {
	RegionCode     string              `json:"region_code"`
	RegionName     string              `json:"region_name"`
	City           string              `json:"city"`
	ZipCode        string              `json:"zip_code"`
	TimeZone       string              `json:"time_zone"`
	Longitude      float64             `json:"longitude"`
	Latitude       float64             `json:"latitude"`
	AccuracyRadius uint                `json:"accuracy_radius"`
	MetroCode      uint                `json:"metro_code"`
	Country        LegacyCountryRecord `json:"country"`
}

type LegacyCountryRecord struct {
	Code                string                 `json:"code"`
	CIOC                string                 `json:"cioc"`
	CCN3                string                 `json:"ccn3"`
	CallCode            []string               `json:"call_code"`
	InternationalPrefix string                 `json:"international_prefix"`
	Capital             string                 `json:"capital"`
	Name                string                 `json:"name"`
	FullName            string                 `json:"full_name"`
	Area                float64                `json:"area"`
	Borders             []string               `json:"borders"`
	Latitude            float64                `json:"latitude"`
	Longitude           float64                `json:"longitude"`
	MaxLatitude         float64                `json:"max_latitude"`
	MaxLongitude        float64                `json:"max_longitude"`
	MinLatitude         float64                `json:"min_latitude"`
	MinLongitude        float64                `json:"min_longitude"`
	Currency            []LegacyCurrencyRecord `json:"currency"`
	Continent           LegacyContinentRecord  `json:"continent"`
}

type LegacyCurrencyRecord struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type LegacyContinentRecord struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	SubRegion string `json:"sub_region"`
}

type LegacyNetworkRecord struct {
	IP        string         `json:"ip"`
	AS        LegacyASRecord `json:"as"`
	Isp       string         `json:"isp"`
	Domain    string         `json:"domain"`
	Tld       []string       `json:"tld"`
	Bot       bool           `json:"bot"`
	Tor       bool           `json:"tor"`
	Proxy     bool           `json:"proxy"`
	ProxyType string         `json:"proxy_type"`
	LastSeen  uint           `json:"last_seen"`
	UsageType string         `json:"usage_type"`
}

type LegacyASRecord struct {
	Number uint   `json:"number"`
	Name   string `json:"name"`
}

type UserRecord struct {
	Language *LanguageRecord `json:"language"`
}

type LanguageRecord struct {
	Language string `json:"language"`
	Region   string `json:"region"`
	Tag      string `json:"tag"`
}

type SystemRecord struct {
	OS        string `json:"os"`
	Browser   string `json:"browser"`
	Version   string `json:"version"`
	OSVersion string `json:"os_version"`
	Device    string `json:"device"`
	Mobile    bool   `json:"mobile"`
	Tablet    bool   `json:"tablet"`
	Desktop   bool   `json:"desktop"`
}
