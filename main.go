package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/hftloai/hftlapiconnector"
)

var max = flag.Int("hist", 10, "number of historical events to generate")
var p = flag.Duration("period", 1*time.Minute, "how frequently to generate new alerts")
var port = flag.Int("port", 8000, "http port")

func main() {
	flag.Parse()
	s := newServer()

	s.apiConn = new(hftlapiconnector.APIClient)
	s.apiConn.APIKey = "debug_token"
	s.apiConn.BaseURL = "http://127.0.0.1:8080/"
	s.apiConn.UserID = "kim"
	s.apiConn.InitREST()

	// populate 10 events up front
	s.seed(*max)

	// emit period events starting now
	go s.generate(*p)

	// initialize routes, and start http server
	http.HandleFunc("/", cors(s.root))
	http.HandleFunc("/annotations", cors(s.annotations))
	http.HandleFunc("/query", cors(s.queries))
	http.HandleFunc("/search", cors(s.searches))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal(err)
	}
}
