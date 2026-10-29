package frame

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestWriteThenRead(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, []byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "Content-Length: 7\r\n\r\n{\"a\":1}"; got != want {
		t.Errorf("wrote %q, want %q", got, want)
	}
	r := bufio.NewReader(&buf)
	body, err := Read(r)
	if err != nil || string(body) != `{"a":1}` {
		t.Errorf("read %q, %v", body, err)
	}
	if _, err := Read(r); !errors.Is(err, ErrClosed) {
		t.Errorf("at the end: %v, want ErrClosed", err)
	}
}

func TestReadHeaders(t *testing.T) {
	// Other headers, in any case, with a line feed alone or with a carriage return.
	r := bufio.NewReader(strings.NewReader("Content-Type: application/json\r\ncontent-length: 2\n\r\nok"))
	body, err := Read(r)
	if err != nil || string(body) != "ok" {
		t.Errorf("read %q, %v", body, err)
	}
}

func TestReadErrors(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"no colon\r\n\r\n", "malformed header"},
		{"Content-Length: many\r\n\r\n", "malformed Content-Length"},
		{"Content-Type: x\r\n\r\n", "without Content-Length"},
		{"Content-Length: 999999999999\r\n\r\n", "more than"},
		{"Content-Length: 5\r\n\r\nab", "EOF"},
		{"Content-Length: 5", "EOF"}, // the stream ends inside a header
	} {
		_, err := Read(bufio.NewReader(strings.NewReader(tc.in)))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: error %v, want one with %q", tc.in, err, tc.want)
		}
		if errors.Is(err, ErrClosed) {
			t.Errorf("%q: ErrClosed is for the end between messages", tc.in)
		}
	}
}

type failing struct{ after int }

func (f *failing) Write(p []byte) (int, error) {
	if f.after <= 0 {
		return 0, errors.New("closed")
	}
	f.after--
	return len(p), nil
}

func TestWriteErrors(t *testing.T) {
	if err := Write(&failing{}, []byte("x")); err == nil {
		t.Error("a header that cannot be written")
	}
	if err := Write(&failing{after: 1}, []byte("x")); err == nil {
		t.Error("a body that cannot be written")
	}
}
