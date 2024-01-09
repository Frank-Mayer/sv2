package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Frank-Mayer/sv2/save"
	"github.com/charmbracelet/log"
)

func isObject(val interface{}) bool {
	if val == nil {
		return true
	}
	switch val.(type) {
	case []interface{}:
		return false
	case string:
		return false
	case float32:
		return false
	default:
		return false
	}
}

// wrap takes a value and returns an object that contains the value as its only element.
// Example: wrap(1) -> {"value": 1}
// Example: wrap([]int{1,2,3}) -> {"value": [1,2,3]}
func wrap(val interface{}) interface{} {
	return map[string]interface{}{"value": val}
}

func jsonWrapper(handler func(req *http.Request) (any, error)) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		log.Debug("request", "url", req.URL, "method", req.Method)
		val, err := handler(req)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			log.Debug("failed to handle request", "error", err)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		var jsonStr []byte
		if isObject(val) {
			jsonStr, err = json.Marshal(val)
		} else {
			jsonStr, err = json.Marshal(wrap(val))
		}
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			log.Debug("failed to handle request", "error", err)
			return
		}
		log.Debug("response", "json", string(jsonStr))
		res.Write(jsonStr)
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
	http.HandleFunc("/", jsonWrapper(func(req *http.Request) (any, error) {
		segments := strings.Split(req.URL.Path, "/")[1:]
		// remove empty segments
		for i := 0; i < len(segments); i++ {
			if segments[i] == "" {
				segments = append(segments[:i], segments[i+1:]...)
				i--
			}
		}
		switch len(segments) {
		case 1:
			if segments[0] == "sensor" {
				return save.Keys(), nil
			}
		case 2:
			if segments[0] == "sensor" {
				key := segments[1]
				data, ok := save.Get(key)
				if !ok {
					return nil, fmt.Errorf("unknown sensor %s", key)
				}
				return data, nil
			}
		}

		return nil, fmt.Errorf("unknown path %s", req.URL.Path)
	}))

	log.Info("Rest server started", "address", addr())

	if err := http.ListenAndServe(addr(), nil); err != nil {
		log.Fatal(err)
	}
}
