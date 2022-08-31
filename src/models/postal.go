package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Postal struct {
	ID     uint `gorm:"primarykey" json:"-"`
	CityID uint `gorm:"not null" json:"-"`

	Zip string `gorm:"index" json:"zip"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	City City `json:"-"`
	Ips  []IP `json:"-"`
}

func UpdatePostal(db *gorm.DB, ip *Postal) (*Postal, error) {
	m := &Postal{}
	if err := db.Where(&Postal{Zip: ip.Zip}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(ip).Error; err != nil {
			return nil, err
		}
		return ip, nil
	}

	return m, nil
}
