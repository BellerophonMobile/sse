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

// EventSink is a Sink that writes events in the SSE format.
type EventSink struct {
	// Writer is the underlying io.Writer to write to.
	Writer io.Writer

	// If non-nil, this is used to flush the stream after each event is written.
	Flusher http.Flusher
}

// Close implements the Sink interface, and is a no-op for EventSink unless
// the underlying writer is also a WriteCloser.
func (s *EventSink) Close() error {
	if closer, ok := s.Writer.(io.WriteCloser); ok {
		return closer.Close()
	}
	return nil
}

// Send writes an event to the underlying writer. ID and Type can be empty
// strings, in which case, they will not be sent. If ID is a single whitespace
// character, an empty ID will be sent to reset the client's last-event-id. If
// Data is an empty string, a value-less "data" event will be sent.
func (s *EventSink) Send(evt *Event) error {
	if s.Flusher != nil {
		defer s.Flusher.Flush()
	}

	if evt.ID == " " {
		// If ID is a single whitespace character, reset the ID.
		if _, err := fmt.Fprintf(s.Writer, "id\n"); err != nil {
			return err
		}

	} else if evt.ID != "" {
		// Otherwise, if it's set to anything else, set the ID to it.
		if _, err := fmt.Fprintf(s.Writer, "id:%s\n", evt.ID); err != nil {
			return err
		}
	}

	if evt.Type != "" {
		if _, err := fmt.Fprintf(s.Writer, "event:%s\n", evt.Type); err != nil {
			return err
		}
	}

	if evt.Data == "" {
		fmt.Fprintf(s.Writer, "data")
	} else {
		if _, err := fmt.Fprintf(s.Writer, "data:"); err != nil {
			return err
		}

		// Be slightly more efficient here by writing the replaced string directly,
		// instead of producing the replaced string first.
		if _, err := sseDataReplacer.WriteString(s.Writer, evt.Data); err != nil {
			return nil
		}
	}

	_, err := fmt.Fprintf(s.Writer, "\n\n")
	return err
}

// SetRetryTime sets the client-side reconnection timeout for SSE-enabled
// clients. If the client is a browser and gets disconnected, it will try to
// reconnect after this long. If the client doesn't support SSE, this is a
// no-op.
func (s *EventSink) SetRetryTime(duration time.Duration) error {
	if s.Flusher != nil {
		defer s.Flusher.Flush()
	}
	_, err := fmt.Fprintf(s.Writer, "retry:%d\n\n", duration/time.Millisecond)
	return err
}
