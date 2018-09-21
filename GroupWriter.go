package sse

import (
	"container/list"
	"time"
)

// A GroupWriter groups multiple writers together, sending events to all of
// them.  The zero value is ready to use.
//
// A history of events can be kept, and sent to clients upon connection.
//
// Writers are removed from the group if any errors occur writing to them.
type GroupWriter struct {
	// Keep the most recent number of events, in order to send to new writers
	// joining the group.
	HistoryLimit int

	writers list.List
	history list.List
}

// Close releases any resources associated with this writer. This will also
// close all Writers in the group. This writer should no longer be used after
// this is called.
func (g *GroupWriter) Close() {
	g.forEachWriter(func(w Writer) error {
		w.Close()
		return nil
	})

	g.writers.Init()
	g.history.Init()
}

// Subscribe adds a writer to this group. It sends any available history to the
// writer. It returns an unsubscription function on success, or an error if
// there was a failure sending history.
func (g *GroupWriter) Subscribe(w Writer, lastEventID string) (func(), error) {
	for el := g.findHistory(lastEventID); el != nil; el = el.Next() {
		if err := w.Send(el.Value.(*Event)); err != nil {
			return nil, err
		}
	}

	el := g.writers.PushBack(w)

	return func() {
		g.writers.Remove(el)
	}, nil
}

// Send writes an event to all writers in the group. ID and Type can be empty
// strings, in which case, they will not be sent. If ID is a single whitespace
// character, an empty ID will be sent to reset the client's last-event-id. If
// Data is an empty string, a value-less "data" event will be sent.
//
// If an error occurs writing to one of the writers in the group, it will be
// removed. The first of such an error will be returned, if any.
func (g *GroupWriter) Send(evt *Event) error {
	g.pushHistory(evt)

	return g.forEachWriter(func(w Writer) error {
		return w.Send(evt)
	})
}

// SetRetryTime sets the client-side reconnection timeout for SSE-enabled
// clients. If the client is a browser and gets disconnected, it will try to
// reconnect after this long. If the client doesn't support SSE, this is a
// no-op.
//
// If an error occurs writing to one of the writers in the group, it will be
// removed. The first of such an error will be returned, if any.
func (g *GroupWriter) SetRetryTime(t time.Duration) error {
	return g.forEachWriter(func(w Writer) error {
		return w.SetRetryTime(t)
	})
}

func (g *GroupWriter) forEachWriter(fn func(w Writer) error) error {
	var firstErr error

	for el := g.writers.Front(); el != nil; el = el.Next() {
		writer := el.Value.(Writer)

		if err := fn(writer); err != nil {
			g.writers.Remove(el)

			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (g *GroupWriter) findHistory(id string) *list.Element {
	if id == "" {
		return g.history.Front()
	}

	for el := g.history.Front(); el != nil; el = el.Next() {
		if el.Value.(*Event).ID == id {
			return el.Next()
		}
	}

	return g.history.Front()
}

func (g *GroupWriter) pushHistory(event *Event) {
	g.history.PushBack(event)

	for g.history.Len() > g.HistoryLimit {
		g.history.Remove(g.history.Front())
	}
}
