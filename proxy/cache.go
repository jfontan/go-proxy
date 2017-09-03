package proxy

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

type Cache struct {
	m       sync.Mutex
	MaxSize int
	Size    int
	Data    map[string]Data
}

type Data struct {
	Time       time.Time
	LastAccess time.Time
	Duration   time.Duration
	Size       int
	Body       []byte
	Header     http.Header
}

func NewCache() *Cache {
	cache := new(Cache)
	cache.Data = make(map[string]Data)
	cache.MaxSize = 1024 * 1024 // 1 Mb

	return cache
}

func (c *Cache) Add(name string, duration time.Duration, bytes []byte, header http.Header) error {
	log.Print("ADD: ", name, len(bytes))
	size := len(bytes)
	size += header_size(header)

	if size > c.MaxSize {
		return errors.New("Data too big for the cache size")
	}

	if c.Size+size > c.MaxSize {
		c.clean(size)
	}

	data := Data{
		Time:       time.Now(),
		LastAccess: time.Now(),
		Duration:   duration,
		Size:       size,
		Body:       bytes,
		Header:     header,
	}

	c.m.Lock()
	c.Data[name] = data
	c.Size += data.Size
	c.m.Unlock()

	log.Printf("Cache size %v/%v, %v%%", c.Size, c.MaxSize,
		float32(c.Size)/float32(c.MaxSize)*100.0)

	return nil
}

func (c *Cache) Get(name string) ([]byte, http.Header) {
	log.Print("GET: ", name)
	data, ok := c.Data[name]
	if !ok {
		return nil, nil
	}

	if time.Since(data.Time) > data.Duration {
		c.remove(name)
		return nil, nil
	}

	data.LastAccess = time.Now()
	return data.Body, data.Header
}

func (c *Cache) remove(name string) {
	data, ok := c.Data[name]
	if ok {
		c.m.Lock()
		c.Size -= data.Size
		delete(c.Data, name)
		c.m.Unlock()

		log.Printf("DELETE: %q", name)
		log.Printf("Cache size %v/%v, %v%%", c.Size, c.MaxSize,
			float32(c.Size)/float32(c.MaxSize)*100.0)

	}
}

type lastAccess struct {
	Time time.Time
	Name string
}
type byLastAccess []lastAccess

func (a byLastAccess) Len() int           { return len(a) }
func (a byLastAccess) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLastAccess) Less(i, j int) bool { return a[i].Time.Before(a[j].Time) }

func (c *Cache) clean(size int) {
	times := c.clean_invalid()

	if c.Size+size < c.MaxSize {
		return
	}

	for _, t := range times {
		c.remove(t.Name)
		if c.Size+size < c.MaxSize {
			break
		}
	}
}

func (c *Cache) clean_invalid() byLastAccess {
	times := byLastAccess{}

	for name, v := range c.Data {
		if time.Since(v.Time) > v.Duration {
			c.remove(name)
		} else {
			times = append(times, lastAccess{v.LastAccess, name})
		}
	}

	return times
}

func header_size(header http.Header) int {
	size := 0

	for _, v := range header {
		for _, vv := range v {
			size += len(vv)
		}
	}

	return size
}
