package value

import (
	"context"
	"errors"
	"sync"
)

// ErrClosed is the error of a send on a closed channel, or of a second close.
var ErrClosed = errors.New("the channel is closed")

// ErrTicker is the error of a send on a ticker's channel, or of closing one: only the ticker
// sends, and the channel is never closed, as Go's ticker channel is receive-only.
var ErrTicker = errors.New("a ticker's channel only receives")

// Chan is a Channel[T] of section 9: a Go channel of values beside a closed signal, a second
// channel closed under the mutex. The channel of values is never closed, so a sender blocked in
// it is ended by the signal, not by the runtime's panic for a send on a closed channel, which
// the race detector would report as a race; a receiver that sees the signal drains what is
// buffered before it answers that the channel is closed. The orderings of section 9 are the Go
// channels' own.
//
// A ticker's channel carries the ticker's context, and a tick taken once that context is done is
// discarded, as Go's Ticker.Stop discards a stale one: a receive that begins after the context
// ends gets no tick, whichever way it ended, since the ticker's own goroutine can always have
// one in flight.
type Chan struct {
	ch     chan Value
	closed chan struct{}
	mu     sync.Mutex
	done   bool
	ticker context.Context // nil but for a ticker's channel
}

// NewChan creates a channel; a capacity of zero is a rendezvous.
func NewChan(capacity int) *Chan {
	return &Chan{ch: make(chan Value, capacity), closed: make(chan struct{})}
}

// NewTickerChan creates a ticker's channel, holding one tick, that stops with ctx.
func NewTickerChan(ctx context.Context) *Chan {
	c := NewChan(1)
	c.ticker = ctx
	return c
}

// IsTicker reports whether the channel is a ticker's.
func (c *Chan) IsTicker() bool { return c.ticker != nil }

// Stopped reports whether the channel is a ticker's whose context is done, so a tick taken from
// it is stale.
func (c *Chan) Stopped() bool { return c.ticker != nil && c.ticker.Err() != nil }

// SyncName returns "Channel".
func (c *Chan) SyncName() string { return "Channel" }

// Send delivers v to a receiver or buffers it, unless ctx is done first, which reports false;
// a context done already gives up at once. It is ErrClosed on a closed channel, including one
// closed while the send waited.
func (c *Chan) Send(ctx context.Context, v Value) (sent bool, err error) {
	if c.IsTicker() {
		return false, ErrTicker
	}
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
// is true when ctx ended first, or was done already. A stale tick is discarded and the wait goes
// on.
func (c *Chan) Receive(ctx context.Context) (v Value, ok, gaveUp bool) {
	if ctx.Err() != nil {
		return nil, false, true
	}
	for c.ticker != nil {
		select {
		case v := <-c.ch:
			if !c.Stopped() {
				return v, true, false
			}
		case <-ctx.Done():
			return nil, false, true
		}
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
	if c.IsTicker() {
		return ErrTicker
	}
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

// Len is the number of values buffered now; none for a stopped ticker, whose tick is stale.
func (c *Chan) Len() int {
	if c.Stopped() {
		return 0
	}
	return len(c.ch)
}

// Raw is the Go channel of values, for a select over several; it is never closed.
func (c *Chan) Raw() chan Value { return c.ch }

// Closed is the signal, closed once Close has been called, for a select over several.
func (c *Chan) Closed() <-chan struct{} { return c.closed }
