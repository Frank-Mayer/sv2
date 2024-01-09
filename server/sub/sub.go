package sub

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	QOS      = 1
	CLIENTID = "mqtt_subscriber"
)

func serverAddress() string {
	addr, ok := os.LookupEnv("SERVERADDRESS")
	if !ok {
		return "tcp://mosquitto:1883"
	}
	return addr
}

func Sub(topic string, callback func(data []byte)) {
	handler := func(_ mqtt.Client, msg mqtt.Message) {
		callback(msg.Payload())
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(serverAddress())
	opts.SetClientID(CLIENTID)

	opts.SetOrderMatters(false)
	opts.ConnectTimeout = time.Second
	opts.WriteTimeout = time.Second
	opts.KeepAlive = 10
	opts.PingTimeout = time.Second

	opts.ConnectRetry = true
	opts.AutoReconnect = true

	opts.DefaultPublishHandler = func(_ mqtt.Client, msg mqtt.Message) {
		log.Warn("received message on unsubscribed topic", "topic", msg.Topic(), "payload", string(msg.Payload()))
	}

	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Error("connection lost", "error", err)
	}

	opts.OnConnect = func(c mqtt.Client) {
		log.Info("connection established")

		t := c.Subscribe(topic, QOS, handler)
		go func() {
			<-t.Done()
			if t.Error() != nil {
				log.Error("failed to subscribe", "error", t.Error())
			} else {
				log.Info("subscribtion successful", "topic", topic)
			}
		}()
	}
	opts.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		log.Info("attempting to reconnect")
	}

	client := mqtt.NewClient(opts)

	client.AddRoute(topic, handler)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("failed to connect", "error", token.Error())
	}
	log.Info("Connection is up")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	log.Info("signal caught - exiting")
	client.Disconnect(1000)
	log.Info("shutdown complete")
}
