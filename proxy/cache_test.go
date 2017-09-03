package proxy

import (
	"net/http"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	c := NewCache()
	c.Add("test", time.Second, []byte("something"), http.Header{})

	if c.Size != len("something") {
		t.Errorf("Size is incorrect %v != %v", c.Size, len("something"))
	}

	data, ok := c.Data["test"]
	if !ok {
		t.Error("Entry for test not found")
	}

	if time.Since(data.Time) > time.Second {
		t.Errorf("Time for cached data too far away %v != %v",
			data.Time,
			time.Now(),
		)
	}

	if data.Duration != time.Second {
		t.Errorf("Duration is incorrect %v != %v", data.Duration, time.Second)
	}

	value := string(data.Body)

	if value != "something" {
		t.Errorf("Value expected to be %s, got %s", "something", value)
	}
}

func TestDuration(t *testing.T) {
	c := NewCache()
	c.MaxSize = 4
	c.Add("test", time.Millisecond, []byte{0, 1, 2, 3}, http.Header{})

	// Wait until the data is invalid
	time.Sleep(time.Millisecond * 2)

	value, _ := c.Get("test")

	if value != nil {
		t.Errorf("An invalid value should return nil but it returns %v", value)
	}

	if c.Size != 0 {
		t.Errorf("After invalidation the size should be 0 but it is %v", c.Size)
	}

	c.Add("test1", time.Millisecond, []byte{0, 1, 2, 3}, http.Header{})
	c.Add("test2", time.Millisecond, []byte{0, 1, 2, 3}, http.Header{})
	c.Add("test3", time.Millisecond, []byte{0, 1, 2, 3}, http.Header{})
	c.Add("test4", time.Millisecond, []byte{0, 1, 2, 3}, http.Header{})

	time.Sleep(time.Millisecond * 2)
	c.Add("test5", time.Millisecond, []byte{0, 1, 2, 3}, http.Header{})

	if c.Size != 4 {
		t.Errorf("The cache should have been cleaned but is not %v != 4", c.Size)
	}
}

func TestClean(t *testing.T) {
	c := NewCache()
	c.MaxSize = 4 * 4

	c.Add("test1", time.Second, []byte{0, 1, 2, 3}, http.Header{})
	c.Add("test2", time.Second, []byte{0, 1, 2, 3}, http.Header{})
	c.Add("test3", time.Second, []byte{0, 1, 2, 3}, http.Header{})
	c.Add("test4", time.Millisecond, []byte{0, 1, 2, 3}, http.Header{})

	time.Sleep(time.Millisecond * 2)

	// This should trigger cleanup
	c.Add("test5", time.Second, []byte{0, 1, 2, 3}, http.Header{})

	if _, ok := c.Data["test4"]; ok {
		t.Error("The entry test4 should not exist")
	}

	// Refresh access time for test2
	_, _ = c.Get("test2")

	c.Add("test6", time.Second, []byte{0, 1, 2, 3}, http.Header{})

	if _, ok := c.Data["test2"]; !ok {
		t.Error("The entry test2 should exist as it was refreshed")
	}

	if _, ok := c.Data["test1"]; ok {
		t.Error("The entry test1 should not exist")
	}
}
