package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type Continent struct {
	ID uint `gorm:"primarykey" json:"-"`

	Code string `gorm:"index" json:"code"`
	Name string `json:"name"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Countries []Country `json:"-"`
}

func UpdateContinent(db *gorm.DB, continent *Continent) (*Continent, error) {
	m := &Continent{}
	if err := db.Where(&Continent{Code: continent.Code}).First(m).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if m.ID <= 0 {
		if err := db.Create(continent).Error; err != nil {
			return nil, err
		}
		return continent, nil
	} else if m.Name != continent.Name && continent.Name != "" {
		m.Name = continent.Name
		if err := db.Save(m).Error; err != nil {
			return nil, err
		}
	}
	return m, nil
}
