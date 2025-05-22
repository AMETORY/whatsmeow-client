package service

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (db *gorm.DB, err error) {
	password := os.Getenv("PASSWORD")
	if password == "" {
		password = "balakutak"
	}
	user := os.Getenv("USER")
	if user == "" {
		user = "postgres"
	}
	/*************  ✨ Codeium Command ⭐  *************/
	dsn := "user=" + user + " dbname=whatsapp sslmode=disable password=" + password
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})

	/******  5bdd7335-cc1a-4535-b30d-faf16340e6f6  *******/
	// return gorm.Open(sqlite.Open("wa.db"), &gorm.Config{})
}
