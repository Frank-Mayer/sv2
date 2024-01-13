package save

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type SensorData struct {
	Value float32 `json:"value"`
	Time  int64   `json:"time"`
}

type Sensor struct {
	Name string        `json:"name"`
	Unit string        `json:"unit"`
	Data *[]SensorData `json:"data"`
}

const (
	maxDataPoints = 1000
)

var (
	store = make([]Sensor, 0)
)

func Add(key string, value float32, unit string) {
	var utcNow = time.Now().Unix()

	for _, sensor := range store {
		if sensor.Name == key {
			*sensor.Data = append(*sensor.Data, SensorData{value, utcNow})
			sensor.Unit = unit
			DaBa.Create(&SensorDB{
				Model: gorm.Model{},
				Name:  key,
				Unit:  unit,
				Value: value,
				Time:  utcNow,
			})
			if len(*sensor.Data) > maxDataPoints {
				*sensor.Data = (*sensor.Data)[1:]
			}
			return
		}
	}
	// If we get here, we didn't find the sensor in the store.
	// Create a new one.
	store = append(store, Sensor{
		Name: key,
		Unit: unit,
		Data: &[]SensorData{
			{value, utcNow},
		},
	})
}

func AddFromDB(key string, value float32, unit string, utcNow int64) {
	utcNow = time.Now().Unix()

	for _, sensor := range store {
		if sensor.Name == key {
			*sensor.Data = append(*sensor.Data, SensorData{value, utcNow})
			sensor.Unit = unit
			return
		}
	}
	// If we get here, we didn't find the sensor in the store.
	// Create a new one.
	store = append(store, Sensor{
		Name: key,
		Unit: unit,
		Data: &[]SensorData{
			{value, utcNow},
		},
	})
}

func Get() []Sensor {
	return store
}

func GetNamed(name string) (*Sensor, error) {
	for _, sensor := range store {
		if sensor.Name == name {
			return &sensor, nil
		}
	}
	return nil, fmt.Errorf("sensor %s not found", name)
}
