package sse

import (
	"net/http"
	"time"
)

const headerAccept = "Accept"

// Writer is a type capable of sending SSE events.
type Writer interface {
	// Close releases any resources associated with this writer. This writer
	// should no longer be used after this is called.
	Close() error

	// Send writes an event to this writer. ID and Type can be empty strings, in
	// which case, they will not be sent. If ID is a single whitespace character,
	// an empty ID will be sent to reset the client's last-event-id. If Data is
	// an empty string, a value-less "data" event will be sent.
	Send(*Event) error

	// SetRetryTime sets the client-side reconnection timeout for SSE-enabled
	// clients. If the client is a browser and gets disconnected, it will try to
	// reconnect after this long. If the client doesn't support SSE, this is a
	// no-op.
	SetRetryTime(time.Duration) error
}

// NewWriter creates a writer based on the Accept header of the given request.
// It also sets an appropriate Content-Type, Cache-Control, and Connection
// headers.
func NewWriter(w http.ResponseWriter, r *http.Request) Writer {
	isSSE := r.Header.Get(headerAccept) == MIMETypeSSE

	w.Header().Set(headerCacheControl, cacheControlNoCache)
	w.Header().Set(headerConnection, connectionKeepAlive)

	if !isSSE {
		w.Header().Add(headerContentType, MIMETypePlain)
		writer := &PlainWriter{Writer: w}
		writer.Flusher, _ = w.(http.Flusher)
		return writer
	}

	w.Header().Add(headerContentType, MIMETypeSSE)

	writer := &EventWriter{Writer: w}
	writer.Flusher, _ = w.(http.Flusher)

	return writer
}
