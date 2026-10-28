package lsp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// conn carries JSON-RPC 2.0 messages over a stream with the base protocol's Content-Length
// framing (spec/lsp.md section 1).
type conn struct {
	r  *bufio.Reader
	w  io.Writer
	mu sync.Mutex // one writer at a time
}

// message is an incoming request, notification, or response.
type message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// request reports whether the message is a request: it has an id and a method.
func (m *message) request() bool { return m.Method != "" && len(m.ID) > 0 }

// rpcError is the error of a response.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return e.Message }

// The error codes the server uses.
const (
	codeParseError       = -32700
	codeInvalidRequest   = -32600
	codeMethodNotFound   = -32601
	codeInvalidParams    = -32602
	codeInternalError    = -32603
	codeServerNotInit    = -32002
	codeRequestCancelled = -32800
	codeContentModified  = -32801
)

var errClosed = errors.New("the connection ended")

// maxMessage bounds a message's body, so that a malformed header cannot ask for all memory.
const maxMessage = 256 << 20

// read reads one message, or returns errClosed at the end of the stream.
func (c *conn) read() (*message, error) {
	length := -1
	for {
		line, err := c.r.ReadString('\n')
		if err != nil {
			if err == io.EOF && line == "" {
				return nil, errClosed
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
	if length > maxMessage {
		return nil, fmt.Errorf("a message of %d bytes, more than the %d allowed", length, maxMessage)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(c.r, body); err != nil {
		return nil, err
	}
	var m message
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, &rpcError{Code: codeParseError, Message: err.Error()}
	}
	return &m, nil
}

// write sends one message.
func (c *conn) write(v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err = c.w.Write(body)
	return err
}

// reply answers a request with a result.
func (c *conn) reply(id json.RawMessage, result any) error {
	return c.write(struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  any             `json:"result"`
	}{"2.0", id, result})
}

// fail answers a request with an error.
func (c *conn) fail(id json.RawMessage, code int, msg string) error {
	return c.write(struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Error   rpcError        `json:"error"`
	}{"2.0", id, rpcError{code, msg}})
}

// notify sends a notification.
func (c *conn) notify(method string, params any) error {
	return c.write(struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params"`
	}{"2.0", method, params})
}
