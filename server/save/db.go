package save

import (
	"github.com/charmbracelet/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// const DBPATH = "/binds/test.db"

func Init() {
	var err error

	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", "error", err)
	}

	err = db.AutoMigrate(&Sensor{}, &SensorData{})
	if err != nil {
		log.Fatal("failed to auto migrate tables", "error", err)
	}

	err = db.Create(&Sensor{Name: "Peter", Unit: "Peta"}).Error
	if err != nil {
		log.Fatal("failed to create initial record", "error", err)
	}

	log.Info("Database initialization successful")
}

func Close() {
	dbInstance, _ := db.DB()
	if err := dbInstance.Close(); err != nil {
		log.Fatal("Failed to close the Database", "error", err)
	}
}
