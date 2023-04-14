package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestUpperCaseHandler(t *testing.T) {
	lock_map, subnet_count_map := get_blank_maps()

	srv := httptest.NewServer(http.HandlerFunc(rate_limiter_handler(lock_map, subnet_count_map)))
	defer srv.Close()

	request := func(wg *sync.WaitGroup, ip string, frequency time.Duration) {
		client := http.Client{}

		req, err := http.NewRequest("GET", srv.URL, nil)
		if err != nil {
			t.Error(err)
		}
		req.Header = http.Header{
			"X-Forwarded-For": {ip},
		}

		for {
			res, err := client.Do(req)
			if err != nil {
				t.Error(err)
			}
			log.Println(res.StatusCode)
			time.Sleep(frequency)
		}
		wg.Done()
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go request(&wg, "111.0.0.1", 2*time.Second)
	go request(&wg, "222.0.0.1", 5*time.Second)
	wg.Wait()
}
