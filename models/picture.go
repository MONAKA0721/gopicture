package models

import (
	"github.com/jinzhu/gorm"
)

//Picture テーブル準備
type Picture struct {
	gorm.Model
	Name    string `gorm:"not null"`
	AlbumID int    `gorm:"index"`
}
