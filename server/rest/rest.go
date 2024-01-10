package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

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
			if segments[0] == "sensor" {
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
