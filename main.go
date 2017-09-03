package main

import (
	"net/http"

	"github.com/jfontan/go-proxy/proxy"
)

func main() {
	proxy := proxy.NewProxy("http://localhost:6060")

	s := &http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}

	s.ListenAndServe()
}
