package save

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DaBa *gorm.DB

const dbpath = "/binds/sensors.db"

type SensorDB struct {
	gorm.Model
	Name  string
	Unit  string
	Value float32
	Time  int64
}

func Init() error {
	var err error

	if DaBa, err = gorm.Open(sqlite.Open(dbpath), &gorm.Config{
		SkipDefaultTransaction: true,
	}); err != nil {
		return errors.Join(
			errors.New("failed to connect database"),
			err,
		)
	}

	if err := DaBa.AutoMigrate(&SensorDB{}); err != nil {
		return errors.Join(
			errors.New("failed to migrate database"),
			err,
		)
	}

	err = ReadLatestFromDB()
	if err != nil {
		return err
	}

	return nil
}

func ReadLatestFromDB() error {
	var sensorsDB []SensorDB

	// Fetch the latest 50 entries from the database ordered by time descending
	if err := DaBa.Order("time desc").Limit(50).Find(&sensorsDB).Error; err != nil {
		return errors.Join(
			errors.New("failed to read latest entries from database"),
			err,
		)
	}

	// Create a temporary slice to store the entries in reverse order
	reversed := make([]SensorDB, len(sensorsDB))
	for i, j := 0, len(sensorsDB)-1; i < len(sensorsDB); i, j = i+1, j-1 {
		reversed[i] = sensorsDB[j]
	}

	// Iterate over the reversed slice and call AddFromDB to store in memory
	for _, dbEntry := range reversed {
		AddFromDB(dbEntry.Name, dbEntry.Value, dbEntry.Unit, dbEntry.Time)
	}

	return nil
}

func Close() error {
	dbInstance, _ := DaBa.DB()
	if err := dbInstance.Close(); err != nil {
		return errors.Join(
			errors.New("failed to close database"),
			err,
		)
	}

	return nil
}
