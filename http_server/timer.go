// 定时器相关 author chenweidong

package http_server

import (
	"time"
)

func NewTimer(duration time.Duration, f func(), name string) *Timer {
	return &Timer{duration: duration, f: f, Name: name}
}

type Timer struct {
	Name     string
	duration time.Duration
	f        func()
	ticker   *time.Ticker
}

func (t *Timer) Start() {
	t.ticker = time.NewTicker(t.duration)
	LOG.Info("%s timer started", t.Name)
	go func() {
		for true {
			t.f()
			<-t.ticker.C
		}
	}()
}

func (t *Timer) Stop() {
	if t.ticker == nil {
		return
	}
	t.ticker.Stop()
	LOG.Info("%s timer stopped", t.Name)
}
