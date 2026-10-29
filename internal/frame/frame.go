// Package frame reads and writes messages framed by a Content-Length header, as the Language
// Server Protocol and the Debug Adapter Protocol frame them (spec/lsp.md and spec/dap.md,
// section 1).
package frame

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrClosed is the error of a read at the end of the stream, between messages.
var ErrClosed = errors.New("the connection ended")

// max bounds a message's body, so that a malformed header cannot ask for all memory.
const max = 256 << 20

// Read reads one message's body, or returns ErrClosed at the end of the stream.
func Read(r *bufio.Reader) ([]byte, error) {
	length := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF && line == "" {
				return nil, ErrClosed
			}
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("malformed header %q", line)
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			length, err = strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("malformed Content-Length %q", value)
			}
		}
	}
	if length < 0 {
		return nil, errors.New("a message without Content-Length")
	}
	if length > max {
		return nil, fmt.Errorf("a message of %d bytes, more than the %d allowed", length, max)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

// Write writes one message with body as its content. The caller keeps writers apart.
func Write(w io.Writer, body []byte) error {
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err := w.Write(body)
	return err
}
