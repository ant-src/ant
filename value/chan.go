package value

import (
	"context"
	"errors"
	"sync"
)

// ErrClosed is the error of a send on a closed channel, or of a second close.
var ErrClosed = errors.New("the channel is closed")

// Chan is a Chan[T] of section 9: a Go channel of values beside a closed signal, a second
// channel closed under the mutex. The channel of values is never closed, so a sender blocked in
// it is ended by the signal, not by the runtime's panic for a send on a closed channel, which
// the race detector would report as a race; a receiver that sees the signal drains what is
// buffered before it answers that the channel is closed. The orderings of section 9 are the Go
// channels' own.
type Chan struct {
	ch     chan Value
	closed chan struct{}
	mu     sync.Mutex
	done   bool
}

// NewChan creates a channel; a capacity of zero is a rendezvous.
func NewChan(capacity int) *Chan {
	return &Chan{ch: make(chan Value, capacity), closed: make(chan struct{})}
}

// SyncName returns "Chan".
func (c *Chan) SyncName() string { return "Chan" }

// Send delivers v to a receiver or buffers it, unless ctx is done first, which reports false;
// a context done already gives up at once. It is ErrClosed on a closed channel, including one
// closed while the send waited.
func (c *Chan) Send(ctx context.Context, v Value) (sent bool, err error) {
	if c.IsClosed() {
		return false, ErrClosed
	}
	if ctx.Err() != nil {
		return false, nil
	}
	select {
	case c.ch <- v:
		return true, nil
	case <-c.closed:
		return false, ErrClosed
	case <-ctx.Done():
		return false, nil
	}
}

// Receive takes the next value. ok is false when the channel is closed and drained, and gaveUp
// is true when ctx ended first, or was done already.
func (c *Chan) Receive(ctx context.Context) (v Value, ok, gaveUp bool) {
	if ctx.Err() != nil {
		return nil, false, true
	}
	select {
	case v := <-c.ch:
		return v, true, false
	case <-c.closed:
		select {
		case v := <-c.ch:
			return v, true, false
		default:
			return nil, false, false
		}
	case <-ctx.Done():
		return nil, false, true
	}
}

// Close closes the channel; ErrClosed when it is closed already.
func (c *Chan) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.done {
		return ErrClosed
	}
	c.done = true
	close(c.closed)
	return nil
}

// IsClosed reports whether Close has been called.
func (c *Chan) IsClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

// Len is the number of values buffered now.
func (c *Chan) Len() int { return len(c.ch) }

// Raw is the Go channel of values, for a select over several; it is never closed.
func (c *Chan) Raw() chan Value { return c.ch }

// Closed is the signal, closed once Close has been called, for a select over several.
func (c *Chan) Closed() <-chan struct{} { return c.closed }
