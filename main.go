package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/jfontan/go-proxy/proxy"
)

func main() {
	addr := flag.String("addr", "", "target URL")
	port := flag.String("port", "8080", "bind to port")
	cache_size := flag.Int("size", 16, "Cache size in Mb")

	flag.Parse()

	if len(*addr) == 0 {
		log.Fatal("Address must be provided")
	}

	proxy := proxy.NewProxy(*addr, *cache_size)

	s := &http.Server{
		Addr:    ":" + *port,
		Handler: proxy,
	}

	s.ListenAndServe()
}
