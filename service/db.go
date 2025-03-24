package service

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (db *gorm.DB, err error) {
	/*************  ✨ Codeium Command ⭐  *************/
	dsn := "user=postgres dbname=whatsapp sslmode=disable password=balakutak"
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})

	/******  5bdd7335-cc1a-4535-b30d-faf16340e6f6  *******/
	// return gorm.Open(sqlite.Open("wa.db"), &gorm.Config{})
}
