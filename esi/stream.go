package esi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
)

// ErrClosed is returned when writing to a closed Stream.
var ErrClosed = errors.New("esi: write to closed stream")

// result is the output of a tag processed concurrently.
type result struct {
	done chan struct{}
	data []byte
}

// segment is a piece of output: either ready bytes or a pending result.
type segment struct {
	data  []byte
	async *result
}

// queue is an unbounded FIFO between the producer (Write) and the output
// goroutine. Being unbounded, the producer never blocks so every include is
// started as soon as it is parsed. Batches are swapped between two slices so
// the steady state does not allocate.
type queue struct {
	mu     sync.Mutex
	items  []segment
	closed bool
	notify chan struct{}
}

func (q *queue) push(s segment) {
	q.mu.Lock()
	q.items = append(q.items, s)
	q.mu.Unlock()
	q.wake()
}

func (q *queue) close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.wake()
}

func (q *queue) wake() {
	select {
	case q.notify <- struct{}{}:
	default:
	}
}

// take returns every queued segment, reusing buf as the next backing array of
// the queue. It calls idle before blocking and returns false once the queue
// is closed and drained.
func (q *queue) take(buf []segment, idle func()) ([]segment, bool) {
	for {
		q.mu.Lock()
		if len(q.items) > 0 {
			batch := q.items
			q.items = buf[:0]
			q.mu.Unlock()

			return batch, true
		}

		closed := q.closed
		q.mu.Unlock()

		if closed {
			return buf, false
		}

		idle()
		<-q.notify
	}
}

// Stream processes ESI tags on the fly. Data written to it is scanned
// incrementally: static content is forwarded as soon as possible, include
// tags are fetched concurrently and their output is written in document
// order. Close must be called once the input is complete; it waits for the
// pending tags and returns the first error encountered while writing.
//
// Write, Flush and Close must not be called concurrently.
type Stream struct {
	dst    io.Writer
	req    *http.Request
	cancel context.CancelFunc

	// buf holds the unprocessed tail of the input (an incomplete tag).
	buf []byte
	// pending accumulates static output not yet handed to the queue.
	pending []byte

	// q is nil until the first concurrent tag. Until then output is written
	// synchronously by the caller, without copying.
	q    *queue
	done chan struct{}

	dirty  bool // written since the last flush; owned by the writing goroutine.
	closed bool

	// err is the first write error. It is guarded by a mutex rather than an
	// atomic.Pointer because Traefik's Yaegi interpreter cannot handle generic
	// standard library types.
	errMu sync.Mutex
	err   error
}

// NewStream returns a Stream writing the processed document to dst. The
// context of req bounds the lifetime of the include requests. If dst
// implements http.Flusher (or Flush() error), it is flushed whenever the
// output has to wait for a tag or for more input.
func NewStream(dst io.Writer, req *http.Request) *Stream {
	ctx, cancel := context.WithCancel(req.Context())

	return &Stream{
		dst:    dst,
		req:    req.WithContext(ctx),
		cancel: cancel,
	}
}

// Write processes p. It never blocks on include requests.
func (s *Stream) Write(p []byte) (int, error) {
	if s.closed {
		return 0, ErrClosed
	}

	if err := s.loadErr(); err != nil {
		return 0, err
	}

	data := p
	if len(s.buf) > 0 {
		s.buf = append(s.buf, p...)
		data = s.buf
	}

	n := scan(data, false, s)
	// Keep the unprocessed tail. copy handles the overlap when data is s.buf.
	s.buf = append(s.buf[:0], data[n:]...)
	s.flushPending()

	return len(p), s.loadErr()
}

// Flush forwards the processed output to the destination as soon as possible.
func (s *Stream) Flush() {
	if s.closed {
		return
	}

	if s.q == nil {
		s.flush()

		return
	}

	// The output goroutine flushes when it becomes idle.
	s.flushPending()
}

// Close processes the remaining input, waits for the pending tags and
// releases the resources. It is safe to call it more than once.
func (s *Stream) Close() error {
	if s.closed {
		return s.loadErr()
	}

	s.closed = true

	if len(s.buf) > 0 {
		scan(s.buf, true, s)
		s.buf = nil
	}

	if s.q != nil {
		s.flushPending()
		s.q.close()
		<-s.done
	}

	s.cancel()

	return s.loadErr()
}

func (s *Stream) static(b []byte) {
	if len(b) == 0 {
		return
	}

	if s.q == nil {
		s.write(b)

		return
	}

	s.pending = append(s.pending, b...)
}

func (s *Stream) tag(t Tag, body []byte) {
	if !isAsync(t, body) {
		s.static(process(t, body, s.req))

		return
	}

	if s.q == nil {
		s.q = &queue{notify: make(chan struct{}, 1)}
		s.done = make(chan struct{})

		go s.run()
	}

	s.flushPending()

	r := &result{done: make(chan struct{})}
	body = bytes.Clone(body)

	go func() {
		defer close(r.done)

		r.data = process(t, body, s.req)
	}()

	s.q.push(segment{async: r})
}

func (s *Stream) flushPending() {
	if s.q != nil && len(s.pending) > 0 {
		s.q.push(segment{data: s.pending})
		s.pending = nil
	}
}

// run writes the queued segments in order until the queue is closed.
func (s *Stream) run() {
	defer close(s.done)

	var batch []segment

	for {
		var ok bool
		if batch, ok = s.q.take(batch, s.flush); !ok {
			return
		}

		for i := range batch {
			seg := &batch[i]
			if r := seg.async; r != nil {
				select {
				case <-r.done:
				default:
					s.flush()
					<-r.done
				}

				seg.data = r.data
			}

			s.write(seg.data)
		}

		// Release the references (the clear builtin is unknown to Yaegi).
		for i := range batch {
			batch[i] = segment{}
		}
	}
}

func (s *Stream) write(b []byte) {
	if len(b) == 0 || s.loadErr() != nil {
		return
	}

	if _, err := s.dst.Write(b); err != nil {
		s.setErr(err)
		// Nobody will read the output anymore: abort the pending includes.
		s.cancel()

		return
	}

	s.dirty = true
}

func (s *Stream) flush() {
	if !s.dirty || s.loadErr() != nil {
		return
	}

	s.dirty = false

	switch f := s.dst.(type) {
	case http.Flusher:
		f.Flush()
	case interface{ Flush() error }:
		if err := f.Flush(); err != nil {
			s.setErr(err)
			s.cancel()
		}
	}
}

func (s *Stream) loadErr() error {
	s.errMu.Lock()
	defer s.errMu.Unlock()

	return s.err
}

// setErr records err unless an error was already recorded.
func (s *Stream) setErr(err error) {
	s.errMu.Lock()
	defer s.errMu.Unlock()

	if s.err == nil {
		s.err = err
	}
}

var _ io.WriteCloser = (*Stream)(nil)
