package main

import (
	"sync"

	"github.com/Frank-Mayer/sv2-types/go"
	"github.com/Frank-Mayer/sv2/rest"
	"github.com/Frank-Mayer/sv2/save"
	"github.com/Frank-Mayer/sv2/sub"
	"github.com/charmbracelet/log"
	"google.golang.org/protobuf/proto"
)

func main() {
	log.SetLevel(log.DebugLevel)

	save.Init()
	defer save.Close()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		sub.Sub("sensordata", func(data []byte) {
			msg := sv2_types.SensorData{}
			if err := proto.Unmarshal(data, &msg); err != nil {
				log.Error("failed to unmarshal message", "error", err, "data", data)
			}
			log.Debug(
				"received message",
				"Name", msg.Name,
				"Value", msg.Value,
				"Unit", msg.Unit,
			)
			save.Add(msg.Name, msg.Value, msg.Unit)
		})
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
