package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

type cacheHttpResp struct {
	statusCode int
	respBody   []byte
	headers    http.Header
}

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
	cache := make(map[string]cacheHttpResp)

	log.Printf("starting cache-proxy at %d for origin %s...\n", portNo, origin)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if cache[path].respBody == nil {
			resp, err := http.Get(origin + path)
			if err != nil {
				log.Println("request to origin failed", err)
				http.Error(w, "failed to get response", http.StatusInternalServerError)
			}

			defer resp.Body.Close()

			for k, vs := range resp.Header {
				for _, v := range vs {
					w.Header().Set(k, v)
				}
			}
			w.Header().Set("X-Cache", "MISS")
			w.WriteHeader(resp.StatusCode)

			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("failed to parse response body", err)
				http.Error(w, "failed to get response from source", http.StatusInternalServerError)
			}

			cache[path] = cacheHttpResp{
				statusCode: resp.StatusCode,
				respBody:   respBytes,
				headers:    resp.Header,
			}
			w.Write(respBytes)
		} else {
			resp := cache[path]
			for k, vs := range resp.headers {
				for _, v := range vs {
					w.Header().Set(k, v)
				}
			}
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(resp.statusCode)

			w.Write(resp.respBody)
		}
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNo), nil))
}
