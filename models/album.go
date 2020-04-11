package models

import (
	"github.com/jinzhu/gorm"

	"gopicture/database"
)

//Album テーブル準備
type Album struct {
	gorm.Model
	Name     string `json:"name" gorm:"not null"`
	Hash     string `json:"hash" gorm:"unique;not null"`
	Pictures []Picture
}

func (a *Album) Create() (err error) {
	db := database.GetDB()
	return db.Create(a).Error
}
