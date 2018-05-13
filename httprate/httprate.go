// Pakcage httprate is a helper package for rate-limiting http endpoints
package httprate

import (
	"log"
	"net/http"
	"time"

	"github.com/as/rate"
)

// DefaultConfig is the default config parameters for a LimitedHandler
var DefaultConfig = Config{
	// The request host
	TaskFunc: host,
	Error:    http.HandlerFunc(LimitExceeded),
}

// LimitedHandler is an http.Handler that uses its TaskFunc to extract the identifying
// task name associated with an http.Request. It then schedules that named task with
// the configured Limiter and Cost, running the underlying handler if and only if that
// task can be executed at time.Now().
type LimitedHandler struct {
	// Cost is the unit of duration for running the underlying handler. The Cost does not
	// have to correspond to real-world execution time.
	Cost time.Duration

	// Limiter this handler will use to decide whether it can run its underlying handler
	rate.Limiter

	// Config has optional settings
	Config

	// Handler is the underlying handler that will be run if the request is allowed
	Handler http.Handler
}

// Config configures a LimitedHandler with supplementay options
type Config struct {
	// TaskFunc extracts a task name from an http request/response pair. The default is the request host.
	TaskFunc func(*http.Request) string

	// Error handler, if set, is called when a rate limit is hit instead of the default handler, which
	// returns a 429 status and writes "rate limit exceeded" to the http.ResponseWriter
	Error http.Handler
}

func (c *Config) ensure() *Config {
	if c == nil {
		d := DefaultConfig
		c = &d
	}
	if c.TaskFunc == nil {
		c.TaskFunc = host
	}
	if c.Error == nil {
		c.Error = http.HandlerFunc(LimitExceeded)
	}
	return c
}

// Handler returns an http.Handler that checks the incoming request against the limiter and cost and executes handler
// if the limiter allows it at time.Now(). A nil conf is silently replaced with the default configuration.
func Handler(lim rate.Limiter, cost time.Duration, conf *Config, handler http.Handler) *LimitedHandler {
	return &LimitedHandler{
		Config:  *conf.ensure(),
		Limiter: lim,
		Handler: handler,
		Cost:    cost,
	}
}

// HandlerFunc is similar to http.HandlerFunc, except it returns a rate-limited handler
func HandlerFunc(lim rate.Limiter, cost time.Duration, conf *Config, h func(http.ResponseWriter, *http.Request)) *LimitedHandler {
	return Handler(lim, cost, conf, http.HandlerFunc(h))
}

// ServeHTTP implements http.Handler
func (l *LimitedHandler) ServeHTTP(tx http.ResponseWriter, rx *http.Request) {
	if !rate.AllowSlice(l.Limiter, l.TaskFunc(rx), l.Cost) {
		l.Error.ServeHTTP(tx, rx)
		return
	}
	l.Handler.ServeHTTP(tx, rx)
}

// LimitExceeded is the default error handler. It writes the http.StatusTooManyRequests message along with
// the standard status test for that message.
func LimitExceeded(tx http.ResponseWriter, rx *http.Request) {
	tx.WriteHeader(http.StatusTooManyRequests)
	tx.Write([]byte(http.StatusText(http.StatusTooManyRequests)))
}

func host(rx *http.Request) string {
	log.Println(rx.Host)
	return rx.Host
}
