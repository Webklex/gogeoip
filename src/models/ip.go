package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type IP struct {
	ID                 uint `gorm:"primarykey" json:"-"`
	IspID              uint `gorm:"not null" json:"-"`
	NetworkID          uint `gorm:"not null" json:"-"`
	CountryID          uint `gorm:"not null" json:"-"`
	CityID             uint `gorm:"not null" json:"-"`
	PostalID           uint `gorm:"not null" json:"-"`
	OrganizationID     uint `gorm:"not null" json:"-"`
	AutonomousSystemID uint `gorm:"not null" json:"-"`

	Address string `gorm:"not null;unique;index" json:"ip"`

	IsAnonymous         bool `json:"is_anonymous"`
	IsAnonymousProxy    bool `json:"is_anonymous_proxy"`
	IsAnonymousVPN      bool `json:"is_anonymous_vpn"`
	IsHostingProvider   bool `json:"is_hosting_provider"`
	IsPublicProxy       bool `json:"is_public_proxy"`
	IsSatelliteProvider bool `json:"is_satellite_provider"`
	IsTorExitNode       bool `json:"is_tor_exit_node"`

	// List of available ProxyType values:
	// VPN		Anonymizing VPN services. These services offer users a publicly accessible VPN for the purpose of
	//			hiding their IP address.
	// TOR		Tor Exit Nodes. The Tor Project is an open network used by those who wish to maintain anonymity.
	// DCH		Hosting Provider, Data Center or Content Delivery Network. Since hosting providers and data centers
	//			can serve to provide anonymity, the Anonymous IP database flags IP addresses associated with them.
	// PUB*		Public Proxies. These are services which make connection requests on a user's behalf. Proxy server
	//			software can be configured by the administrator to listen on some specified port. These differ from
	//			VPNs in that the proxies usually have limited functions compare to VPNs.
	// WEB		Web Proxies. These are web services which make web requests on a user's behalf. These differ from VPNs
	//			or Public Proxies in that they are simple web-based proxies rather than operating at the IP address and
	//			other ports level.
	// SES		Search Engine Robots. These are services which perform crawling or scraping to a website, such as, the
	//			search engine spider or bots engine.
	// RES		Residential proxies. These services offer users proxy connections through residential ISP with or
	//			without consents of peers to share their idle resources.
	ProxyType string `json:"proxy_type"`

	// List of available Type values:
	// (COM) Commercial
	// (ORG) Organization
	// (GOV) Government
	// (MIL) Military
	// (EDU) University/College/School
	// (LIB) Library
	// (CDN) Content Delivery Network
	// (ISP) Fixed Line ISP
	// (MOB) Mobile ISP
	// (DCH) Data Center/Web Hosting/Transit
	// (SES) Search Engine Spider
	// (RSV) Reserved
	// (TRA) Traveler
	// (RTR) Router
	// (RDL) Residential
	// (CPN) Consumer Privacy Network
	Type string `json:"type"`

	StaticIpScore string `json:"score"`
	Threat        string `json:"threat"`
	UserCount     string `json:"user_count"`

	Latitude       float32 `gorm:"type:decimal(10,2);" json:"latitude"`
	Longitude      float32 `gorm:"type:decimal(10,2);" json:"longitude"`
	AccuracyRadius uint    `json:"accuracy_radius"`

	LastSeen int `json:"last_seen"`

	Country          Country          `json:"country"`
	City             City             `json:"city"`
	Postal           Postal           `json:"postal"`
	Isp              ISP              `json:"isp"`
	Network          Network          `json:"network"`
	Organization     Organization     `json:"organization"`
	Domains          []Domain         `gorm:"many2many:domain_ips;" json:"domains"`
	AutonomousSystem AutonomousSystem `json:"autonomous_system"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func UpdateIP(db *gorm.DB, ip *IP) (*IP, error) {
	m := &IP{}
	if err := db.Where(&IP{Address: ip.Address}).
		Preload("Domains").First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(ip).Error; err != nil {
			return nil, err
		}
		return ip, nil
	} else if m.CountryID != ip.CountryID || m.CityID != ip.CityID || m.IsAnonymous != ip.IsAnonymous || m.IsAnonymousProxy != ip.IsAnonymousProxy || m.IsAnonymousVPN != ip.IsAnonymousVPN || m.IsHostingProvider != ip.IsHostingProvider || m.IsPublicProxy != ip.IsPublicProxy || m.IsSatelliteProvider != ip.IsSatelliteProvider || m.IsTorExitNode != ip.IsTorExitNode || m.Latitude != ip.Latitude || m.Longitude != ip.Longitude || m.AccuracyRadius != ip.AccuracyRadius || m.StaticIpScore != ip.StaticIpScore || m.UserCount != ip.UserCount {
		m.CountryID = ip.CountryID
		m.CityID = ip.CityID
		m.IsAnonymous = ip.IsAnonymous
		m.IsAnonymousProxy = ip.IsAnonymousProxy
		m.IsAnonymousVPN = ip.IsAnonymousVPN
		m.IsHostingProvider = ip.IsHostingProvider
		m.IsPublicProxy = ip.IsPublicProxy
		m.IsSatelliteProvider = ip.IsSatelliteProvider
		m.IsTorExitNode = ip.IsTorExitNode
		m.Latitude = ip.Latitude
		m.Longitude = ip.Longitude
		m.AccuracyRadius = ip.AccuracyRadius
		m.StaticIpScore = ip.StaticIpScore
		m.UserCount = ip.UserCount

		if err := db.Save(m).Error; err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *IP) HasDomain(d *Domain) bool {
	for _, id := range m.Domains {
		if id.Name == d.Name {
			return true
		}
	}
	return false
}
