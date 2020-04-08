package database

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"gopicture/config"
	"gopicture/models"
)

var db *gorm.DB

// Init initializes database
func Init(isReset bool) {
	db, err := gorm.Open(config.GetDBConfig())
	if err != nil {
		fmt.Println(err)
	}
	db.LogMode(true)
	if isReset {
		db.DropTableIfExists()
	}
	db.AutoMigrate(models.User{})
}

// GetDB returns database connection
func GetDB() *gorm.DB {
	return db
}

// Close closes database
func Close() {
	db.Close()
}
