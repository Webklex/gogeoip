package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Network struct {
	ID uint `gorm:"primarykey" json:"-"`

	Network string `gorm:"index" json:"network"`
	Domain  string `json:"domain"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Ips []IP `json:"-"`
}

func UpdateNetwork(db *gorm.DB, nw *Network) (*Network, error) {
	m := &Network{}
	if err := db.Where(&Network{Network: nw.Network, Domain: nw.Domain}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(nw).Error; err != nil {
			return nil, err
		}
		return nw, nil
	}

	return m, nil
}
