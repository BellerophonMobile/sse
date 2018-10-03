package sse

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventSinkSend_noflush(t *testing.T) {
	var buf bytes.Buffer

	w := EventSink{Writer: &buf}

	assert.NoError(t, w.Send(&Event{Data: "one"}))
	assert.NoError(t, w.Send(&Event{ID: "id2", Data: "two"}))
	assert.NoError(t, w.Send(&Event{ID: "id3", Data: "three", Type: "typeThree"}))
	assert.NoError(t, w.Send(&Event{ID: " ", Data: "line1\rline2\nline3\r\nline4"}))

	expected := `data:one

id:id2
data:two

id:id3
event:typeThree
data:three

id
data:line1
data:line2
data:line3
data:line4

`

	assert.Equal(t, expected, buf.String())
}

func TestEventSinkSend_flush(t *testing.T) {
	var buf bytes.Buffer
	var flusher testFlusher

	w := EventSink{
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

	assert.NoError(t, w.Send(&Event{ID: " ", Data: "line1\rline2\nline3\r\nline4"}))
	assert.True(t, flusher.flushed)
	flusher.flushed = false

	expected := `data:one

id:id2
data:two

id:id3
event:typeThree
data:three

id
data:line1
data:line2
data:line3
data:line4

`

	assert.Equal(t, expected, buf.String())
}

func TestEventSinkSend_error(t *testing.T) {
	w := EventSink{Writer: &badWriter{err: errors.New("test write error")}}

	assert.EqualError(t, w.Send(&Event{Data: "test"}), "test write error")
	assert.EqualError(t, w.Send(&Event{ID: "1", Data: "test"}), "test write error")
	assert.EqualError(t, w.Send(&Event{ID: " ", Data: "test"}), "test write error")
	assert.EqualError(t, w.Send(&Event{Type: "foo", Data: "test"}), "test write error")
}

func TestEventSinkClose(t *testing.T) {
	var w EventSink
	// Close should be a no-op on a EventSink.
	assert.NotPanics(t, func() { w.Close() })
}

func TestEventSinkSetRetryTime_noflush(t *testing.T) {
	var buf bytes.Buffer

	w := EventSink{Writer: &buf}

	assert.NoError(t, w.SetRetryTime(10*time.Second))

	expected := "retry:10000\n\n"

	assert.Equal(t, expected, buf.String())
}

func TestEventSinkSetRetryTime_flush(t *testing.T) {
	var buf bytes.Buffer
	var flusher testFlusher

	w := EventSink{
		Writer:  &buf,
		Flusher: &flusher,
	}

	assert.NoError(t, w.SetRetryTime(10*time.Second))
	assert.True(t, flusher.flushed)

	expected := "retry:10000\n\n"

	assert.Equal(t, expected, buf.String())
}
