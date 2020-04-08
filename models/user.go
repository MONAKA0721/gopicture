package models

import (
	"github.com/jinzhu/gorm"
)

//User テーブル準備
type User struct {
	gorm.Model
	Name  string `json:"name" gorm:"unique;not null"`
	Email string `json:"email" gorm:"unique;not null"`
}
