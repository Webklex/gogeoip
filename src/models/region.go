package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Region struct {
	ID        uint `gorm:"primarykey" json:"-"`
	CountryID uint `gorm:"index" json:"-"`

	Code string `json:"code"`
	Name string `gorm:"index" json:"name"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Country Country `json:"-"`
	Cities  []City  `gorm:"many2many:city_regions;" json:"-"`
}

func UpdateRegion(db *gorm.DB, region *Region) (*Region, error) {
	m := &Region{}
	if err := db.Where(&Region{Name: region.Name, CountryID: region.CountryID}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(region).Error; err != nil {
			return nil, err
		}
		return region, nil
	} else if m.CountryID != region.CountryID || m.Code != region.Code {
		m.CountryID = region.CountryID
		m.Code = region.Code
		if err := db.Save(m).Error; err != nil {
			return nil, err
		}
	}

	return m, nil
}
