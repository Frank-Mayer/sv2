package mqtt

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	paho "github.com/eclipse/paho.mqtt.golang"
)

const (
	QOS      = 1
	CLIENTID = "wetterstation"
)

func serverAddress() string {
	addr, ok := os.LookupEnv("SERVERADDRESS")
	if !ok {
		return "tcp://mosquitto:1883"
	}
	return addr
}

var (
	subs   = make(map[string]func(paho.Client, paho.Message))
	client paho.Client
)

func Sub(topic string, callback func([]byte)) {
	subs[topic] = func(_ paho.Client, msg paho.Message) {
		callback(msg.Payload())
	}
	log.Debug("added subscription", "topic", topic)
}

func Pub(topic string, payload []byte) error {
	token := client.Publish(topic, QOS, false, payload)
	<-token.Done()
	if err := token.Error(); err == nil {
		return nil
	} else {
		return errors.Join(
			fmt.Errorf("failed to publish message to topic %s", topic),
			token.Error(),
		)
	}
}

func Start() {
	client = paho.NewClient(options())

	// subscribe
	for topic, handler := range subs {
		client.AddRoute(topic, handler)
	}

	// connect
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("failed to connect", "error", token.Error())
	}
	log.Info("Connection is up")

	// wait for signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	log.Info("signal caught - exiting")
	client.Disconnect(1000)
	log.Info("shutdown complete")
}

func options() *paho.ClientOptions {
	opts := paho.NewClientOptions()
	opts.AddBroker(serverAddress())
	opts.SetClientID(CLIENTID)

	opts.SetOrderMatters(false)
	opts.ConnectTimeout = time.Second
	opts.WriteTimeout = time.Second
	opts.KeepAlive = 10
	opts.PingTimeout = time.Second

	opts.ConnectRetry = true
	opts.AutoReconnect = true

	opts.OnConnectionLost = func(_ paho.Client, err error) {
		log.Error("connection lost", "error", err)
	}

	opts.DefaultPublishHandler = func(_ paho.Client, msg paho.Message) {
		log.Warn(
			"received message on unsubscribed topic",
			"topic", msg.Topic(),
			"payload", string(msg.Payload()),
		)
	}

	opts.OnConnectionLost = func(_ paho.Client, err error) {
		log.Error("connection lost", "error", err)
	}

	opts.OnConnect = func(c paho.Client) {
		log.Info("connection established")

		for topic, handler := range subs {
			topic := topic
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
	}

	opts.OnReconnecting = func(paho.Client, *paho.ClientOptions) {
		log.Info("attempting to reconnect")
	}

	return opts
}
