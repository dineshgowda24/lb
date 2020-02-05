package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/dineshgowda24/loadbalancer/backendserver/config"
)

func main() {
	c := flag.String("c", "config/config.json", "Please specify conf.json")
	flag.Parse()
	file, err := os.Open(*c)
	if err != nil {
		log.Fatal("Unable to open config file")
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := bconfig.BackendConfiguration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Unable to decode conf.json file")
	}

	for _, s := range config.Servers {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			response := "I came from " + s.Host + s.Port
			w.Write([]byte(response))
			w.WriteHeader(http.StatusOK)
		})
		if err := http.ListenAndServe(s.Host+":"+s.Port, mux); err != nil {
			log.Printf("Unable to start server at : %s : %s ", s.Host, s.Port)
		} else {
			log.Printf("Started server at : %s : %s ", s.Host, s.Port)
		}
	}
}
