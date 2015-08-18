package sse

import (
	"net/http"
	"io"
	"bufio"
	"bytes"
	"fmt"
)

type Reader struct {
	Response *http.Response
	StatusCode int

	bytereader *bufio.Reader
}

type Event struct {
	Tag string
	Data []byte
}

func NewReader(method string, url string, reader io.Reader) (*Reader,error) {

	x := &Reader{}
	
	client := &http.Client{}
	req,err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil,err
	}

	req.Header.Set("Accept", "text/event-stream")

	x.Response, err = client.Do(req)
	if err != nil {
		return nil,err
	}

	x.StatusCode = x.Response.StatusCode

	x.bytereader = bufio.NewReader(x.Response.Body)

	return x,nil
}


func (x *Reader) Next() (*Event,error) {

	event := &Event{}

	var line []byte
	var err error
	
	for {
		line, err = x.bytereader.ReadBytes('\n')
		if err != nil {
			return nil,err
		}
		if !bytes.Equal(line, []byte("\n")) {
			break
		}
	}
	fmt.Printf("'%s'", line)
	
	if !bytes.HasPrefix(line, []byte("event:")) {
		return nil, fmt.Errorf("Expected event tag, got '%s'", line)
	}
	event.Tag = string(bytes.TrimSpace(line[6:]))

	line, err = x.bytereader.ReadBytes('\n')
	if err != nil {
		return nil,err
	}
	fmt.Printf("'%s'", line)
	
	i := bytes.Index(line, []byte{':'})
	if i == -1 {
		return nil, fmt.Errorf("Expected event data label, got '%s'", line)
	}

	event.Data = bytes.TrimSpace(line[i+1:])

	return event, nil

}

func (x *Reader) Close() {
	x.Response.Body.Close()
}
