package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Organization struct {
	ID uint `gorm:"primarykey" json:"-"`

	Name string `gorm:"index" json:"name"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Ips []IP `json:"-"`
}

func UpdateOrganization(db *gorm.DB, org *Organization) (*Organization, error) {
	m := &Organization{}
	if err := db.Where(&Organization{Name: org.Name}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(org).Error; err != nil {
			return nil, err
		}
		return org, nil
	}

	return m, nil
}
