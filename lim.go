// Package rate provides a very simple rate limiter based on the passage of time.
package rate

import (
	"time"
)

const (
	tickInterval    = time.Second * 3
	preallocEntries = 64
	maxSweep        = 10
)

// Limiter provides a way to schedule named tasks for execution.
type Limiter interface {
	// Quantum returns the duration allocated for every named task. This value is a
	// limiter-scoped, maximum-duration watermark. It does not represent the quantum
	// available for a specific task.
	//
	// If the Limiter is a aggregate of multiple Limiters, Quantum() should return the smallest
	// time.Duration in the aggregate.
	Quantum() time.Duration

	// Schedule schedules the task to run for the given time slice if there is quantum
	// available for that task.
	//
	// If the delay is <= 0 the task can run immediately and the time slice provided
	// is subtracted from the task's quantum. If delay is > 0, the caller may wait the delay
	// and attempt to schedule the task again, otherwise the task should be abandoned.
	Schedule(task string, slice time.Duration) (delay time.Duration)

	// Close closes the limiter
	Close() error
}

// AllowSlice returns true if task may execute for 1s at time.Now()
func Allow(l Limiter, task string) bool {
	return l.Schedule(task, time.Second) <= 0
}

// AllowSlice returns true if task may execute for the slice duration at time.Now()
func AllowSlice(l Limiter, task string, slice time.Duration) bool {
	return l.Schedule(task, slice) <= 0
}

// New returns a limiter that allows task to run for the specified quantum
// Calls to Allow and AllowSlice reduce a task's available quantum if that
// task is allowed to run. The quantum is replenished naturally via the passage
// of time.
func New(quantum time.Duration) *limiter {
	l := &limiter{
		quantum:  quantum,
		schedule: make(chan ask, 1),
		closecap: make(chan bool, 1),
		done:     make(chan bool),
	}
	l.closecap <- true
	go l.run()
	return l
}

// limiter is a rate limiter
type limiter struct {
	quantum        time.Duration
	schedule       chan ask
	closecap, done chan bool
}

// Schedule schedules the task to run for the given time slice if there is quantum. See interface
// documentation.
func (l *limiter) Schedule(task string, slice time.Duration) (delay time.Duration) {
	reply := make(chan time.Duration, 1)
	l.schedule <- ask{
		string:   task,
		Duration: slice,
		reply:    reply,
	}
	return <-reply
}

func (l *limiter) Quantum() time.Duration {
	return l.quantum
}

// Close releases the rate limiter's resources.
func (l *limiter) Close() error {
	select {
	case first := <-l.closecap:
		if first {
			close(l.closecap)
			close(l.done)
		}
	default:
	}
	return nil
}

func (l *limiter) run() {
	m := make(map[string]time.Time, preallocEntries)
	tick := time.NewTicker(tickInterval)

	defer close(l.schedule)
	defer tick.Stop()

	for {
		select {
		case ask := <-l.schedule:
			now := time.Now()
			then := l.floor(m[ask.string], now).Add(ask.Duration)
			delta := then.Sub(now)
			ask.reply <- delta
			if delta <= 0 {
				m[ask.string] = then
			}
		case <-tick.C:
			select {
			case <-l.done:
				return
			default:
			}

			// TODO(as): The best number is probably not the current MaxSweep
			i := 0
			t := time.Now()
			for k, v := range m {
				if l.floor(v, t) != v {
					delete(m, k)
				}
				if i >= maxSweep {
					break
				}
				i++
			}
		}
	}
}

// floor returns the mark time clamped to [now-window, +inf)
func (l *limiter) floor(mark time.Time, now time.Time) time.Time {
	if t := now.Add(-l.quantum); !mark.After(t) {
		return t
	}
	return mark
}

type ask struct {
	string
	time.Duration
	reply chan time.Duration
}
