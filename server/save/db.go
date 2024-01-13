package save

import (
	"errors"

	"github.com/charmbracelet/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

const dbpath = "/binds/sensors.db"

func Init() error {
	var err error

	if db, err = gorm.Open(sqlite.Open(dbpath), &gorm.Config{}); err != nil {
		return errors.Join(
			errors.New("failed to connect database"),
			err,
		)
	}

	if err := db.AutoMigrate(&Sensor{}, &SensorData{}); err != nil {
		return errors.Join(
			errors.New("failed to migrate database"),
			err,
		)
	}

	if tx := db.Create(&Sensor{Name: "Peter", Unit: "Peta"}); tx.Error != nil {
		return errors.Join(
			errors.New("failed to create initial record"),
			tx.Error,
		)
	}

	return nil
}

func Close() error {
	dbInstance, _ := db.DB()
	if err := dbInstance.Close(); err != nil {
		return errors.Join(
			errors.New("failed to close database"),
			err,
		)
	}

	return nil
}
