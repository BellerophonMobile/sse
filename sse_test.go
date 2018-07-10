package sse

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testFlusher struct {
	flushed bool
}

func (f *testFlusher) Flush() {
	f.flushed = true
}

type badWriter struct {
	n   int
	err error
}

func (w *badWriter) Write(p []byte) (int, error) {
	return w.n, w.err
}

type testWriter struct {
	events    []*Event
	closed    bool
	retryTime time.Duration

	err error
}

func (w *testWriter) Close() {
	w.closed = true
}

func (w *testWriter) Send(event *Event) error {
	if w.err == nil {
		w.events = append(w.events, event)
	}
	return w.err
}

func (w *testWriter) SetRetryTime(retry time.Duration) error {
	w.retryTime = retry
	return w.err
}

func TestMessage(t *testing.T) {
	var w testWriter
	expected := &Event{Data: "Test"}

	assert.NoError(t, Message(&w, "Test"))
	assert.Equal(t, expected, w.events[0])

	w.err = errors.New("Test Error")
	assert.EqualError(t, Message(&w, "Test"), "Test Error")
}
