package value

import (
	"context"
	"errors"
	"sync"
)

// ErrClosed is the error of a send on a closed channel, or of a second close.
var ErrClosed = errors.New("the channel is closed")

// Chan is a Chan[T] of section 9: a Go channel of values with a closed flag, so that a send on
// a closed channel and a second close are errors for the caller to turn into panics. The
// orderings of section 9 are the Go channel's own.
type Chan struct {
	ch     chan Value
	mu     sync.Mutex
	closed bool
}

// NewChan creates a channel; a capacity of zero is a rendezvous.
func NewChan(capacity int) *Chan { return &Chan{ch: make(chan Value, capacity)} }

// SyncName returns "Chan".
func (c *Chan) SyncName() string { return "Chan" }

// Send delivers v to a receiver or buffers it, unless ctx is done first, which reports false.
// It is ErrClosed on a closed channel, including one closed while the send waited.
func (c *Chan) Send(ctx context.Context, v Value) (sent bool, err error) {
	if c.IsClosed() {
		return false, ErrClosed
	}
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok && isClosedPanic(e) {
				sent, err = false, ErrClosed
				return
			}
			panic(r)
		}
	}()
	select {
	case c.ch <- v:
		return true, nil
	case <-ctx.Done():
		return false, nil
	}
}

// Receive takes the next value. ok is false when the channel is closed and drained, and gaveUp
// is true when ctx ended first.
func (c *Chan) Receive(ctx context.Context) (v Value, ok, gaveUp bool) {
	select {
	case v, ok := <-c.ch:
		return v, ok, false
	case <-ctx.Done():
		return nil, false, true
	}
}

// Close closes the channel; ErrClosed when it is closed already.
func (c *Chan) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrClosed
	}
	c.closed = true
	close(c.ch)
	return nil
}

// IsClosed reports whether Close has been called.
func (c *Chan) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// Len is the number of values buffered now.
func (c *Chan) Len() int { return len(c.ch) }

// Raw is the Go channel, for a select over several.
func (c *Chan) Raw() chan Value { return c.ch }

// isClosedPanic recognizes the runtime's panic for a send on a closed channel.
func isClosedPanic(e error) bool {
	msg := e.Error()
	return msg == "send on closed channel" || msg == "close of closed channel"
}
