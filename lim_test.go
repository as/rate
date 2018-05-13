package rate

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestLimiterBasic(t *testing.T) {
	l := New(time.Second*30)
	n := 0
	for ; n < 100000; n++ {
		if !Allow(l, "bar") {
			break
		}
	}
	if n != 30 {
		t.Fatalf("bad request count: want 30, have %d", n)
	}
}

func TestLimiterSchedule(t *testing.T) {
	l := New(time.Second*2)
	Allow(l,"a")
	Allow(l,"a")
	delay := l.Schedule("a", time.Second)
	if delay.Truncate(time.Second) != time.Second{
		t.Fatalf("bad delay: want ~1s, have %s", delay)
	}
}

func TestLimiterReplenish(t *testing.T) {
	l := New(time.Second*3)
	Allow(l,"bar")
	Allow(l,"bar")
	Allow(l,"bar")
	Allow(l,"bar")
	Allow(l,"bar")
	Allow(l,"bar")
	if Allow(l,"bar") {
		t.Fatalf("1/3: have allow, want deny")
	}
	time.Sleep(time.Second)
	if !Allow(l,"bar") {
		t.Fatalf("2/3: have deny, want allow")
	}
	if Allow(l,"bar") {
		t.Fatalf("3/3: have allow, want deny")
	}
}

func TestLimiterMulti(t *testing.T) {
	l := New(time.Second*30)
	for i := 0; i < 300; i++ {
		go Allow(l, "baz")
	}

	n := 0
	for ; n < 100000; n++ {
		if !Allow(l,"bar") {
			break
		}
	}
	if n != 30 {
		t.Fatalf("bad request count: want 30, have %d", n)
	}
}

func BenchmarkLimiter(b *testing.B) {
	l := New(time.Second*30)
	body := func(pb *testing.PB) {
		name := fmt.Sprint(rand.Int31n(7))
		for pb.Next() {
			Allow(l,name)
		}
	}
	b.RunParallel(body)
}
