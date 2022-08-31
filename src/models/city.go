package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type City struct {
	ID        uint `gorm:"primarykey" json:"-"`
	CountryID uint `gorm:"not null;index" json:"-"`
	RegionID  uint `gorm:"not null" json:"-"`

	Name              string `gorm:"index" json:"name"`
	MetroCode         uint   `json:"metro_code"`
	TimeZone          string `json:"time_zone"`
	PopulationDensity uint   `json:"population_density"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Regions []Region `gorm:"many2many:city_regions;" json:"regions"`
	Country Country  `json:"-"`
	Postals []Postal `json:"-"`
	Ips     []IP     `json:"-"`
}

func UpdateCity(db *gorm.DB, city *City) (*City, error) {
	m := &City{}
	if err := db.Where(&City{Name: city.Name, CountryID: city.CountryID}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(city).Error; err != nil {
			return nil, err
		}
		return city, nil
	}

	m.MetroCode = city.MetroCode
	m.TimeZone = city.TimeZone
	m.PopulationDensity = city.PopulationDensity
	if city.Name != "" {
		m.Name = city.Name
	}
	if err := db.Save(m).Error; err != nil {
		return nil, err
	}

	return m, nil
}
