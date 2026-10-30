package lib

import (
	"io"
	"os"
	"sync"
)

// pump reads standard input on a goroutine of its own, since no deadline and no close ends a
// terminal's read: the goroutine blocks in the real read and hands chunks over a channel, and
// Read takes them, or gives up when the current reader's context is done without taking one, so
// a read the context ended loses nothing and the next read gets the chunk. The goroutine reads
// only when asked, so between the program's reads the input is left to a child that is handed
// it. It starts at the first read and ends with the input, or stays in its read when the run
// ends first.
type pump struct {
	src    io.Reader
	done   <-chan struct{} // the run's end, after which a chunk has no taker
	asks   chan struct{}
	chunks chan chunk
	once   sync.Once
	asked  bool            // a read was asked for and its chunk not yet taken
	rest   []byte          // the unread part of the last chunk
	err    error           // what ended the input, given once rest is empty
	cancel <-chan struct{} // the current reader's context, set under the read slot
}

type chunk struct {
	data []byte
	err  error
}

func (p *pump) run() {
	for {
		select {
		case <-p.asks:
		case <-p.done:
			return
		}
		buf := make([]byte, 4096)
		n, err := p.src.Read(buf)
		select {
		case p.chunks <- chunk{buf[:n], err}:
		case <-p.done:
			return
		}
		if err != nil {
			return
		}
	}
}

func (p *pump) Read(b []byte) (int, error) {
	p.once.Do(func() {
		p.asks = make(chan struct{}, 1)
		p.chunks = make(chan chunk)
		go p.run()
	})
	if len(p.rest) == 0 && p.err == nil {
		if !p.asked {
			p.asks <- struct{}{}
			p.asked = true
		}
		select {
		case c := <-p.chunks:
			p.rest, p.err, p.asked = c.data, c.err, false
		case <-p.cancel:
			return 0, os.ErrDeadlineExceeded
		}
	}
	n := copy(b, p.rest)
	p.rest = p.rest[n:]
	if n == 0 {
		return 0, p.err
	}
	return n, nil
}

// viaReader is a pipe's reading end as its handle reads it, so the source can become a pump,
// under the read slot, once the end is handed to a child.
type viaReader struct{ r io.Reader }

func (v *viaReader) Read(b []byte) (int, error) { return v.r.Read(b) }
