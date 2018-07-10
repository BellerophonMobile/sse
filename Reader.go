package sse

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"
)

// Reader wraps a normal io.Reader to read SSE events.
type Reader struct {
	RetryTime   time.Duration
	LastEventID string

	scanner *bufio.Scanner

	idBuf   string
	typeBuf string
	dataBuf strings.Builder
}

// NewReader wraps the given io.Reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		scanner: bufio.NewScanner(r),
	}
}

// NextEvent reads the next SSE event from the reader.
func (r *Reader) NextEvent() (*Event, error) {
	// From section 9.2.5 of the spec
	// https://html.spec.whatwg.org/multipage/server-sent-events.html#event-stream-interpretation
	for r.scanner.Scan() {
		line := r.scanner.Bytes()

		// If the line is empty (a blank line), dispatch the event.
		if len(line) == 0 {
			if evt := r.dispatch(); evt != nil {
				return evt, nil
			}
			continue
		}

		// If the line starts with a COLON character, ignore the line.
		if line[0] == ':' {
			continue
		}

		// If the line contains a COLON character...
		if idx := bytes.IndexByte(line, ':'); idx != -1 {
			// Collect the characters before the first COLON as field.
			field := string(line[:idx])

			// Collect the characters after the first COLON as value.
			value := line[idx+1:]

			// If value starts with a SPACE character, remove it.
			if len(value) > 0 && value[0] == ' ' {
				value = value[1:]
			}

			r.process(field, value)
			continue
		}

		// Otherwise, process with the whole line as field name
		r.process(string(line), []byte{})
	}

	// In-progress events are not dispatched.

	return nil, r.scanner.Err()
}

func (r *Reader) process(field string, value []byte) {
	switch field {
	case "event":
		r.typeBuf = string(value)

	case "data":
		r.dataBuf.Write(value)
		r.dataBuf.WriteByte('\n')

	case "id":
		if bytes.IndexByte(value, 0) == -1 {
			r.idBuf = string(value)
		}

	case "retry":
		value := string(value)
		if retry, err := strconv.Atoi(value); err == nil {
			r.RetryTime = time.Duration(retry) * time.Millisecond
		}
	}
}

func (r *Reader) dispatch() *Event {
	// Set last event ID string, the buffer does not get reset.
	r.LastEventID = r.idBuf

	// If the data buffer is an empty string, set data and type to empty
	// string and stop.
	if r.dataBuf.Len() == 0 {
		r.typeBuf = ""
		return nil
	}

	data := r.dataBuf.String()

	// If the data buffer's last character is a \n character, remove it.
	if data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	// Initialize message type to "message"
	evtType := "message"

	if r.typeBuf != "" {
		evtType = r.typeBuf
	}

	event := &Event{ID: r.LastEventID, Type: evtType, Data: data}

	// Set the data buffer and event type buffer to the empty string.
	r.typeBuf = ""
	r.dataBuf.Reset()

	return event
}
