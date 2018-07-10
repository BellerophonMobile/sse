package sse

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// PlainWriter is a EventWriter that sends events in plain text. The data of
// each event is written separated by two newline characters. IDs and event
// types are not sent.
type PlainWriter struct {
	// Writer is the underlying io.Writer to write to.
	Writer io.Writer

	// If non-nil, this is used to flush after each event is written.
	Flusher http.Flusher
}

// Close implements the Writer interface, and is a no-op here.
func (w *PlainWriter) Close() {}

// Send sends the event to the client, returning an error if any.
func (w *PlainWriter) Send(event *Event) error {
	if w.Flusher != nil {
		defer w.Flusher.Flush()
	}

	_, err := fmt.Fprintf(w.Writer, "%s\n\n", event.Data)
	return err
}

// SetRetryTime is a no-op for non-SEE clients.
func (w *PlainWriter) SetRetryTime(duration time.Duration) error {
	return nil
}
