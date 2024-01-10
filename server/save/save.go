package save

import (
	"fmt"
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
