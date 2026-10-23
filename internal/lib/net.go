package lib

import (
	"context"
	"net"
	"sync"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// The Net namespace (stdlib.md): TCP listeners and connections, a connection being a
// File.Handle whose writes flush and whose reads a context can end through the deadline.

// Listener is a listening TCP socket: Net.Listener. Connections are accepted by a pump, one
// goroutine started at the first Net.accept, which hands each over conns to the next caller,
// so a caller selects on its own context and disturbs no other; a connection accepted while no
// caller waits is held for the next one.
type Listener struct {
	mu     sync.Mutex
	l      *net.TCPListener
	closed bool
	err    error // what ended the pump, read once conns is closed
	once   sync.Once
	conns  chan net.Conn
}

// errListenerClosed is the message of the Err a closed listener gives back.
const errListenerClosed = "the listener is closed"

// pump accepts until the listener fails, which a close causes, then records the error and
// closes conns; a connection it holds when the run ends is closed.
func (l *Listener) pump(done <-chan struct{}) {
	for {
		c, err := l.l.Accept()
		if err != nil {
			l.mu.Lock()
			l.err = err
			l.mu.Unlock()
			close(l.conns)
			return
		}
		select {
		case l.conns <- c:
		case <-done:
			c.Close()
			return
		}
	}
}

// serveListener is the listener as Http.serve's server sees it: connections come from the
// pump, so a server on a listener that Net.accept has used does not race the pump for them.
// Close ends a pending Accept even when the pump has already stopped.
type serveListener struct {
	l      *Listener
	done   <-chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func newServeListener(l *Listener, done <-chan struct{}) serveListener {
	ctx, cancel := context.WithCancel(context.Background())
	return serveListener{l, done, ctx, cancel}
}

func (s serveListener) Accept() (net.Conn, error) { return s.l.accept(s.done, s.ctx) }

func (s serveListener) Close() error {
	s.cancel()
	s.l.markClosed()
	return s.l.l.Close()
}

func (s serveListener) Addr() net.Addr { return s.l.l.Addr() }

func (l *Listener) markClosed() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true
}

// SyncName names the listener for rendering.
func (l *Listener) SyncName() string { return "Net.Listener" }

func listenerArg(g *interp.Goroutine, who string, v value.Value) *Listener {
	return typedArg[*Listener](g, who, "a Net.Listener", v)
}

// connHandle wraps a connection as a handle: buffered both ways, writes flushed, reads ended
// by a context through the deadline.
func connHandle(c net.Conn) *Handle {
	h := newHandle(c, c, c)
	h.dl, h.conn, h.peer = c, true, c.RemoteAddr().String()
	return h
}

// accept waits for the next connection from the pump, or gives up when ctx is done.
func (l *Listener) accept(done <-chan struct{}, ctx context.Context) (net.Conn, error) {
	l.once.Do(func() {
		l.conns = make(chan net.Conn)
		go l.pump(done)
	})
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	select {
	case c, ok := <-l.conns:
		if !ok {
			l.mu.Lock()
			defer l.mu.Unlock()
			return nil, l.err
		}
		if l.isClosed() {
			// A connection the pump held from before the close is not handed out after it.
			c.Close()
			return nil, net.ErrClosed
		}
		return c, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func init() {
	register("Net", func() value.Value {
		return namespace("Net",
			constant("Listener", value.NewTag("Net.Listener", value.SyncTag, nil)),
			fn("Net.listen", "addr", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				l, err := net.Listen("tcp", stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(&Listener{l: l.(*net.TCPListener)})
			}),
			fn("Net.address", "l", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return listenerArg(g, who, a[0]).l.Addr().String()
			}),
			fn("Net.accept", "l, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				l := listenerArg(g, who, a[0])
				ctx := waitContext(g, who, a[1])
				g.Instance().Flush()
				c, err := l.accept(g.Context().Done(), ctx)
				g.AfterWait()
				if err != nil {
					if ctx.Err() != nil {
						return gaveUp(ctx)
					}
					if l.isClosed() {
						return value.Err(value.ErrorValue(errListenerClosed))
					}
					return errResult(err)
				}
				return value.Ok(connHandle(c))
			}),
			fn("Net.close", "l", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				l := listenerArg(g, who, a[0])
				l.mu.Lock()
				defer l.mu.Unlock()
				if !l.closed {
					l.closed = true
					l.l.Close()
				}
				return value.Void
			}),
			fn("Net.dial", "addr, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				ctx := waitContext(g, who, a[1])
				g.Instance().Flush()
				var d net.Dialer
				c, err := d.DialContext(ctx, "tcp", stringArg(g, who, a[0]))
				g.AfterWait()
				if err != nil {
					return failed(ctx, err)
				}
				return value.Ok(connHandle(c))
			}),
			fn("Net.peer", "f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0])
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
