package sse

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// PlainSink is a Sink that sends events in plain text. The data of each event
// is written separated by two newline characters. IDs and event types are not
// sent.
type PlainSink struct {
	// Writer is the underlying io.Writer to write to.
	Writer io.Writer

	// If non-nil, this is used to flush after each event is written.
	Flusher http.Flusher
}

// Close implements the Sink interface, and is a no-op for PlainSink unless
// the underlying writer is also a WriteCloser.
func (s *PlainSink) Close() error {
	if closer, ok := s.Writer.(io.WriteCloser); ok {
		return closer.Close()
	}
	return nil
}

// Send sends the event to the client, returning an error if any. Send writes an
// event the underlying writer. Only the Data field is written by PlainSink.
// If Data is an empty string, nothing is sent.
func (s *PlainSink) Send(evt *Event) error {
	if evt.Data == "" {
		return nil
	}

	if s.Flusher != nil {
		defer s.Flusher.Flush()
	}

	_, err := fmt.Fprintf(s.Writer, "%s\n\n", evt.Data)
	return err
}

// SetRetryTime is a no-op for non-SEE clients.
func (s *PlainSink) SetRetryTime(duration time.Duration) error {
	return nil
}
