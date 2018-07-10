package sse

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SSE fields are separated by newlines. If a data value has newlines in it,
// need to replace with multiple data fields.
var sseDataReplacer = strings.NewReplacer(
	// This case needs to be first, since it contains the other two.
	"\r\n", "\ndata:",
	"\r", "\ndata:",
	"\n", "\ndata:",
)

// EventWriter is a Writer that sends events in SSE format.
type EventWriter struct {
	// Writer is the underlying io.Writer to write to.
	Writer io.Writer

	// If non-nil, this is used to flush after each event is written.
	Flusher http.Flusher
}

// Close implements the Writer interface, and is a no-op here.
func (w *EventWriter) Close() {}

// Send sends the event to the client, returning an error if any.
func (w *EventWriter) Send(event *Event) error {
	if w.Flusher != nil {
		defer w.Flusher.Flush()
	}

	if event.ID == " " {
		// If ID is a single whitespace character, reset the ID.
		if _, err := fmt.Fprintf(w.Writer, "id\n"); err != nil {
			return err
		}

	} else if event.ID != "" {
		// Otherwise, if it's set to anything else, set the ID to it.
		if _, err := fmt.Fprintf(w.Writer, "id:%s\n", event.ID); err != nil {
			return err
		}
	}

	if event.Type != "" {
		if _, err := fmt.Fprintf(w.Writer, "event:%s\n", event.Type); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w.Writer, "data:"); err != nil {
		return err
	}

	// Be slightly more efficient here by writing the replaced string directly,
	// instead of producing the replaced string first.
	if _, err := sseDataReplacer.WriteString(w.Writer, event.Data); err != nil {
		return nil
	}

	_, err := fmt.Fprintf(w.Writer, "\n\n")
	return err
}

// SetRetryTime sets the client-side reconnection timeout for SSE-enabled
// clients. If the client disconnects, it will try to reconnect after this long.
// If the client doesn't support SSE, this is a no-op.
func (w *EventWriter) SetRetryTime(duration time.Duration) error {
	if w.Flusher != nil {
		defer w.Flusher.Flush()
	}
	_, err := fmt.Fprintf(w.Writer, "retry:%d\n\n", duration/time.Millisecond)
	return err
}
