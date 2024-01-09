package main

import (
	"sync"

	sv2_types "github.com/Frank-Mayer/sv2-types/go"
	"github.com/Frank-Mayer/sv2/sub"
	"github.com/charmbracelet/log"
	"google.golang.org/protobuf/proto"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		sub.Sub("sensordata", func(data []byte) {
			msg := sv2_types.SensorData{}
			if err := proto.Unmarshal(data, &msg); err != nil {
				log.Error("failed to unmarshal message", "error", err, "data", data)
			}
		})
	}()

	wg.Wait()
}
