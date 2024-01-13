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
	gorm.Model
	Name string        `json:"name"`
	Unit string        `json:"unit"`
	Data *[]SensorData `json:"data" gorm:"-"`
}

const (
	maxDataPoints = 1000
)

var (
	store = make([]Sensor, 0)
)

func utcNow() int64 {
	return time.Now().Unix()
}

func Add(key string, value float32, unit string) {
	for _, sensor := range store {
		if sensor.Name == key {
			*sensor.Data = append(*sensor.Data, SensorData{value, utcNow()})
			sensor.Unit = unit
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
			{value, utcNow()},
		},
	})

	db.Create(&Sensor{Name: key, Unit: unit})
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
