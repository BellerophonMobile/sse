package sse

import (
	"fmt"
	"net/http"
	"encoding/json"
)

type Writer struct {
	// Response is the HTTP response object.
	Response http.ResponseWriter
	Flusher http.Flusher
	
	// Request is the incoming HTTP request.  The Accept header is
	// checked to determine what the output Content-Type should be.
	Request *http.Request

	// If no event tag is provided and SSE is enabled, the next write
	// will include this event tag.
	DefaultEvent string

	// True if the connection accepts SSE streams.
	SSE bool
}

var FlushUnsupported = fmt.Errorf("Streaming not supported")

func NewWriter(w http.ResponseWriter, r *http.Request) (*Writer,error) {

	x := &Writer{Response: w, Request: r}

	var ok bool
	x.Flusher,ok = w.(http.Flusher)
	if !ok {
		return nil,FlushUnsupported
	}
	
	if r.Header.Get("Accept") == "text/event-stream" {
		w.Header().Add("Content-Type", "text/event-stream")
		x.SSE = true
	}

	return x,nil
}

func (x *Writer) Event(event string, p []byte) (int, error) {
	var count, n int
	var err error

	defer x.Flusher.Flush()


	// If stream is not SSE, just print the data
	if !x.SSE {
		n,err = fmt.Fprintf(x.Response, "%s\n", p)
		return n,err
	}


	// Otherwise, it's SSE, include the optional event tag and data prefix	
	if event != "" {
		n, err = fmt.Fprintf(x.Response, "event: %s\n", event)
		count += n
		if err != nil {
			return count, err
		}	
	}

	n, err = fmt.Fprintf(x.Response, "data: %s\n\n", p)
	count += n
	return count, err

}

func (x *Writer) Write(p []byte) (int, error) {
	return x.Event(x.DefaultEvent, p)
}

func (x *Writer) JSONEvent(event string, obj interface{}) (int, error) {
	bits, err := json.Marshal(obj)
	if err != nil {
		return 0,err
	}

	return x.Event(event, bits)
}
