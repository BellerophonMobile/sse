package sse

import (
	"container/list"
	"io"
	"net/http"
)

// A fanOutWriter duplicates writes to all writers. Writers are removed from the
// set if any errors occur writing to them.
type fanOutWriter struct {
	writers list.List
}

// Add adds a writer to the set, returning a function to remove the writer.
func (f *fanOutWriter) Add(w io.Writer) func() {
	el := f.writers.PushBack(w)
	return func() {
		f.writers.Remove(el)
	}
}

// Write writes len(p) bytes to every writer in the set. len(p) is always
// returned as the number of bytes written, and the first error encountered will
// be returned if any. If an error occurs, that writer is removed from the set,
// and writing continues.
func (f *fanOutWriter) Write(p []byte) (int, error) {
	return len(p), f.forEach(func(w io.Writer) error {
		_, err := w.Write(p)
		return err
	})
}

// Close removes all writers from the set. It also closes all Closers in the
// set. The first error encountered is returned.
func (f *fanOutWriter) Close() error {
	err := f.forEach(func(w io.Writer) error {
		if c, ok := w.(io.Closer); ok {
			return c.Close()
		}
		return nil
	})

	f.writers.Init()

	return err
}

// Flush calls flush on any writers that implement http.Flusher.
func (f *fanOutWriter) Flush() {
	f.forEach(func(w io.Writer) error {
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		return nil
	})
}

func (f *fanOutWriter) forEach(fn func(w io.Writer) error) error {
	var firstErr error

	for el := f.writers.Front(); el != nil; el = el.Next() {
		writer := el.Value.(io.Writer)

		if err := fn(writer); err != nil {
			f.writers.Remove(el)

			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
