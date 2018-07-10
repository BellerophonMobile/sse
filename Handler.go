package sse

import (
	"errors"
	"net/http"
	"time"
)

const (
	headerCacheControl = "Cache-Control"
	headerConnection   = "Connection"
	headerContentType  = "Content-Type"
	headerLastEventID  = "Last-Event-ID"
)

const (
	cacheControlNoCache = "no-cache"
	connectionKeepAlive = "keep-alive"
)

// ErrHandlerClosed is returned when calling Send on a closed Handler.
var ErrHandlerClosed = errors.New("handler closed")

// HandlerConfig is optional configuration for the Handler.
type HandlerConfig struct {
	// RetryTime sets the client-side reconnection timeout for SSE-enabled clients.
	// If the client disconnects, it will try to reconnect after this long. If the
	// client doesn't support SSE, this is a no-op. If zero, this is not sent to
	// clients.
	RetryTime time.Duration

	// WriteTimeout sets the maximum time to write a single event.
	WriteTimeout time.Duration

	// BufferSize sets how many events can be queued to a single client. If zero,
	// sending events will block until sent.
	BufferSize int

	// HistoryLimit sets the number of events that will be re-played to clients
	// upon connecting. Will take into account the Last-Event-ID header to attempt
	// to skip over events the client definitely has.
	HistoryLimit int
}

// Handler implements http.Handler and EventWriter, and allows multiple clients
// to receive events.
//
// A history of events can be kept, and sent to clients upon connection. The
// history sending will respect the Last-Event-ID header if set, to not send
// events the client already has.
//
// Events can also be buffered per-client, so a slow client doesn't block
// sending events. In this case, a write deadline should also be set on the
// request before sending it to this handler.
type Handler struct {
	WriteTimeout time.Duration
	BufferSize   int

	group GroupWriter

	closed    bool
	closeChan chan struct{}
}

// NewHandler constructs a new Handler with the given options.
func NewHandler(config HandlerConfig) *Handler {
	return &Handler{
		WriteTimeout: config.WriteTimeout,
		BufferSize:   config.BufferSize,
		group: GroupWriter{
			RetryTime: config.RetryTime,
			history:   history{limit: config.HistoryLimit},
		},
		closeChan: make(chan struct{}),
	}
}

// Close disconnects all connected clients from the handler. The handler should
// no longer be used after this is called.
func (h *Handler) Close() {
	if h.closed {
		return
	}

	h.closed = true
	h.group.Close()
	close(h.closeChan)
}

// Send writes an event to this sink. ID and event can be empty strings, in
// which case, they will not be sent. If id is a single whitespace character,
// an empty ID will be sent to reset the client's last-event-id.
func (h *Handler) Send(event *Event) error {
	if h.closed {
		return ErrHandlerClosed
	}

	return h.group.Send(event)
}

// ServeHTTP implements http.Handler, responding to HTTP requests by listening
// on an SSE sink.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.closed {
		// Sending this status code will prevent browser clients from attempting to
		// re-connect.
		w.WriteHeader(http.StatusNoContent)
		return
	}

	closeNotifier, ok := w.(http.CloseNotifier)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer := NewWriter(w, r)

	if h.BufferSize > 0 || h.WriteTimeout > 0 {
		writer = NewBufferWriter(writer, h.BufferSize, h.WriteTimeout)
	}

	lastEventID := r.Header.Get(headerLastEventID)
	unsubscribe, err := h.group.Subscribe(writer, lastEventID)
	if err != nil {
		return
	}
	defer unsubscribe()

	select {
	case <-closeNotifier.CloseNotify():
	case <-h.closeChan:
	}
}
