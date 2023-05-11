package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))
}

func GenError(w http.ResponseWriter, code int, message string) {
	var errorM = ErrorFormat{
		Code:    code,
		Message: fmt.Sprintf("error: %s", message),
	}

	w.WriteHeader(500)

	json.NewEncoder(w).Encode(errorM)
}
