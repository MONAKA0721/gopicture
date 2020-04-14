package models

import (
	"database/sql"
	"gopicture/database"

	"github.com/jinzhu/gorm"
)

var db *gorm.DB

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

func FindAlbums(uid uint) (*sql.Rows, error) {
	db := database.GetDB()
	rows, err := db.Raw(`SELECT albums.name, albums.hash, albums.id
		FROM albums INNER JOIN user_albums ON albums.id = user_albums.album_id
		WHERE user_albums.user_id = ?`, uid).Rows()
	return rows, err
}

func FindTopPicture(aid int) *sql.Row {
	db := database.GetDB()
	row := db.Raw(`SELECT temp.pname FROM
		(SELECT p.name pname, count(*) cnt
		FROM (albums a INNER JOIN pictures p on a.id = p.album_id)
		INNER JOIN user_fav_pictures f
		ON p.id = f.picture_id where a.id = ? GROUP BY p.name) temp
		WHERE temp.cnt = (SELECT max(cnt2)
		FROM(SELECT p.name pname, count(*) cnt2
		FROM (albums a INNER JOIN pictures p on a.id = p.album_id)
		INNER JOIN user_fav_pictures f ON p.id = f.picture_id where a.id = ?
		GROUP BY p.name) num)`, aid, aid).Row()
	return row
}
