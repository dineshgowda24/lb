package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dineshgowda24/lb/backendserver/config"
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
	http.HandleFunc("/", handle)
	for _, s := range config.Servers {
		log.Printf("Started server at : %s : %s ", s.Host, s.Port)
		go func(host, port string) {
			if err := http.ListenAndServe(host+":"+port, nil); err != nil {
				log.Printf("Unable to start server at : %s : %s ", host, port)
			}
		}(s.Host, s.Port)
	}
	select {}
}

func handle(w http.ResponseWriter, r *http.Request) {
	s := r.Context().Value(http.ServerContextKey).(*http.Server)
	fmt.Fprint(w, "I came from "+s.Addr)
}
