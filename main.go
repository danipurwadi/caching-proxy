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
	body       []byte
	headers    http.Header
}

func main() {
	var port int
	var origin string

	flag.IntVar(&port, "port", 3000, "port on which the caching proxy server will run")
	flag.StringVar(&origin, "origin", "", "URL of the server to which the requests will be forwarded")
	flag.Parse()

	if origin == "" {
		log.Fatal("failed to start program due to missing parameter: --origin <string>")
	}

	initServer(port, origin)
}

func initServer(portNo int, origin string) {
	cache := make(map[string]cacheHttpResp)

	log.Printf("starting cache-proxy at %d for origin %s...\n", portNo, origin)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if cache[path].body == nil {
			resp, err := http.Get(origin + path)
			if err != nil {
				log.Println("request to origin failed", err)
				http.Error(w, "failed to get response", http.StatusInternalServerError)
			}

			defer resp.Body.Close()
			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("failed to parse response body", err)
				http.Error(w, "failed to get response from source", http.StatusInternalServerError)
			}

			cache[path] = cacheHttpResp{
				statusCode: resp.StatusCode,
				body:       respBytes,
				headers:    resp.Header,
			}
			resp.Header["X-Cache"] = []string{"MISS"}
			writeResp(w, resp.StatusCode, resp.Header, respBytes)
		} else {
			resp := cache[path]
			resp.headers["X-Cache"] = []string{"HIT"}
			writeResp(w, resp.statusCode, resp.headers, resp.body)
		}
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNo), nil))
}

func writeResp(w http.ResponseWriter, statusCode int, headers map[string][]string, body []byte) {
	for k, vs := range headers {
		for _, v := range vs {
			w.Header().Set(k, v)
		}
	}
	w.WriteHeader(statusCode)
	w.Write(body)
}
