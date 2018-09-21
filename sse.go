// Package sse implements the HTML Server-Sent events specification. It provides
// a Reader to parse SSE events from an io.Reader, and several Writers to manage
// dispatching SSE events to clients.
package sse

// MIME Types for HTTP Content-Type and Accept headers.
const (
	MIMETypePlain = "text/plain"
	MIMETypeSSE   = "text/event-stream"
)

// Event is an SSE event. ID and event can be empty strings, in which case, they
// will not be sent. If id is a single whitespace character, an empty ID will be
// sent to reset the client's Last-Event-ID. If Data is an empty string, a
// value-less data field will be sent.
type Event struct {
	ID, Type, Data string
}

// Message is a convenience function for sending an event without a type or ID.
func Message(w Writer, data string) error {
	return w.Send(&Event{Data: data})
}
