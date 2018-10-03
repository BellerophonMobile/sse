package sse

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSink_plain(t *testing.T) {
	assertions := assert.New(t)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	resp := httptest.NewRecorder()

	w := NewSink(resp, req)

	assertions.Equal("no-cache", resp.Header().Get("Cache-Control"))
	assertions.Equal("keep-alive", resp.Header().Get("Connection"))
	assertions.Equal("text/plain", resp.Header().Get("Content-Type"))

	if _, ok := w.(*PlainSink); !ok {
		t.Error("Sink is not a PlainSink")
	}
}

func TestNewSink_sse(t *testing.T) {
	assertions := assert.New(t)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	resp := httptest.NewRecorder()

	req.Header.Set("Accept", "text/event-stream")

	w := NewSink(resp, req)

	assertions.Equal("no-cache", resp.Header().Get("Cache-Control"))
	assertions.Equal("keep-alive", resp.Header().Get("Connection"))
	assertions.Equal("text/event-stream", resp.Header().Get("Content-Type"))

	if _, ok := w.(*EventSink); !ok {
		t.Error("Sink is not a EventSink")
	}
}
