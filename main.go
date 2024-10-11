package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
)

func main() {
	portPtr := flag.Int("port", 3000, "port on which the caching proxy server will run")
	originPtr := flag.String("origin", "", "URL of the server to which the requests will be forwarded")
	flag.Parse()

	if *originPtr == "" {
		log.Fatal("failed to start program due to missing parameter: --origin <string>")
	}

	initServer(*portPtr, *originPtr)
}

func initServer(portNo int, origin string) {
	log.Printf("starting cache-proxy at %d for origin %s...\n", portNo, origin)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Forwarding request to %s%s", origin, html.EscapeString(r.URL.Path))
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNo), nil))
}