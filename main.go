package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/hftloai/hftlapiconnector"
)

var serv *server

var max = flag.Int("hist", 10, "number of historical events to generate")
var p = flag.Duration("period", 1*time.Minute, "how frequently to generate new alerts")
var port = flag.Int("port", 8000, "http port")

func main() {
	flag.Parse()
	serv = newServer()

	serv.apiConn = new(hftlapiconnector.APIClient)
	serv.apiConn.APIKey = "debug_token"
	serv.apiConn.BaseURL = "http://127.0.0.1:8081/"
	serv.apiConn.UserID = "kim"
	serv.apiConn.InitREST()

	// populate 10 events up front
	serv.seed(*max)

	// emit period events starting now
	go serv.generate(*p)
	// go serv.latencyLoop()

	// initialize routes, and start http server
	http.HandleFunc("/", cors(serv.root))
	http.HandleFunc("/annotations", cors(serv.annotations))
	http.HandleFunc("/query", cors(serv.queries))
	http.HandleFunc("/search", cors(serv.searches))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal(err)
	}
}
