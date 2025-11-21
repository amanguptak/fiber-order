package helpers

import (
	"github.com/amanguptak/fiber-api/database"
	"gorm.io/gorm"
)

func DB() *gorm.DB {
	return database.Database.Db
}