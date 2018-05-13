package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/as/rate"
	"github.com/as/rate/httprate"
)

func main() {
	lim := rate.New(time.Minute * 5)
	http.Handle("/light", httprate.HandlerFunc(lim, time.Second, nil, func(tx http.ResponseWriter, rx *http.Request) {
		tx.Write([]byte("small page loaded"))
	}))

	heavyconf := httprate.Config{
		Error: http.HandlerFunc(func(tx http.ResponseWriter, rx *http.Request) {
			tx.Write([]byte("An HTTP handler can tell you things you don't want to tell yourself."))
		}),
	}
	http.Handle("/heavy", httprate.HandlerFunc(lim, time.Minute, &heavyconf, func(tx http.ResponseWriter, rx *http.Request) {
		tx.Write([]byte("heavy page loaded"))
	}))

	http.Handle("/", httprate.HandlerFunc(lim, time.Second, &heavyconf, func(tx http.ResponseWriter, rx *http.Request) {
		tx.Write([]byte("welcome to /"))
	}))
	fmt.Println(http.ListenAndServe(":80", nil))
}
