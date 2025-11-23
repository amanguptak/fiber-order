package database

import (
	"log"
	"os"

	"github.com/amanguptak/fiber-api/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbInstance struct {
	Db *gorm.DB
}

var Database DbInstance

func ConnectDb() {
	db, err := gorm.Open(sqlite.Open("api.db"), &gorm.Config{})

	if err != nil {
		log.Fatal("failed to connect with database ! \n", err.Error())
		os.Exit(2)
	}
	log.Println("Connected to database")
	db.Logger = logger.Default.LogMode(logger.Info)
	log.Println("Running Migration")
	//Add Migration

	err = db.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}, &models.RefreshToken{})
	if err != nil {
		log.Fatal("Migration Failed: " + err.Error())
		os.Exit(2)
	}

	Database = DbInstance{Db: db}
}
