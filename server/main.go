package main

import (
	"encoding/json"
	"sync"

	"github.com/Frank-Mayer/sv2-types/go"
	"github.com/Frank-Mayer/sv2/mqtt"
	"github.com/Frank-Mayer/sv2/rest"
	"github.com/Frank-Mayer/sv2/save"
	"github.com/charmbracelet/log"
	"google.golang.org/protobuf/proto"
)

func main() {
	log.SetLevel(log.DebugLevel)

	if err := save.Init(); err != nil {
		log.Fatal("failed to initialize database", "error", err)
	}
	defer func() {
		if err := save.Close(); err != nil {
			log.Error("failed to close database", "error", err)
		}
	}()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		mqtt.Sub("sensordata", func(data []byte) {
			msg := sv2_types.SensorData{}
			// try protobuf
			if err := proto.Unmarshal(data, &msg); err != nil {
				// try json
				if err := json.Unmarshal(data, &msg); err != nil {
					log.Error("failed to unmarshal message", "error", err, "data", data)
				}
			}
			log.Debug(
				"received message",
				"Name", msg.Name,
				"Value", msg.Value,
				"Unit", msg.Unit,
			)
			save.Add(msg.Name, msg.Value, msg.Unit)
		})
		mqtt.Start()
		log.Info("MQTT Subscriber stopped")
	}()

	go func() {
		defer wg.Done()
		rest.Rest()
		log.Info("Rest server stopped")
	}()

	wg.Wait()

	log.Info("Server exiting")
}
