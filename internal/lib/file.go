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
	"github.com/ant-src/ant/value"
)

// InfoTag is File.Info: { name: string, size: int, isDir: bool, isSymlink: bool, modified: Time.Instant, mode: int }.
var InfoTag = value.NewTag("File.Info", value.RecordTag, []string{"name", "size", "isDir", "isSymlink", "modified", "mode"})

// The File namespace (stdlib.md): every operation that can fail returns a Result with Error.

// deadliner is what a connection has and a regular file lacks: a read deadline, which a context
// sets to now to end a read that waits.
type deadliner interface {
	SetReadDeadline(t time.Time) error
}

// Handle is an open file, a standard stream, a connection, or a buffer over one: File.Handle. A
// reader holds the read slot, a writer the write mutex, and close takes the state mutex alone, so
// one goroutine may read a connection while another writes it, and a close from a third ends
// both. A write goes to the OS before it returns, except into the buffer of a buffered handle.
type Handle struct {
	state   sync.Mutex
	closed  bool
	c       io.Closer     // the file or connection, closed by File.close; nil for a standard stream
	reads   chan struct{} // the read slot, taken with the caller's context; nil when the handle does not read
	r       *bufio.Reader // the input; a read consumes it into pending
	pending string        // what a read consumed before its context ended it, for the next read
	scratch []byte        // what File.readChunk reads into when nothing is buffered; held by the read slot
	dl      deadliner     // a connection's read deadline; nil for a file, whose reads never wait
	pump    *pump         // standard input, read on a goroutine so a terminal cannot hold a cancelled run
	writes  sync.Mutex
	w       io.Writer     // the output: the file, the connection, the stream, or buf
	buf     *bufio.Writer // File.buffered's buffer, over under; nil for a handle that is not buffered
	under   *Handle       // the handle a buffered handle writes out to, reads, and closes
	std     bool          // a standard stream: close does nothing
	conn    bool          // a connection: Net.peer answers, and a child is given it through a copy
	via     *viaReader    // a pipe's reading end: its source, which handing it to a child changes
	peer    string        // a connection's remote address, for Net.peer
	// What a child process is given for a standard stream: the run's own (process.go).
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
	h.w = w
	return h
}

// defaultBuffer is the size of File.buffered's buffer when none is given, Go's.
const defaultBuffer = 4096

// buffered is File.buffered: f itself when it is buffered with at least size bytes, as Go's
// bufio.NewWriterSize returns a writer already large enough, and otherwise a handle whose writes
// collect in a buffer of size bytes over f. A handle that cannot write gives one that cannot.
func buffered(f *Handle, size int) *Handle {
	if size < 1 {
		size = defaultBuffer
	}
	if f.buf != nil && f.buf.Size() >= size {
		return f
	}
	h := &Handle{under: f, buf: bufio.NewWriterSize(handleWriter{f}, size)}
	if f.w != nil {
		h.w = h.buf
	}
	return h
}

// source is the handle a read of h reads: h, or for a buffered handle the one under it, whose
// read buffer it shares. A closed buffered handle is its own source, and reads nothing.
func (h *Handle) source() *Handle {
	for h.under != nil && !h.isClosed() {
		h = h.under
	}
	return h
}

// sink is the handle under every buffer of h, where its writes end up.
func (h *Handle) sink() *Handle {
	for h.under != nil {
		h = h.under
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

// write writes s under the write mutex: to the OS, all of it or an error, or into the buffer of
// a buffered handle.
func (h *Handle) write(s string) value.Value {
	if h.w == nil {
		return value.Err(value.ErrorValue(errNotWritable))
	}
	h.writes.Lock()
	defer h.writes.Unlock()
	if h.isClosed() {
		return value.Err(value.ErrorValue(errNotWritable))
	}
	_, err := io.WriteString(h.w, s)
	return okVoid(err)
}

// flush is File.flush: the buffer of h written out, then that of each buffered handle under it,
// so that what was written to h has reached the OS. A handle that is not buffered has nothing
// to write out. The first error of a buffer's writing out is kept by its bufio.Writer, which
// returns it from every later write and flush.
func (h *Handle) flush() error {
	for ; h.buf != nil; h = h.under {
		h.writes.Lock()
		var err error
		if h.isClosed() {
			err = errors.New(errNotWritable)
		} else {
			err = h.buf.Flush()
		}
		h.writes.Unlock()
		if err != nil {
			return err
		}
	}
	return nil
}

// close closes h: a buffered handle writes its buffer out and closes the handle under it,
// returning the first error of the two. A write or flush in progress on another goroutine, which
// may wait on a peer that stopped reading, is not waited for: the handle under is closed first,
// which ends it, and what the buffer still holds is lost and reported. Under a standard stream,
// which a close leaves open, the close waits for it and then writes out. A standard stream and a
// closed handle give nil, and a close ends a read or a write of a file or a connection in
// progress on another goroutine.
func (h *Handle) close() error {
	if h.std {
		return nil
	}
	h.state.Lock()
	if h.closed {
		h.state.Unlock()
		return nil
	}
	h.closed = true
	h.state.Unlock()
	if h.buf == nil {
		return h.c.Close()
	}
	if h.writes.TryLock() {
		err := h.buf.Flush()
		h.writes.Unlock()
		if cerr := h.under.close(); err == nil {
			err = cerr
		}
		return err
	}
	cerr := h.under.close()
	h.writes.Lock()
	err := h.buf.Flush() // Ok when nothing is left; Err when the buffer is lost or the write failed
	h.writes.Unlock()
	if err == nil {
		err = cerr
	}
	return err
}

// handleWriter writes into h under its write mutex: a buffered handle's buffer written out to
// the handle under it, and a child's output copied into a handle that is not a file.
type handleWriter struct{ h *Handle }

func (w handleWriter) Write(b []byte) (int, error) {
	w.h.writes.Lock()
	defer w.h.writes.Unlock()
	if w.h.w == nil || w.h.isClosed() {
		return 0, errors.New(errNotWritable)
	}
	return w.h.w.Write(b)
}

// copyTo is File.pipe's copy, the read slot of h and the write mutex of the destination held:
// what h read ahead, then the rest of its input. w is not buffered unless File.buffered made it
// so, so io.Copy writes each chunk as it reads it, and a source that waits streams; between
// files and sockets it takes the kernel's copy. What was written of pending is dropped from it,
// so a later read does not deliver it again.
func (h *Handle) copyTo(w io.Writer) error {
	if h.pending != "" {
		n, err := io.WriteString(w, h.pending)
		h.pending = h.pending[n:]
		if err != nil {
			return err
		}
	}
	_, err := io.Copy(w, h.r)
	return err
}

// SyncName names the handle for rendering.
func (h *Handle) SyncName() string { return "File.Handle" }

func handleArg(g *interp.Goroutine, who string, v value.Value) *Handle {
	return typedArg[*Handle](g, who, "a File.Handle", v)
}

func init() {
	register("File", func() value.Value {
		return namespace("File", append([]any{
			constant("Handle", value.NewTag("File.Handle", value.SyncTag, nil)),
			constant("Info", InfoTag),
			fn("File.stat", "path, follow?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return statFile(stringArg(g, who, a[0]), boolOr(g, who, a[1], true))
			}),
			fn("File.mkdir", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(os.MkdirAll(stringArg(g, who, a[0]), newDirMode))
			}),
			fn("File.removeAll", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(os.RemoveAll(stringArg(g, who, a[0])))
			}),
			fn("File.rename", "from, to", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(os.Rename(stringArg(g, who, a[0]), stringArg(g, who, a[1])))
			}),
			fn("File.tempDir", "pattern, dir?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return tempDir(stringArg(g, who, a[0]), tempParent(g, who, a[1]))
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
				f, err := os.OpenFile(stringArg(g, who, a[0]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, newFileMode)
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
			fn("File.write", "path, text, replace?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return writeFile(stringArg(g, who, a[0]), []byte(stringArg(g, who, a[1])), boolOr(g, who, a[2], true))
			}),
			fn("File.writeBytes", "path, bytes, replace?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return writeFile(stringArg(g, who, a[0]), view(bytesArg(g, who, a[1])), boolOr(g, who, a[2], true))
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
			fn("File.create", "path, replace?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				f, err := os.OpenFile(stringArg(g, who, a[0]), os.O_RDWR|createFlags(boolOr(g, who, a[1], true)), newFileMode)
				if err != nil {
					return errResult(err)
				}
				return value.Ok(newHandle(f, nil, f))
			}),
			fn("File.readLine", "f, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0]).source()
				ctx := waitContext(g, who, a[1])
				defer g.AfterWait()
				return h.read(ctx, h.readLine)
			}),
			fn("File.readAll", "f, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0]).source()
				ctx := waitContext(g, who, a[1])
				defer g.AfterWait()
				return h.read(ctx, h.readAll)
			}),
			fn("File.readChunk", "f, max, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0]).source()
				max := intArg(g, who, a[1])
				ctx := waitContext(g, who, a[2])
				if max < 1 {
					return value.Err(value.ErrorValue(fmt.Sprintf("the most to read is %d; it must be at least 1", max)))
				}
				defer g.AfterWait()
				return h.read(ctx, func() value.Value { return h.readChunk(int(min(max, scratchCap))) })
			}),
			fn("File.readExactly", "f, n, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				h := handleArg(g, who, a[0]).source()
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
			fn("File.buffered", "f, size?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				size := int64(defaultBuffer)
				if a[1] != nil {
					size = intArg(g, who, a[1])
				}
				return buffered(handleArg(g, who, a[0]), int(min(size, math.MaxInt32)))
			}),
			fn("File.flush", "f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(handleArg(g, who, a[0]).flush())
			}),
			fn("File.pipe", "src, dst, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				src := handleArg(g, who, a[0])
				dst := handleArg(g, who, a[1])
				ctx := waitContext(g, who, a[2])
				src = src.source()
				if src == dst.sink() {
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
					return okVoid(src.copyTo(dst.w))
				})
			}),
			fn("File.close", "f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return okVoid(handleArg(g, who, a[0]).close())
			}),
		}, fileFS...)...)
	})
}
