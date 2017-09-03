package proxy

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var urlCache *Cache

type Proxy struct {
	Cache *Cache
	URL   string
}

func NewProxy(url string, cache_size int) *Proxy {
	proxy := new(Proxy)
	proxy.Cache = NewCache()
	proxy.Cache.MaxSize = 1024 * 1024 * cache_size
	proxy.URL = url

	return proxy
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body_data, header := p.Cache.Get(r.URL.Path)

	if len(body_data) == 0 {
		log.Print("NORMAL", r.URL.Path)
		resp, _ := http.Get(p.URL + r.URL.Path)
		defer resp.Body.Close()

		body_data, _ = ioutil.ReadAll(resp.Body)
		header = resp.Header

		err := p.Cache.Add(r.URL.Path, time.Second*600, body_data, header)
		if err != nil {
			log.Print(err)
		}
	} else {
		log.Print("CACHED", r.URL.Path)
	}

	for k, v := range header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.Write(body_data)
}
