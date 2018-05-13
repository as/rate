# Rate 
[![Go Report Card](https://goreportcard.com/badge/github.com/as/rate)](https://goreportcard.com/badge/github.com/as/rate)

# Usage

```
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
	http.Handle("/", httprate.HandlerFunc(lim, time.Minute, nil, func(tx http.ResponseWriter, rx *http.Request) {
		tx.Write([]byte("welcome to /"))
	}))
	fmt.Println(http.ListenAndServe(":80", nil))
}
```
