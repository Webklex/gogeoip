package app

import (
	"fmt"
	"github.com/ammario/ipisp/v2"
	"github.com/webklex/gogeoip/src/api"
	"github.com/webklex/gogeoip/src/models"
	"net"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	ip      net.IP
	fails   int
	domains []string

	mm   *api.MaxMindDefaultQuery
	mma  *api.MaxMindASNDefaultQuery
	i2l  *api.Ip2LocationResponse
	ispl *ipisp.Response
}

func (a *Application) importRecord(record *Record) error {
	if record.fails > 10 {
		return fmt.Errorf("[error] failed to import: %s", record.ip.String())
	} else if record.fails > 0 {
		timeout := time.Second * time.Duration(record.fails)
		fmt.Printf("[error] failed to import: %s. trying again in %s.\n", record.ip.String(), timeout.String())
		time.Sleep(timeout)
	} else {
		//fmt.Printf(":")
	}
	var err error

	continent := &models.Continent{
		Code: record.mm.Continent.Code,
		Name: record.mm.Continent.Code,
	}
	if cn, _ := record.mm.Continent.Names["en"]; cn != "" {
		continent.Name = cn
	}
	if continent.Code == "" {
		continent.Code = record.mm.Country.ContinentCode
	}
	continent, err = models.UpdateContinent(a.db, continent)
	if err, reload := a.handleImportError(err); err != nil {
		return err
	} else if reload {
		record.fails++
		go a.QueueRecord(record)
		return nil
	}

	country := &models.Country{
		ContinentID:       continent.ID,
		ISOCode:           record.mm.Country.ISOCode,
		IsInEuropeanUnion: record.mm.Country.IsInEuropeanUnion,
		Name:              record.i2l.CountryLong,
	}
	if record.i2l.CountryShort != "" && country.ISOCode == "" {
		country.ISOCode = record.i2l.CountryShort
	}
	if cn, _ := record.mm.Country.Names["en"]; cn != "" {
		country.Name = cn
	}

	country, err = models.UpdateCountry(a.db, country)
	if err, reload := a.handleImportError(err); err != nil {
		return err
	} else if reload {
		record.fails++
		go a.QueueRecord(record)
		return nil
	}

	var regions []*models.Region
	if record.i2l.Region != "" {
		regions = append(regions, &models.Region{
			CountryID: country.ID,
			Name:      record.i2l.Region,
		})
	}
	if len(record.mm.Region) > 0 {
		for _, r := range record.mm.Region {
			regions = append(regions, &models.Region{
				CountryID: country.ID,
				Code:      r.ISOCode,
				Name:      r.Names["en"],
			})
		}
	}

	for i, r := range regions {
		regions[i], err = models.UpdateRegion(a.db, r)
		if err, reload := a.handleImportError(err); err != nil {
			return err
		} else if reload {
			record.fails++
			go a.QueueRecord(record)
			return nil
		}
	}

	city := &models.City{
		CountryID:         country.ID,
		Name:              record.i2l.City,
		MetroCode:         record.mm.Location.MetroCode,
		TimeZone:          record.mm.Location.TimeZone,
		PopulationDensity: record.mm.Location.PopulationDensity,
	}

	if cn, _ := record.mm.City.Names["en"]; cn != "" {
		city.Name = cn
	}
	city, err = models.UpdateCity(a.db, city)
	if err, reload := a.handleImportError(err); err != nil {
		return err
	} else if reload {
		record.fails++
		go a.QueueRecord(record)
		return nil
	}
	if err, reload := a.handleImportError(a.db.Model(city).Association("Regions").Replace(regions)); err != nil {
		return err
	} else if reload {
		return a.importRecord(record)
	}

	postal := &models.Postal{
		CityID: city.ID,
		Zip:    record.mm.Postal.Code,
	}
	if postal.Zip != "" {
		postal, err = models.UpdatePostal(a.db, postal)
		if err, reload := a.handleImportError(err); err != nil {
			return err
		} else if reload {
			record.fails++
			go a.QueueRecord(record)
			return nil
		}
	}

	isp := &models.ISP{
		Name: record.mm.Traits.ISP,
	}
	if isp.Name == "" {
		isp.Name = record.i2l.ISP
	}
	if isp.Name == "" && record.ispl != nil {
		isp.Name = record.ispl.ISPName
	}
	if isp.Name != "" {
		isp, err = models.UpdateISP(a.db, isp)
		if err, reload := a.handleImportError(err); err != nil {
			return err
		} else if reload {
			record.fails++
			go a.QueueRecord(record)
			return nil
		}
	}

	network := &models.Network{
		Network: record.mm.Traits.Network,
		Domain:  record.mm.Traits.Domain,
	}
	if network.Domain == "" {
		network.Domain = record.i2l.Domain
	}
	if network.Domain != "" || network.Network != "" {
		network, err = models.UpdateNetwork(a.db, network)
		if err, reload := a.handleImportError(err); err != nil {
			return err
		} else if reload {
			record.fails++
			go a.QueueRecord(record)
			return nil
		}
	}

	org := &models.Organization{
		Name: record.mm.Traits.Organization,
	}
	if org.Name != "" {
		org, err = models.UpdateOrganization(a.db, org)
		if err, reload := a.handleImportError(err); err != nil {
			return err
		} else if reload {
			record.fails++
			go a.QueueRecord(record)
			return nil
		}
	}

	domains := make([]models.Domain, 0)
	if record.domains != nil && len(record.domains) > 0 {
		for _, name := range record.domains {
			domain, err := models.UpdateDomain(a.db, &models.Domain{
				Name: name,
			})
			if err, reload := a.handleImportError(err); err != nil {
				return err
			} else if reload {
				record.fails++
				go a.QueueRecord(record)
				return nil
			}
			domains = append(domains, *domain)
		}
	}

	as := &models.AutonomousSystem{
		Name:   record.mm.Traits.AutonomousSystemOrganization,
		Number: record.mm.Traits.AutonomousSystemNumber,
	}
	if as.Name == "" {
		as.Name = record.i2l.AS
	}
	if as.Name == "" {
		as.Name = record.mma.AutonomousSystemOrganization
	}
	if as.Number == 0 {
		as.Number = record.mma.AutonomousSystemNumber
	}
	if as.Number == 0 {
		asn, _ := strconv.Atoi(record.i2l.ASN)
		as.Number = uint(asn)
	}
	if as.Number == 0 && record.ispl != nil {
		as.Number = uint(record.ispl.ASN)
	}
	if as.Number != 0 || as.Name != "" {
		as, err = models.UpdateAutonomousSystem(a.db, as)
		if err, reload := a.handleImportError(err); err != nil {
			return err
		} else if reload {
			record.fails++
			go a.QueueRecord(record)
			return nil
		}
	}

	userType := record.i2l.UsageType
	if userType == "" {
		switch record.mm.Traits.UserType {
		case "business":
			userType = "COM"
		case "cafe":
			userType = "CAF"
		case "cellular":
			userType = "MOB"
		case "college":
			userType = "EDU"
		case "consumer_privacy_network":
			userType = "CPN"
		case "content_delivery_network":
			userType = "CDN"
		case "government":
			userType = "GOV"
		case "hosting":
			userType = "DCH"
		case "library":
			userType = "LIB"
		case "military":
			userType = "MIL"
		case "residential":
			userType = "RDL"
		case "router":
			userType = "RTR"
		case "school":
			userType = "EDU"
		case "search_engine_spider":
			userType = "SES"
		case "traveler":
			userType = "TRA"
		}
	}

	ip := &models.IP{
		IspID:               isp.ID,
		NetworkID:           network.ID,
		CountryID:           country.ID,
		CityID:              city.ID,
		PostalID:            postal.ID,
		OrganizationID:      org.ID,
		AutonomousSystemID:  as.ID,
		Address:             record.ip.String(),
		IsTorExitNode:       a.Tor.Lookup(record.ip),
		IsAnonymous:         record.mm.Traits.IsAnonymous,
		IsAnonymousProxy:    record.mm.Traits.IsAnonymousProxy,
		IsAnonymousVPN:      record.mm.Traits.IsAnonymousVPN,
		IsHostingProvider:   record.mm.Traits.IsHostingProvider,
		IsPublicProxy:       record.mm.Traits.IsPublicProxy,
		IsSatelliteProvider: record.mm.Traits.IsSatelliteProvider,
		ProxyType:           record.i2l.ProxyType,
		Type:                userType,

		StaticIpScore: record.mm.Traits.StaticIpScore,
		Threat:        record.i2l.Threat,
		UserCount:     record.mm.Traits.UserCount,

		Latitude:       float32(record.mm.Location.Latitude),
		Longitude:      float32(record.mm.Location.Longitude),
		AccuracyRadius: record.mm.Location.AccuracyRadius,
		LastSeen:       record.i2l.LastSeen,
	}
	if ip.Threat == "NOT SUPPORTED" {
		ip.Threat = ""
	}
	if ip.IsTorExitNode == false {
		ip.IsTorExitNode = record.mm.Traits.IsTorExitNode
	}
	if record.i2l.IsProxy && !ip.IsAnonymousProxy && !ip.IsPublicProxy {
		if record.i2l.ProxyType == "PUB" {
			ip.IsPublicProxy = true
		} else {
			ip.IsAnonymousProxy = true
		}
	}

	ip, err = models.UpdateIP(a.db, ip)
	if err, reload := a.handleImportError(err); err != nil {
		return err
	} else if reload {
		record.fails++
		go a.QueueRecord(record)
		return nil
	}
	domains = append(domains, ip.Domains...)
	if err, reload := a.handleImportError(a.db.Model(ip).Association("Domains").Replace(domains)); err != nil {
		return err
	} else if reload {
		return a.importRecord(record)
	}

	return nil
}

func (a *Application) handleImportError(err error) (error, bool) {
	if err == nil {
		return nil, false
	}

	serr := fmt.Sprintf("%s", err.Error())
	if serr == "database is locked" || strings.Contains(serr, "UNIQUE constraint failed") {
		return nil, true
	}

	return err, false
}
