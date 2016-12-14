package sse

import (
//	"fmt"
	"net/http"
	"encoding/json"
	"log"
)

type EventServerOptions struct {
	ActionQueueLength int
	LogLabel string
	RetryMillis int
}

type EventServer struct {
	label string
	actions chan func()
	connections map[*Writer]*Writer
	retrymillis int
}

func NewEventServer(opts *EventServerOptions) *EventServer {
	if opts == nil {
		opts = &EventServerOptions{
		}
	}

	if opts.LogLabel == "" {
		opts.LogLabel = "sse"
	}
	
	x := &EventServer{
		label: opts.LogLabel,
		actions: make(chan func(), opts.ActionQueueLength),
		connections: make(map[*Writer]*Writer),
		retrymillis: opts.RetryMillis,
	}

	log.Printf("[%v] New EventServer", x.label)
	
	go x.process()

	return x
}

func (x *EventServer) process() {
	for {
		// log.Printf("[%v] Waiting for action", x.label)
		action := <- x.actions
		action()
	}
}

func (x *EventServer) Handle(w http.ResponseWriter, r *http.Request) {

	conn,err := NewWriter(w, r, x.retrymillis)
	if err != nil {
		log.Printf("[%v] Error handling connection: %v", x.label, err)
		return
	}

	x.actions <- func() {
		log.Printf("[%v] Attach", x.label)
		x.connections[conn] = conn
	}

	notify := w.(http.CloseNotifier).CloseNotify()
	<- notify

	x.actions <- func() {
		log.Printf("[%v] Detach", x.label)
		delete(x.connections, conn)
	}

}

func (x *EventServer) Message(msg string) {
	x.Event("", "", msg)
}

func (x *EventServer) Event(id, event, msg string) {

	x.actions <- func() {
		// log.Printf("[%v] Message '%v': '%v'", x.label, event, msg)
		for writer := range(x.connections) {
			_, err := writer.Event(id, event, msg)
			if err != nil {
				log.Printf("[%v] Detaching on error: ", x.label, err)
				delete(x.connections, writer)
			}
		}
	}

}

func (x *EventServer) JSONMessage(obj interface{}) {
	x.JSONEvent("", "", obj)
}

func (x *EventServer) JSONEvent(id string, event string, obj interface{}) {

	bits, err := json.Marshal(obj)
	if err != nil {
		log.Printf("[%v] JSON marshaling error: %v", x.label, err)
		return
	}

	x.Event(id, event, string(bits))

}
