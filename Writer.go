package sse

import (
	"fmt"
	"net/http"
	"strings"
)

type Writer struct {

	// Response is the HTTP response object.
	Response http.ResponseWriter
	Flusher http.Flusher
	
	// Request is the incoming HTTP request.  The Accept header is
	// checked to determine what the output Content-Type should be.
	Request *http.Request

	// True if the connection accepts SSE streams.
	SSE bool

}

var FlushUnsupported = fmt.Errorf("Streaming not supported")

func NewWriter(w http.ResponseWriter, r *http.Request, retrymillis int) (*Writer,error) {

	x := &Writer{
		Response: w,
		Request: r,
	}

	var ok bool
	x.Flusher,ok = w.(http.Flusher)
	if !ok {
		return nil,FlushUnsupported
	}
	
	if r.Header.Get("Accept") == "text/event-stream" {
		w.Header().Add("Content-Type", "text/event-stream")
		x.SSE = true
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if retrymillis != 0 {
		_, err := fmt.Fprintf(x.Response, "retry: %s\n\n", retrymillis)
		if err != nil {
			return nil, err
		}
	}
	
	return x,nil
}

func (x *Writer) Event(id string, event string, data string) (int, error) {
	var count, n int
	var err error

	defer x.Flusher.Flush()

	// If stream is not SSE, just print the data
	if !x.SSE {
		n,err = fmt.Fprintf(x.Response, "%s\n\n", data)
		return n,err
	}

	// Otherwise, it's SSE

	if id != "" {
		n, err = fmt.Fprintf(x.Response, "id: %s\n", id)
		count += n
		if err != nil {
			return count, err
		}	
	}

	if event != "" {
		n, err = fmt.Fprintf(x.Response, "event: %s\n", event)
		count += n
		if err != nil {
			return count, err
		}	
	}

	data = strings.Replace(data, "\n", "\ndata: ", -1)
	
	n, err = fmt.Fprintf(x.Response, "data: %s\n\n", data)
	count += n
	return count, err

}
