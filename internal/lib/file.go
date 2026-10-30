package lib

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/temporal"
	"github.com/ant-src/ant/value"
)

// InfoTag is File.Info: { name: string, size: int, isDir: bool, modified: Time.Instant }.
var InfoTag = value.NewTag("File.Info", value.RecordTag, []string{"name", "size", "isDir", "modified"})

// The File namespace (stdlib.md): every operation that can fail returns a Result with Error.

// deadliner is what a connection has and a regular file lacks: a read deadline, which a context
// sets to now to end a read that waits.
type deadliner interface {
	SetReadDeadline(t time.Time) error
}

// Handle is an open file, a standard stream, or a connection: File.Handle. A reader holds the
// read slot, a writer the write mutex, and close takes the state mutex alone, so one goroutine
// may read a connection while another writes it, and a close from a third ends both.
type Handle struct {
	state   sync.Mutex
	closed  bool
	c       io.Closer        // the file or connection, closed by File.close; nil for a standard stream
	reads   chan struct{}    // the read slot, taken with the caller's context; nil when the handle does not read
	r       *bufio.Reader    // the input; a read consumes it into pending
	pending string           // what a read consumed before its context ended it, for the next read
	scratch []byte           // what File.readChunk reads into when nothing is buffered; held by the read slot
	dl      deadliner        // a connection's read deadline; nil for a file, whose reads never wait
	pump    *pump            // standard input, read on a goroutine so a terminal cannot hold a cancelled run
	in      *interp.Instance // standard input: a read flushes stdout first
	writes  sync.Mutex
	w       *bufio.Writer
	std     bool       // a standard stream: close does nothing, a write flushes
	conn    bool       // a connection: a write flushes
	pipe    bool       // a pipe's writing end: a write flushes, since the other end waits for it
	via     *viaReader // a pipe's reading end: its source, which handing it to a child changes
	peer    string     // a connection's remote address, for Net.peer
	// What a child process is given for a standard stream: the run's own, not the handle's
	// buffer (process.go).
	childIn  io.Reader
	childOut io.Writer
}

// newHandle makes a handle over c that reads r and writes w, either of which may be nil.
func newHandle(c io.Closer, r io.Reader, w io.Writer) *Handle {
	h := &Handle{c: c}
	if r != nil {
		h.r = bufio.NewReader(r)
		h.reads = make(chan struct{}, 1)
	}
	if w != nil {
		h.w = bufio.NewWriter(w)
	}
	return h
}

func (h *Handle) isClosed() bool {
	h.state.Lock()
	defer h.state.Unlock()
	return h.closed
}

// The messages of the Errs a handle gives back. Each Err gets its own Error record, since
// records are mutable and a program may change the one it receives.
const (
	errNotReadable = "the file is not open for reading"
	errNotWritable = "the file is not open for writing"
)

// read runs f, which reads from h.r into h.pending and answers from there, holding the read
// slot, and gives up with the context's reason when ctx is done, whether while waiting for the
// slot behind another reader or during the read: a connection has its read deadline set to now,
// the pump of standard input stops handing out chunks, and a regular file never waits, so there
// the context is looked at only before the read begins. What f consumed before giving up stays
// in h.pending for the next read, so nothing is lost and the handle stays open.
func (h *Handle) read(ctx context.Context, f func() value.Value) value.Value {
	if h.r == nil {
		return value.Err(value.ErrorValue(errNotReadable))
	}
	select {
	case h.reads <- struct{}{}:
	case <-ctx.Done():
		return gaveUp(ctx)
	}
	defer func() { <-h.reads }()
	if h.isClosed() {
		return value.Err(value.ErrorValue(errNotReadable))
	}
	if ctx.Err() != nil {
		return gaveUp(ctx)
	}
	if h.in != nil {
		h.in.Flush()
	}
	switch {
	case h.pump != nil:
		h.pump.cancel = ctx.Done()
		v := f()
		h.pump.cancel = nil
		return h.ended(ctx, v)
	case h.dl != nil:
		fired := make(chan struct{})
		stop := context.AfterFunc(ctx, func() {
			h.dl.SetReadDeadline(time.Now())
			close(fired)
		})
		v := f()
		if !stop() {
			<-fired // set before it is cleared
			h.dl.SetReadDeadline(time.Time{})
		}
		return h.ended(ctx, v)
	}
	return f()
}

// ended gives a read that failed under a done context the context's reason.
func (h *Handle) ended(ctx context.Context, v value.Value) value.Value {
	if ctx.Err() != nil {
		if r, ok := v.(*value.Record); ok && r.Tag == value.ErrTag {
			return gaveUp(ctx)
		}
	}
	return v
}

// readLine is File.readLine's read: a line from pending when it holds one, else more of the
// input. At the end of the input the rest is the last line, or null when there is nothing.
func (h *Handle) readLine() value.Value {
	for {
		if i := strings.IndexByte(h.pending, '\n'); i >= 0 {
			line := strings.TrimSuffix(h.pending[:i], "\r")
			h.pending = h.pending[i+1:]
			return value.Ok(value.Text(line))
		}
		s, err := h.r.ReadString('\n')
		h.pending += s
		if err == nil {
			continue
		}
		if err == io.EOF {
			if h.pending == "" {
				return value.Ok(value.Null)
			}
			line := h.pending
			h.pending = ""
			return value.Ok(value.Text(line))
		}
		return errResult(err)
	}
}

// readAll is File.readAll's read: pending and the rest of the input.
func (h *Handle) readAll() value.Value {
	data, err := io.ReadAll(h.r)
	h.pending += string(data)
	if err != nil {
		return errResult(err)
	}
	text := h.pending
	h.pending = ""
	return value.Ok(value.Text(text))
}

// readChunk is File.readChunk's read: what pending holds, else what the input has buffered, else
// one read of the input into the handle's scratch buffer, 1 to max bytes each way; null at the
// end of the input. A chunk is allocated at the size it has, so a socket read with a large max
// that gets a short message allocates the message, not max.
func (h *Handle) readChunk(max int) value.Value {
	if h.pending != "" {
		n := min(max, len(h.pending))
		out := h.pending[:n]
		h.pending = h.pending[n:]
		return value.Ok(value.Bytes(out))
	}
	if n := h.r.Buffered(); n > 0 {
		buf := make([]byte, min(max, n))
		h.r.Read(buf) // a buffered read returns all it is asked for, without an error
		return value.Ok(bytesValue(buf))
	}
	if cap(h.scratch) < min(max, scratchCap) {
		h.scratch = make([]byte, min(max, scratchCap))
	}
	for {
		n, err := h.r.Read(h.scratch[:min(max, cap(h.scratch))])
		switch {
		case n > 0:
			return value.Ok(value.Bytes(h.scratch[:n]))
		case err == io.EOF:
			return value.Ok(value.Null)
		case err != nil:
			return errResult(err)
		}
	}
}

// scratchCap bounds the buffer a File.readChunk reads into when nothing is buffered, so one read
// returns at most that much, whatever max it is given.
const scratchCap = 64 << 10

// readExactly is File.readExactly's read: n bytes, gathered as they arrive. The buffer grows with
// what arrives rather than with n, which may come from the other end; a read that the context
// or an error ends early leaves what it gathered in pending for the next read. null when the input
// is already at its end, and an Err when it ends partway, which discards the partial bytes.
func (h *Handle) readExactly(n int) value.Value {
	if len(h.pending) >= n {
		out := h.pending[:n]
		h.pending = h.pending[n:]
		return value.Ok(value.Bytes(out))
	}
	buf := []byte(h.pending)
	h.pending = ""
	for len(buf) < n {
		if len(buf) == cap(buf) {
			buf = slices.Grow(buf, min(n-len(buf), max(len(buf), 4096)))
		}
		k, err := h.r.Read(buf[len(buf):min(n, cap(buf))])
		buf = buf[:len(buf)+k]
		if len(buf) == n {
			break // a reader may return the last bytes with io.EOF; they are all here
		}
		if err == io.EOF {
			if len(buf) == 0 {
				return value.Ok(value.Null)
			}
			return value.Err(value.ErrorValue("unexpected end of file"))
		}
		if err != nil {
			h.pending = string(buf)
			return errResult(err)
		}
	}
	return value.Ok(bytesValue(buf))
}

// write writes s under the write mutex and flushes a stream or a connection.
func (h *Handle) write(s string) value.Value {
	if h.w == nil {
		return value.Err(value.ErrorValue(errNotWritable))
	}
	h.writes.Lock()
	defer h.writes.Unlock()
	if h.isClosed() {
		return value.Err(value.ErrorValue(errNotWritable))
	}
	_, err := h.w.WriteString(s)
	return h.flushed(err)
}

// flushed flushes a standard stream or a connection after a write, whatever the write's error,
// so that what a program writes to stdout interleaves with print in program order, what it
// writes to stderr is not lost at exit, what it writes to a connection reaches the other end,
// which is waiting, and what a pipe copied before its context ended it is delivered.
func (h *Handle) flushed(err error) value.Value {
	if h.std || h.conn || h.pipe {
		if ferr := h.w.Flush(); err == nil {
			err = ferr
		}
	}
	return okVoid(err)
}

// close flushes and closes h, returning the first error of the two, so a write that failed only
// when its buffer was written out is reported. A standard stream and a closed handle give nil.
func (h *Handle) close() error {
	if h.std {
		if h.w != nil {
			h.writes.Lock()
			h.w.Flush()
			h.writes.Unlock()
		}
		return nil
	}
	h.state.Lock()
	if h.closed {
		h.state.Unlock()
		return nil
	}
	h.closed = true
	h.state.Unlock()
	// A file's writes never block, so its buffer is flushed before the close; a connection's are
	// flushed as they are written, and its close ends a write or a read in progress on another
	// goroutine.
	var err error
	if h.w != nil && !h.conn {
		h.writes.Lock()
		err = h.w.Flush()
		h.writes.Unlock()
	}
	if cerr := h.c.Close(); err == nil {
		err = cerr
	}
	return err
}

// SyncName names the handle for rendering.
func (h *Handle) SyncName() string { return "File.Handle" }

func handleArg(g *interp.Goroutine, who string, v value.Value) *Handle {
	return typedArg[*Handle](g, who, "a File.Handle", v)
}

func init() {
	register("File", func() value.Value {
		return namespace("File",
			constant("Handle", value.NewTag("File.Handle", value.SyncTag, nil)),
			constant("Info", InfoTag),
			fn("File.stat", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				fi, err := os.Stat(stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				modified, err := temporal.NewInstant(fi.ModTime().Unix(), int64(fi.ModTime().Nanosecond()))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(value.NewRecord(InfoTag, []value.Value{value.Text(fi.Name()), fi.Size(), fi.IsDir(), modified}))
			}),
			fn("File.mkdir", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(os.MkdirAll(stringArg(g, who, a[0]), 0o755))
			}),
			fn("File.removeAll", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(os.RemoveAll(stringArg(g, who, a[0])))
			}),
			fn("File.rename", "from, to", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(os.Rename(stringArg(g, who, a[0]), stringArg(g, who, a[1])))
			}),
			fn("File.tempDir", "prefix", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				dir, err := os.MkdirTemp("", stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(value.Text(dir))
			}),
			fn("File.walk", "dir", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				dir := stringArg(g, who, a[0])
				if fi, err := os.Stat(dir); err != nil {
					return errResult(err)
				} else if !fi.IsDir() {
					return errResult(errors.New(dir + " is not a directory"))
				}
				// The links below dir are not followed, but dir itself is: walked as given, a link
				// to a directory would be reported as one file, ".".
				root, err := filepath.EvalSymlinks(dir)
				if err != nil {
					return errResult(err)
				}
				var files []value.Value
				err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}
					rel, err := filepath.Rel(root, p)
					if err != nil {
						return err
					}
					files = append(files, value.Text(rel))
					return nil
				})
				if err != nil {
					return errResult(err)
				}
				return value.Ok(value.NewArray(files))
			}),
			fn("File.append", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				f, err := os.OpenFile(stringArg(g, who, a[0]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				if err != nil {
					return errResult(err)
				}
				return value.Ok(newHandle(f, nil, f))
			}),
			fn("File.read", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				data, err := os.ReadFile(stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(value.Text(string(data)))
			}),
			fn("File.readBytes", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				data, err := os.ReadFile(stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(bytesValue(data))
			}),
			fn("File.write", "path, text", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(os.WriteFile(stringArg(g, who, a[0]), []byte(stringArg(g, who, a[1])), 0o644))
			}),
			fn("File.writeBytes", "path, bytes", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(os.WriteFile(stringArg(g, who, a[0]), view(bytesArg(g, who, a[1])), 0o644))
			}),
			fn("File.exists", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				_, err := os.Stat(stringArg(g, who, a[0]))
				return err == nil
			}),
			fn("File.remove", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(os.Remove(stringArg(g, who, a[0])))
			}),
			fn("File.list", "dir", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				entries, err := os.ReadDir(stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				names := make([]value.Value, len(entries))
				for i, e := range entries {
					names[i] = value.Text(e.Name())
				}
				return value.Ok(value.NewArray(names))
			}),
			fn("File.open", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				f, err := os.Open(stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(newHandle(f, f, nil))
			}),
			fn("File.create", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				f, err := os.Create(stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(newHandle(f, nil, f))
			}),
			fn("File.readLine", "f, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0])
				ctx := waitContext(g, who, a[1])
				defer g.AfterWait()
				return h.read(ctx, h.readLine)
			}),
			fn("File.readAll", "f, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0])
				ctx := waitContext(g, who, a[1])
				defer g.AfterWait()
				return h.read(ctx, h.readAll)
			}),
			fn("File.readChunk", "f, max, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0])
				max := intArg(g, who, a[1])
				ctx := waitContext(g, who, a[2])
				if max < 1 {
					return value.Err(value.ErrorValue(fmt.Sprintf("the most to read is %d; it must be at least 1", max)))
				}
				defer g.AfterWait()
				return h.read(ctx, func() value.Value { return h.readChunk(int(min(max, scratchCap))) })
			}),
			fn("File.readExactly", "f, n, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0])
				n := intArg(g, who, a[1])
				ctx := waitContext(g, who, a[2])
				if n < 0 {
					return value.Err(value.ErrorValue(fmt.Sprintf("the count to read is %d; it must be at least 0", n)))
				}
				defer g.AfterWait()
				return h.read(ctx, func() value.Value { return h.readExactly(int(min(n, math.MaxInt))) })
			}),
			fn("File.writeChunk", "f, bytes", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0])
				return h.write(string(bytesArg(g, who, a[1])))
			}),
			fn("File.writeLine", "f, line", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0])
				return h.write(stringArg(g, who, a[1]) + "\n")
			}),
			fn("File.writeText", "f, text", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0])
				return h.write(stringArg(g, who, a[1]))
			}),
			fn("File.flush", "f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0])
				h.writes.Lock()
				defer h.writes.Unlock()
				if h.isClosed() {
					return value.Err(value.ErrorValue("the file is closed"))
				}
				if h.w == nil {
					return value.Ok(value.Void)
				}
				if err := h.w.Flush(); err != nil {
					return errResult(err)
				}
				return value.Ok(value.Void)
			}),
			fn("File.pipe", "src, dst, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				src := handleArg(g, who, a[0])
				dst := handleArg(g, who, a[1])
				ctx := waitContext(g, who, a[2])
				if src == dst {
					return value.Err(value.ErrorValue("a handle cannot be piped into itself"))
				}
				if src.r == nil || src.isClosed() {
					return value.Err(value.ErrorValue("the source is not open for reading"))
				}
				if dst.w == nil || dst.isClosed() {
					return value.Err(value.ErrorValue("the destination is not open for writing"))
				}
				// The source's read slot, then the destination's write mutex: a pipe the other
				// way takes the other two, so the two cannot deadlock.
				defer g.AfterWait()
				return src.read(ctx, func() value.Value {
					dst.writes.Lock()
					defer dst.writes.Unlock()
					if dst.isClosed() {
						return value.Err(value.ErrorValue("the destination is not open for writing"))
					}
					_, err := dst.w.WriteString(src.pending)
					src.pending = ""
					if err == nil {
						_, err = io.Copy(dst.w, src.r)
					}
					return dst.flushed(err)
				})
			}),
			fn("File.close", "f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(handleArg(g, who, a[0]).close())
			}),
		)
	})
}
