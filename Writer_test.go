package sse

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWriter_plain(t *testing.T) {
	assertions := assert.New(t)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	resp := httptest.NewRecorder()

	w := NewWriter(resp, req)

	assertions.Equal("no-cache", resp.Header().Get("Cache-Control"))
	assertions.Equal("keep-alive", resp.Header().Get("Connection"))
	assertions.Equal("text/plain", resp.Header().Get("Content-Type"))

	if _, ok := w.(*PlainWriter); !ok {
		t.Error("Writer is not a PlainWriter")
	}
}

func TestNewWriter_sse(t *testing.T) {
	assertions := assert.New(t)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	resp := httptest.NewRecorder()

	req.Header.Set("Accept", "text/event-stream")

	w := NewWriter(resp, req)

	assertions.Equal("no-cache", resp.Header().Get("Cache-Control"))
	assertions.Equal("keep-alive", resp.Header().Get("Connection"))
	assertions.Equal("text/event-stream", resp.Header().Get("Content-Type"))

	if _, ok := w.(*EventWriter); !ok {
		t.Error("Writer is not a EventWriter")
	}
}
