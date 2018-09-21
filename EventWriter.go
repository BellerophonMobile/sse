package sse

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SSE fields are separated by newlines. If a data value has newlines in it,
// we need to replace with multiple data fields.
var sseDataReplacer = strings.NewReplacer(
	// This case needs to be first, since it contains the other two.
	"\r\n", "\ndata:",
	"\r", "\ndata:",
	"\n", "\ndata:",
)

// EventWriter is a Writer that writes events in the SSE format.
type EventWriter struct {
	// Writer is the underlying io.Writer to write to.
	Writer io.Writer

	// If non-nil, this is used to flush the stream after each event is written.
	Flusher http.Flusher
}

// Close implements the Writer interface, and is a no-op for EventWriter unless
// the underlying writer is also a WriteCloser.
func (w *EventWriter) Close() error {
	if closer, ok := w.Writer.(io.WriteCloser); ok {
		return closer.Close()
	}
	return nil
}

// Send writes an event to the underlying writer. ID and Type can be empty
// strings, in which case, they will not be sent. If ID is a single whitespace
// character, an empty ID will be sent to reset the client's last-event-id. If
// Data is an empty string, a value-less "data" event will be sent.
func (w *EventWriter) Send(evt *Event) error {
	if w.Flusher != nil {
		defer w.Flusher.Flush()
	}

	if evt.ID == " " {
		// If ID is a single whitespace character, reset the ID.
		if _, err := fmt.Fprintf(w.Writer, "id\n"); err != nil {
			return err
		}

	} else if evt.ID != "" {
		// Otherwise, if it's set to anything else, set the ID to it.
		if _, err := fmt.Fprintf(w.Writer, "id:%s\n", evt.ID); err != nil {
			return err
		}
	}

	if evt.Type != "" {
		if _, err := fmt.Fprintf(w.Writer, "event:%s\n", evt.Type); err != nil {
			return err
		}
	}

	if evt.Data == "" {
		fmt.Fprintf(w.Writer, "data")
	} else {
		if _, err := fmt.Fprintf(w.Writer, "data:"); err != nil {
			return err
		}

		// Be slightly more efficient here by writing the replaced string directly,
		// instead of producing the replaced string first.
		if _, err := sseDataReplacer.WriteString(w.Writer, evt.Data); err != nil {
			return nil
		}
	}

	_, err := fmt.Fprintf(w.Writer, "\n\n")
	return err
}

// SetRetryTime sets the client-side reconnection timeout for SSE-enabled
// clients. If the client is a browser and gets disconnected, it will try to
// reconnect after this long. If the client doesn't support SSE, this is a
// no-op.
func (w *EventWriter) SetRetryTime(duration time.Duration) error {
	if w.Flusher != nil {
		defer w.Flusher.Flush()
	}
	_, err := fmt.Fprintf(w.Writer, "retry:%d\n\n", duration/time.Millisecond)
	return err
}
