package sse

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testSlowWriter struct {
	events    []*Event
	closed    bool
	retryTime time.Duration

	sendTime time.Duration
	err      error

	lock sync.Mutex
}

func (w *testSlowWriter) Close() {
	w.closed = true
}

func (w *testSlowWriter) Send(event *Event) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	time.Sleep(w.sendTime)

	if w.err == nil {
		w.events = append(w.events, event)
	}

	return w.err
}

func (w *testSlowWriter) SetRetryTime(retry time.Duration) error {
	w.retryTime = retry
	return w.err
}
func TestBufferWriter_fast(t *testing.T) {
	var w testSlowWriter

	b := NewBufferWriter(&w, 0, 0)

	assert.NoError(t, b.Send(&Event{Data: "test"}))

	w.lock.Lock()
	defer w.lock.Unlock()
	assert.Equal(t, []*Event{{Data: "test"}}, w.events)
}
