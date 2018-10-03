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

type testSink struct {
	events    []*Event
	closed    bool
	retryTime time.Duration

	err error
}

func (s *testSink) Close() error {
	s.closed = true
	return nil
}

func (s *testSink) Send(event *Event) error {
	if s.err == nil {
		s.events = append(s.events, event)
	}
	return s.err
}

func (s *testSink) SetRetryTime(retry time.Duration) error {
	s.retryTime = retry
	return s.err
}

func TestMessage(t *testing.T) {
	var w testSink
	expected := &Event{Data: "Test"}

	assert.NoError(t, Message(&w, "Test"))
	assert.Equal(t, expected, w.events[0])

	w.err = errors.New("Test Error")
	assert.EqualError(t, Message(&w, "Test"), "Test Error")
}
