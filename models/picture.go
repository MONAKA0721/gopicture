package models

import (
	"gopicture/database"

	"github.com/jinzhu/gorm"
)

//Picture テーブル準備
type Picture struct {
	gorm.Model
	Name    string `gorm:"not null"`
	AlbumID int    `gorm:"index"`
}

func (p *Picture) FindFirstPicture(aid int) (err error) {
	db := database.GetDB()
	return db.Where("album_id = ?", aid).First(&p).Error
}
