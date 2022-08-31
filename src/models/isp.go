package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type ISP struct {
	ID uint `gorm:"primarykey" json:"-"`

	Name string `gorm:"index" json:"name"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Ips []IP `json:"-"`
}

func UpdateISP(db *gorm.DB, isp *ISP) (*ISP, error) {
	m := &ISP{}
	if err := db.Where(&ISP{Name: isp.Name}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(isp).Error; err != nil {
			return nil, err
		}
		return isp, nil
	}

	return m, nil
}
