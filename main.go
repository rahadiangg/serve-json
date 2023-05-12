package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type ErrorFormat struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Config struct {
	Endpoint    string `yaml:"endpoint"`
	DefaultCode int    `yaml:"defaultCode"`
	File        string `yaml:"file"`
}

type Configs struct {
	Configs []Config `yaml:"configs"`
}

func main() {

	envVars := map[string]string{
		"CONFIG": "config.yaml",
		"PORT":   "3000",
	}

	for env, defaultValue := range envVars {
		_, exist := os.LookupEnv(env)
		if !(exist) {
			os.Setenv(env, defaultValue)
		}
	}

	// load yaml
	f, err := os.ReadFile(os.Getenv("CONFIG"))
	if err != nil {
		log.Fatal(err)
	}

	// initialize variable
	var c Configs

	// parse yaml config
	if err := yaml.Unmarshal(f, &c); err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	logger := logrus.New()

	loggingMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			logger.WithField("nil", logrus.Fields{
				"method":   r.Method,
				"URI":      r.RequestURI,
				"duration": time.Since(start),
			}).Info("Recived request")

			h.ServeHTTP(w, r)
		})
	}

	// fmt.Printf("%+v\n", c)
	for _, x := range c.Configs {

		handleFunc := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			// read json file
			file, err := os.ReadFile(x.File)
			if err != nil {
				GenError(w, 500, err.Error())
				return
			}

			// parse json
			var parsed interface{}
			err = json.Unmarshal(file, &parsed)
			if err != nil {
				GenError(w, 500, err.Error())
				return
			}

			// return json
			w.WriteHeader(x.DefaultCode)
			err = json.NewEncoder(w).Encode(parsed)
			if err != nil {
				GenError(w, 500, err.Error())
				return
			}
		}

		router.HandleFunc(x.Endpoint, handleFunc).Methods("GET")
	}

	wrapperMiddleware := loggingMiddleware(router)
	address := ":" + os.Getenv("PORT")
	logger.WithField("addr", address).Info("Starting server")
	if err := http.ListenAndServe(address, wrapperMiddleware); err != nil {
		logger.WithField("event", "start server").Fatal(err)
	}
}

func GenError(w http.ResponseWriter, code int, message string) {
	var errorM = ErrorFormat{
		Code:    code,
		Message: fmt.Sprintf("error: %s", message),
	}

	w.WriteHeader(500)

	json.NewEncoder(w).Encode(errorM)
}
