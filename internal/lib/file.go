package lib

import (
	"bufio"
	"io"
	"os"
	"sync"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// The File namespace (stdlib.md): every operation that can fail returns a Result with Error.

// Handle is an open file: File.Handle.
type Handle struct {
	mu     sync.Mutex
	f      *os.File
	r      *bufio.Reader
	w      *bufio.Writer
	closed bool
}

// SyncName names the handle for rendering.
func (h *Handle) SyncName() string { return "File.Handle" }

func handleArg(g *interp.Goroutine, who string, v value.Value) *Handle {
	h, ok := v.(*Handle)
	if !ok {
		g.Panic("%s needs a File.Handle, got %s", who, value.Describe(v))
	}
	return h
}

func errResult(err error) value.Value { return value.Err(value.ErrorValue(err.Error())) }

func bytesArg(g *interp.Goroutine, who string, v value.Value) []byte {
	vals := arrayArg(g, who, v).Values()
	out := make([]byte, len(vals))
	for i, x := range vals {
		b, ok := x.(uint8)
		if !ok {
			g.Panic("%s needs an array of bytes, got an element of kind %s", who, value.Describe(x))
		}
		out[i] = b
	}
	return out
}

func bytesValue(b []byte) value.Value {
	out := make([]value.Value, len(b))
	for i, x := range b {
		out[i] = uint8(x)
	}
	return value.NewArray(out)
}

func init() {
	register("File", func() value.Value {
		return namespace("File",
			constant("Handle", value.NewTag("File.Handle", value.SyncTag, nil)),
			fn("File.read", "path", func(g *interp.Goroutine, a []value.Value) value.Value {
				data, err := os.ReadFile(stringArg(g, "File.read", a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(string(data))
			}),
			fn("File.readBytes", "path", func(g *interp.Goroutine, a []value.Value) value.Value {
				data, err := os.ReadFile(stringArg(g, "File.readBytes", a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(bytesValue(data))
			}),
			fn("File.write", "path, text", func(g *interp.Goroutine, a []value.Value) value.Value {
				err := os.WriteFile(stringArg(g, "File.write", a[0]), []byte(stringArg(g, "File.write", a[1])), 0o644)
				if err != nil {
					return errResult(err)
				}
				return value.Ok(value.Void)
			}),
			fn("File.writeBytes", "path, bytes", func(g *interp.Goroutine, a []value.Value) value.Value {
				err := os.WriteFile(stringArg(g, "File.writeBytes", a[0]), bytesArg(g, "File.writeBytes", a[1]), 0o644)
				if err != nil {
					return errResult(err)
				}
				return value.Ok(value.Void)
			}),
			fn("File.exists", "path", func(g *interp.Goroutine, a []value.Value) value.Value {
				_, err := os.Stat(stringArg(g, "File.exists", a[0]))
				return err == nil
			}),
			fn("File.remove", "path", func(g *interp.Goroutine, a []value.Value) value.Value {
				if err := os.Remove(stringArg(g, "File.remove", a[0])); err != nil {
					return errResult(err)
				}
				return value.Ok(value.Void)
			}),
			fn("File.list", "dir", func(g *interp.Goroutine, a []value.Value) value.Value {
				entries, err := os.ReadDir(stringArg(g, "File.list", a[0]))
				if err != nil {
					return errResult(err)
				}
				names := make([]value.Value, len(entries))
				for i, e := range entries {
					names[i] = e.Name()
				}
				return value.Ok(value.NewArray(names))
			}),
			fn("File.open", "path", func(g *interp.Goroutine, a []value.Value) value.Value {
				f, err := os.Open(stringArg(g, "File.open", a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(&Handle{f: f, r: bufio.NewReader(f)})
			}),
			fn("File.create", "path", func(g *interp.Goroutine, a []value.Value) value.Value {
				f, err := os.Create(stringArg(g, "File.create", a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(&Handle{f: f, w: bufio.NewWriter(f)})
			}),
			fn("File.readLine", "f", func(g *interp.Goroutine, a []value.Value) value.Value {
				h := handleArg(g, "File.readLine", a[0])
				h.mu.Lock()
				defer h.mu.Unlock()
				if h.closed || h.r == nil {
					return value.Err(value.ErrorValue("the file is not open for reading"))
				}
				line, err := h.r.ReadString('\n')
				if err != nil && err != io.EOF {
					return errResult(err)
				}
				if line == "" && err == io.EOF {
					return value.Ok(value.Null)
				}
				if len(line) > 0 && line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
					if len(line) > 0 && line[len(line)-1] == '\r' {
						line = line[:len(line)-1]
					}
				}
				return value.Ok(line)
			}),
			fn("File.writeLine", "f, line", func(g *interp.Goroutine, a []value.Value) value.Value {
				h := handleArg(g, "File.writeLine", a[0])
				line := stringArg(g, "File.writeLine", a[1])
				h.mu.Lock()
				defer h.mu.Unlock()
				if h.closed || h.w == nil {
					return value.Err(value.ErrorValue("the file is not open for writing"))
				}
				if _, err := h.w.WriteString(line + "\n"); err != nil {
					return errResult(err)
				}
				return value.Ok(value.Void)
			}),
			fn("File.pipe", "src, dst", func(g *interp.Goroutine, a []value.Value) value.Value {
				src := handleArg(g, "File.pipe", a[0])
				dst := handleArg(g, "File.pipe", a[1])
				src.mu.Lock()
				defer src.mu.Unlock()
				dst.mu.Lock()
				defer dst.mu.Unlock()
				if src.closed || src.r == nil {
					return value.Err(value.ErrorValue("the source is not open for reading"))
				}
				if dst.closed || dst.w == nil {
					return value.Err(value.ErrorValue("the destination is not open for writing"))
				}
				if _, err := io.Copy(dst.w, src.r); err != nil {
					return errResult(err)
				}
				return value.Ok(value.Void)
			}),
			fn("File.close", "f", func(g *interp.Goroutine, a []value.Value) value.Value {
				h := handleArg(g, "File.close", a[0])
				h.mu.Lock()
				defer h.mu.Unlock()
				if h.closed {
					return value.Void
				}
				if h.w != nil {
					h.w.Flush()
				}
				h.f.Close()
				h.closed = true
				return value.Void
			}),
		)
	})
}
