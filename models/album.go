package models

import (
	"github.com/jinzhu/gorm"
)

//Album テーブル準備
type Album struct {
	gorm.Model
	Name string `json:"name" gorm:"unique;not null"`
}
