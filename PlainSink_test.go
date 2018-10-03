package sse

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlainSinkSend_noflush(t *testing.T) {
	var buf bytes.Buffer

	w := PlainSink{Writer: &buf}

	assert.NoError(t, w.Send(&Event{Data: "one"}))
	assert.NoError(t, w.Send(&Event{ID: "id2", Data: "two"}))
	assert.NoError(t, w.Send(&Event{ID: "id3", Data: "three", Type: "typeThree"}))

	expected := "one\n\ntwo\n\nthree\n\n"

	assert.Equal(t, expected, buf.String())
}

func TestPlainSinkSend_flush(t *testing.T) {
	var buf bytes.Buffer
	var flusher testFlusher

	w := PlainSink{
		Writer:  &buf,
		Flusher: &flusher,
	}

	assert.NoError(t, w.Send(&Event{Data: "one"}))
	assert.True(t, flusher.flushed)
	flusher.flushed = false

	assert.NoError(t, w.Send(&Event{ID: "id2", Data: "two"}))
	assert.True(t, flusher.flushed)
	flusher.flushed = false

	assert.NoError(t, w.Send(&Event{ID: "id3", Data: "three", Type: "typeThree"}))
	assert.True(t, flusher.flushed)
	flusher.flushed = false

	expected := "one\n\ntwo\n\nthree\n\n"

	assert.Equal(t, expected, buf.String())
}

func TestPlainSinkSend_error(t *testing.T) {
	w := PlainSink{
		Writer: &badWriter{err: errors.New("test write error")},
	}

	assert.EqualError(t, w.Send(&Event{Data: "test"}), "test write error")
}

func TestPlainSinkClose(t *testing.T) {
	var w PlainSink
	// Close should be a no-op on a PlainSink.
	assert.NotPanics(t, func() { w.Close() })
}

func TestPlainSinkSetRetryTime(t *testing.T) {
	var w PlainSink
	// SetRetryTime should be a no-op on a PlainSink.
	assert.Nil(t, w.SetRetryTime(10*time.Second))
}
