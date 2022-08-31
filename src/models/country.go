package models

import (
	"errors"
	"github.com/pariz/gountries"
	"gorm.io/gorm"
	"time"
)

type Country struct {
	ID          uint `gorm:"primarykey" json:"-"`
	ContinentID uint `gorm:"not null;index" json:"-"`

	ISOCode           string `gorm:"index" json:"code"`
	IsInEuropeanUnion bool   `json:"european_member"`
	Name              string `json:"name"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Continent Continent `json:"continent"`
	Cities    []City    `json:"-"`
	Ips       []IP      `json:"-"`
}

func UpdateCountry(db *gorm.DB, country *Country) (*Country, error) {
	m := &Country{}
	if err := db.Where(&Country{ISOCode: country.ISOCode, ContinentID: country.ContinentID}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if m.Name == "" {
		query := gountries.New()
		c, err := query.FindCountryByAlpha(country.ISOCode)
		if err == nil {
			if c.Name.Common != "" {
				country.Name = c.Name.Common
			}
		}
	}

	if m.ID <= 0 {
		if err := db.Create(country).Error; err != nil {
			return nil, err
		}
		return country, nil
	} else if m.Name == "" || m.IsInEuropeanUnion != country.IsInEuropeanUnion || m.ContinentID != country.ContinentID {
		if m.Name == "" {
			m.Name = country.Name
		}
		m.IsInEuropeanUnion = country.IsInEuropeanUnion
		m.ContinentID = country.ContinentID
		if err := db.Save(m).Error; err != nil {
			return nil, err
		}
	}

	return m, nil
}
