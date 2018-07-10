package sse

import (
	"errors"
	"time"
)

// ErrBufferFull is returned when sending an event to a BufferWriter that is
// full.
var ErrBufferFull = errors.New("buffer writer full")

// BufferWriter is a Writer that can buffer events for slower clients. Also can
// enforce a timeout if the client is too slow.
type BufferWriter struct {
	Writer  Writer
	Timeout time.Duration
	Size    int

	eventChan chan *Event
}

// NewBufferWriter wraps the specified writer with a given buffer size and
// write timeout.
func NewBufferWriter(w Writer, size int, timeout time.Duration) *BufferWriter {
	bufferWriter := &BufferWriter{
		Writer:  w,
		Timeout: timeout,
		Size:    size,

		eventChan: make(chan *Event, size),
	}

	go bufferWriter.loop()

	return bufferWriter
}

// Close releases any resources associated with this writer. This will also
// close the underlying Writer. This writer should no longer be used after this
// is called.
func (w *BufferWriter) Close() {
	w.Writer.Close()
	close(w.eventChan)
}

// Send sends an event to this writer. ID and event can be empty strings, in
// which case, they will not be sent. If id is a single whitespace character,
// an empty ID will be sent to reset the client's last-event-id.
//
// If the buffer size is zero, this will block until the event is consumed by
// the underlying writer. Otherwise, this will not block. If the buffer is full,
// ErrBufferFull will be returned.
func (w *BufferWriter) Send(event *Event) error {
	if w.Size == 0 {
		w.eventChan <- event
		return nil
	}

	select {
	case w.eventChan <- event:
		return nil
	default:
		return ErrBufferFull
	}
}

// SetRetryTime sets the client-side reconnection timeout for SSE-enabled
// clients. If the client disconnects, it will try to reconnect after this
// long. If the client doesn't support SSE, this is a no-op.
func (w *BufferWriter) SetRetryTime(time time.Duration) error {
	return w.Writer.SetRetryTime(time)
}

func (w *BufferWriter) loop() {
	var timer *time.Timer
	var timerChan chan time.Time

	if w.Timeout > 0 {
		timer = time.NewTimer(w.Timeout)
	}

	for {
		select {
		case event, ok := <-w.eventChan:
			if !ok {
				return
			}
			if err := w.Writer.Send(event); err != nil {
				return
			}

		case <-timerChan:
			return
		}

		// If we received an event, reset the timer. See documentation for Reset for
		// why this is necessary.
		if timer != nil {
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(w.Timeout)
		}
	}
}
