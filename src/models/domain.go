package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Domain struct {
	ID uint `gorm:"primarykey" json:"-"`

	Name string `gorm:"index" json:"name"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Ips []IP `gorm:"many2many:domain_ips;" json:"-"`
}

func UpdateDomain(db *gorm.DB, domain *Domain) (*Domain, error) {
	m := &Domain{}
	if err := db.Where(&Domain{Name: domain.Name}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(domain).Error; err != nil {
			return nil, err
		}
		return domain, nil
	}

	return m, nil
}
