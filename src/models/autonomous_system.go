package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type AutonomousSystem struct {
	ID uint `gorm:"primarykey" json:"-"`

	Name   string `json:"name"`
	Number uint   `gorm:"index" json:"number"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Ips []IP `json:"-"`
}

func UpdateAutonomousSystem(db *gorm.DB, as *AutonomousSystem) (*AutonomousSystem, error) {
	m := &AutonomousSystem{}
	if err := db.Where(&AutonomousSystem{Number: as.Number}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(as).Error; err != nil {
			return nil, err
		}
		return as, nil
	} else if m.Name != as.Name {
		m.Name = as.Name
		if err := db.Save(m).Error; err != nil {
			return nil, err
		}
	}

	return m, nil
}
