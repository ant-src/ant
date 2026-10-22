package lib

import (
	"bufio"
	"context"
	"net"
	"sync"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// The Net namespace (stdlib.md): TCP listeners and connections, a connection being a
// File.Handle whose writes flush and whose reads a context can end through the deadline.

// Listener is a listening TCP socket: Net.Listener.
type Listener struct {
	mu     sync.Mutex
	l      *net.TCPListener
	closed bool
}

// SyncName names the listener for rendering.
func (l *Listener) SyncName() string { return "Net.Listener" }

func listenerArg(g *interp.Goroutine, who string, v value.Value) *Listener {
	l, ok := v.(*Listener)
	if !ok {
		g.Panic("%s needs a Net.Listener, got %s", who, value.Describe(v))
	}
	return l
}

// connHandle wraps a connection as a handle: buffered both ways, writes flushed, reads ended
// by a context through the deadline.
func connHandle(c net.Conn) *Handle {
	return &Handle{c: c, dl: c, r: bufio.NewReader(c), w: bufio.NewWriter(c), conn: true, peer: c.RemoteAddr().String()}
}

// accept waits for the next connection, or gives up when ctx is done by setting the listener's
// deadline to now, as a read does.
func (l *Listener) accept(ctx context.Context) (net.Conn, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	stop := make(chan struct{})
	watcher := make(chan struct{})
	go func() {
		defer close(watcher)
		select {
		case <-ctx.Done():
			l.l.SetDeadline(time.Now())
		case <-stop:
		}
	}()
	c, err := l.l.Accept()
	close(stop)
	<-watcher
	if ctx.Err() != nil {
		l.l.SetDeadline(time.Time{})
		if err != nil {
			return nil, ctx.Err()
		}
	}
	return c, err
}

func init() {
	register("Net", func() value.Value {
		return namespace("Net",
			constant("Listener", value.NewTag("Net.Listener", value.SyncTag, nil)),
			fn("Net.listen", "addr", func(g *interp.Goroutine, a []value.Value) value.Value {
				l, err := net.Listen("tcp", stringArg(g, "Net.listen", a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(&Listener{l: l.(*net.TCPListener)})
			}),
			fn("Net.address", "l", func(g *interp.Goroutine, a []value.Value) value.Value {
				return listenerArg(g, "Net.address", a[0]).l.Addr().String()
			}),
			fn("Net.accept", "l, ctx?", func(g *interp.Goroutine, a []value.Value) value.Value {
				l := listenerArg(g, "Net.accept", a[0])
				ctx := waitContext(g, optionalContext(g, "Net.accept", a[1]))
				g.Instance().Flush()
				c, err := l.accept(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return gaveUp(ctx)
					}
					if l.isClosed() {
						return value.Err(value.ErrorValue("the listener is closed"))
					}
					return errResult(err)
				}
				return value.Ok(connHandle(c))
			}),
			fn("Net.close", "l", func(g *interp.Goroutine, a []value.Value) value.Value {
				l := listenerArg(g, "Net.close", a[0])
				l.mu.Lock()
				defer l.mu.Unlock()
				if !l.closed {
					l.closed = true
					l.l.Close()
				}
				return value.Void
			}),
			fn("Net.dial", "addr, ctx?", func(g *interp.Goroutine, a []value.Value) value.Value {
				ctx := waitContext(g, optionalContext(g, "Net.dial", a[1]))
				g.Instance().Flush()
				var d net.Dialer
				c, err := d.DialContext(ctx, "tcp", stringArg(g, "Net.dial", a[0]))
				if err != nil {
					if ctx.Err() != nil {
						return gaveUp(ctx)
					}
					return errResult(err)
				}
				return value.Ok(connHandle(c))
			}),
			fn("Net.peer", "f", func(g *interp.Goroutine, a []value.Value) value.Value {
				h := handleArg(g, "Net.peer", a[0])
				if !h.conn {
					return value.Null
				}
				return h.peer
			}),
		)
	})
}

func (l *Listener) isClosed() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.closed
}
