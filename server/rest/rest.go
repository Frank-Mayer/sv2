package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Frank-Mayer/sv2/morse"
	"github.com/Frank-Mayer/sv2/mqtt"
	"github.com/Frank-Mayer/sv2/save"
	"github.com/Frank-Mayer/sv2/webui"
	"github.com/charmbracelet/log"
)

// wrap takes a value and returns an object that contains the value as its only element.
// Example: wrap(1) -> {"value": 1}
// Example: wrap([]int{1,2,3}) -> {"value": [1,2,3]}
func wrap(key string, val interface{}) interface{} {
	return map[string]interface{}{key: val}
}

func writeJson(res http.ResponseWriter, val interface{}) {
	res.Header().Set("Content-Type", "application/json")
	jsonWriter := json.NewEncoder(res)
	if err := jsonWriter.Encode(val); err != nil {
		log.Error("error encoding json", "error", err)
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(fmt.Sprintf("500 error encoding json: %s", err)))
	}
}

// try to get addr from env var
func addr() string {
	port, ok := os.LookupEnv("REST_PORT")
	if !ok {
		log.Warn("REST_PORT not set, using default port 8080")
		return ":8080"
	}
	return ":" + port
}

// restful json api server
func Rest() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		segments := strings.Split(req.URL.Path, "/")[1:]
		// remove empty segments
		for i := 0; i < len(segments); i++ {
			if segments[i] == "" {
				segments = append(segments[:i], segments[i+1:]...)
				i--
			}
		}
		switch len(segments) {
		case 0:
			webui.Dashboard(res, req)
			return
		case 1:
			if segments[0] == "sensor" {
				writeJson(res, wrap("sensors", save.Get()))
				return
			}
		case 2:
			switch segments[0] {
			case "sensor":
				key := segments[1]
				data, err := save.GetNamed(key)
				if err != nil {
					res.WriteHeader(http.StatusNotFound)
					if _, err := res.Write([]byte(fmt.Sprintf("404 sensor %s not found", key))); err != nil {
						log.Error(
							"error writing 404 response",
							"error", err,
							"url", req.URL.Path,
						)
					}
					return
				}
				writeJson(res, data)
				return
			case "actor":
				if segments[1] == "led" {
					switch req.Method {
					case http.MethodGet:
						res.WriteHeader(http.StatusMethodNotAllowed)
						return
					case http.MethodPost:
						on := req.URL.Query().Get("on")
						if on == "on" || on == "off" {
							log.Debug("sending led command", "command", on)
							mqtt.Pub("led", []byte("{\"command\":\""+on+"\"}"))
							return
						}
						msg := req.URL.Query().Get("morse")
						if len(msg) > 16 {
							res.WriteHeader(http.StatusBadRequest)
							_, _ = res.Write([]byte("400 message too long (max 64 chars)"))
							return
						}
						if msg != "" {
							go func() {
								if err := Morse(msg); err != nil {
									log.Error("error sending morse", "error", err)
								}
							}()
							return
						}
						res.WriteHeader(http.StatusBadRequest)
					}
				}
			}
		}

		res.WriteHeader(http.StatusNotFound)
		if _, err := res.Write([]byte("404 page not found")); err != nil {
			log.Error("error writing 404 response", "error", err)
		}
	})

	log.Info("Rest server started", "address", addr())

	if err := http.ListenAndServe(addr(), nil); err != nil {
		log.Fatal(err)
	}
}

const (
	signPause     = 200 * time.Millisecond
	longSignPause = 400 * time.Millisecond
	shortPause    = 400 * time.Millisecond
	longPause     = 600 * time.Millisecond
)

func Morse(msg string) error {
	morseMsg := morse.Marshal(msg)
	log.Debug("sending led command", "command", morseMsg)
	for _, char := range morseMsg {
		switch char {
		case morse.Short:
			if err := mqtt.Pub("led", []byte("{\"command\":\"on\"}")); err != nil {
				return err
			}
			time.Sleep(signPause)
			if err := mqtt.Pub("led", []byte("{\"command\":\"off\"}")); err != nil {
				return err
			}
			time.Sleep(signPause)
		case morse.Long:
			if err := mqtt.Pub("led", []byte("{\"command\":\"on\"}")); err != nil {
				return err
			}
			time.Sleep(longSignPause)
			if err := mqtt.Pub("led", []byte("{\"command\":\"off\"}")); err != nil {
				return err
			}
			time.Sleep(signPause)
		case morse.ShortPause:
			time.Sleep(shortPause)
		case morse.LongPause:
			time.Sleep(longPause)
		}
	}
	return nil
}
