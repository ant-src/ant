package lib

import (
	"bytes"
	"io"
	"runtime"
	"testing"

	"github.com/ant-src/ant/value"
)

// eofWithData returns io.EOF together with its last bytes, as io.Reader allows.
type eofWithData struct{ data []byte }

func (r *eofWithData) Read(p []byte) (int, error) {
	n := copy(p, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		return n, io.EOF
	}
	return n, nil
}

// TestReadExactlyWithEOF: a reader may return its last bytes together with io.EOF, which bufio
// passes through on a read at least as large as its buffer; those bytes complete the read.
func TestReadExactlyWithEOF(t *testing.T) {
	for _, n := range []int{1, 4095, 4096, 8192, 100000} {
		data := bytes.Repeat([]byte{0xab}, n)
		h := newHandle(nil, &eofWithData{data}, nil)
		got := h.readExactly(n)
		if r, ok := got.(*value.Record); !ok || r.Tag != value.OkTag || r.Get(0) != value.Bytes(data) {
			t.Errorf("readExactly(%d) = %s", n, value.Render(got))
		}
		if again := h.readExactly(1); value.Render(again) != "Ok{ value: null }" {
			t.Errorf("after %d bytes: %s", n, value.Render(again))
		}
	}
	h := newHandle(nil, &eofWithData{[]byte("abc")}, nil)
	if got := value.Render(h.readExactly(5)); got != `Err{ error: Error{ message: "unexpected end of file" } }` {
		t.Errorf("short: %s", got)
	}
	if got := value.Render(h.readChunk(10)); got != "Ok{ value: null }" {
		t.Errorf("a short read discards its bytes: %s", got)
	}
}

// drip returns size bytes per read, as a socket delivers short messages.
type drip struct{ size int }

func (d drip) Read(p []byte) (int, error) { return copy(p, bytes.Repeat([]byte{'m'}, d.size)), nil }

// TestReadChunkAllocatesTheChunk: a large max over short messages allocates about the
// messages, not max per call.
func TestReadChunkAllocatesTheChunk(t *testing.T) {
	h := newHandle(nil, drip{100}, nil)
	h.readChunk(scratchCap) // the scratch buffer, once
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for range 1000 {
		r := h.readChunk(scratchCap).(*value.Record)
		if b := r.Get(0).(value.Bytes); len(b) != 100 {
			t.Fatalf("chunk of %d bytes", len(b))
		}
	}
	runtime.ReadMemStats(&after)
	if per := (after.TotalAlloc - before.TotalAlloc) / 1000; per > 1024 {
		t.Errorf("%d bytes allocated per 100-byte chunk", per)
	}
}

// TestBytesValueTrims: Bytes share a nearly full array and copy one mostly unused, so a value
// holds no more memory than it needs.
func TestBytesValueTrims(t *testing.T) {
	full := make([]byte, 100, 110)
	if b := bytesValue(full); len(b) != 100 || &view(b)[0] != &full[0] {
		t.Error("a nearly full array is shared")
	}
	loose := make([]byte, 100, 4096)
	if b := bytesValue(loose); len(b) != 100 || &view(b)[0] == &loose[0] {
		t.Error("a mostly unused array is copied")
	}
	data, _ := io.ReadAll(io.LimitReader(drip{1000}, 3000))
	if b := bytesValue(data); len(b) != 3000 {
		t.Errorf("len %d", len(b))
	}
}
