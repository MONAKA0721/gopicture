package models

import (
	"gopicture/database"

	"github.com/jinzhu/gorm"
)

//User テーブル準備
type User struct {
	gorm.Model
	Name        string    `json:"name" gorm:"not null"`
	Email       string    `json:"email" gorm:"unique;not null"`
	Password    string    `json:"password"`
	Albums      []Album   `gorm:"many2many:user_albums;"`
	FavPictures []Picture `gorm:"many2many:user_fav_pictures;"`
}

// FindByEmail finds a user by email
func (u *User) FindByEmail(email string) (err error) {
	db := database.GetDB()
	return db.Where("email = ?", email).First(u).Error
}

func (u *User) First() (err error) {
	db := database.GetDB()
	return db.First(u).Error
}

func (u *User) FirstOrCreate(email string, name string) (err error) {
	db := database.GetDB()
	return db.Where(User{Email: email}).Attrs(User{Name: name}).FirstOrCreate(u).Error
}

func (u *User) AppendUserAlbums(album Album) (err error) {
	db := database.GetDB()
	return db.Model(&u).Association("Albums").Append(album).Error
}

func (u *User) AppendFavPictures(pid int) {
	picture := Picture{}
	db := database.GetDB()
	db.First(&picture, "id = ?", pid)
	db.Model(&u).Association("FavPictures").Append(picture)
}

func (u *User) Create() (err error) {
	db := database.GetDB()
	return db.Create(u).Error
}
