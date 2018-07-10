package sse

import (
	"container/list"
	"sync"
	"time"
)

// A GroupWriter groups multiple writers together, sending events to all of
// them.  The zero value is ready to use.
//
// A history of events can be kept, and sent to clients upon connection. The
// history sending will respect the Last-Event-ID header if set, to not send
// events the client already received.
type GroupWriter struct {
	RetryTime time.Duration

	writers    list.List
	writerLock sync.Mutex

	history     history
	historyLock sync.RWMutex
}

// Close releases any resources associated with this writer. This will also
// close all Writers in the group. This writer should no longer be used after
// this is called.
func (g *GroupWriter) Close() {
	g.writerLock.Lock()
	defer g.writerLock.Unlock()

	g.historyLock.Lock()
	defer g.historyLock.Unlock()

	for el := g.writers.Front(); el != nil; el = el.Next() {
		writer := el.Value.(Writer)
		writer.Close()
	}

	g.writers.Init()
	g.history.items.Init()
}

// Subscribe adds a writer to this group. It sends any available history to the
// writer. It returns an unsubscription function on success, or an error if
// there was a failure sending history.
func (g *GroupWriter) Subscribe(w Writer, lastEventID string) (func(), error) {
	if g.RetryTime > 0 {
		if err := w.SetRetryTime(g.RetryTime); err != nil {
			return nil, err
		}
	}

	g.historyLock.RLock()
	defer g.historyLock.RUnlock()

	for el := g.history.find(lastEventID); el != nil; el = el.Next() {
		if err := w.Send(el.Value.(*Event)); err != nil {
			return nil, err
		}
	}

	g.writerLock.Lock()
	defer g.writerLock.Unlock()

	wEl := g.writers.PushBack(w)

	return func() {
		g.writerLock.Lock()
		defer g.writerLock.Unlock()
		g.writers.Remove(wEl)
	}, nil
}

// Send writes an event to this sink. ID and event can be empty strings, in
// which case, they will not be sent. If id is a single whitespace character,
// an empty ID will be sent to reset the client's last-event-id.
//
// If an error occurs writing to one of the writers in the group, it will be
// removed. The first of such an error will be returned, if any.
func (g *GroupWriter) Send(event *Event) error {
	g.pushHistory(event)

	g.writerLock.Lock()
	defer g.writerLock.Unlock()

	var firstErr error

	for el := g.writers.Front(); el != nil; el = el.Next() {
		writer := el.Value.(Writer)
		if err := writer.Send(event); err != nil {
			g.writers.Remove(el)

			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// SetRetryTime sets the client-side reconnection timeout for SSE-enabled
// clients. If the client disconnects, it will try to reconnect after this long.
// If the client doesn't support SSE, this is a no-op.
func (g *GroupWriter) SetRetryTime(time time.Duration) error {
	g.RetryTime = time

	g.writerLock.Lock()
	defer g.writerLock.Unlock()

	var firstErr error

	for el := g.writers.Front(); el != nil; el = el.Next() {
		writer := el.Value.(Writer)
		if err := writer.SetRetryTime(time); err != nil {
			g.writers.Remove(el)

			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (g *GroupWriter) pushHistory(event *Event) {
	g.historyLock.Lock()
	defer g.historyLock.Unlock()

	g.history.append(event)
}
