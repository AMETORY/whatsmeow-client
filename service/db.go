package service

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() (db *gorm.DB, err error) {
	return gorm.Open(sqlite.Open("wa.db"), &gorm.Config{})
}
