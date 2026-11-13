package lib

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// ResponseTag is Http.Response: { status: int, headers: Map[string, string], bytes: Bytes }, the
// body held once, as received; Http.text decodes it.
var ResponseTag = value.NewTag("Http.Response", value.RecordTag, []string{"status", "headers", "bytes"})

// ReplyTag is Http.Reply: what a handler of Http.serve returns, with a body that is a string or
// Bytes.
var ReplyTag = value.NewTag("Http.Reply", value.RecordTag, []string{"status", "headers", "body"})

// timeout converts an optional Time.Duration to a Go duration by goDuration's rules; zero means
// no deadline, and an invalid duration is the Err of the request.
func timeout(g *interp.Goroutine, who string, v value.Value) (time.Duration, error) {
	if !given(v) {
		return 0, nil
	}
	return goDuration(durationArg(g, who, v))
}

// headers reads an optional Map[string, string] of request headers. Two names that differ only in
// case name one header, which the map would give two values, so they are an error.
func headers(g *interp.Goroutine, who string, v value.Value, req *http.Request) error {
	if !given(v) {
		return nil
	}
	seen := map[string]string{}
	for _, e := range mapArg(g, who, v).Entries() {
		k, ok := e.Key.(string)
		if !ok {
			g.Panic("%s needs string header names, got %s", who, value.Describe(e.Key))
		}
		val, ok := e.Value.(string)
		if !ok {
			g.Panic("%s needs string header values, got %s", who, value.Describe(e.Value))
		}
		lower := strings.ToLower(k)
		if first, ok := seen[lower]; ok {
			return fmt.Errorf("the headers %q and %q differ only in case", first, k)
		}
		seen[lower] = k
		if http.CanonicalHeaderKey(k) == "Host" {
			req.Host = val // which Go sends from the request, ignoring a Host in the headers
			continue
		}
		req.Header.Set(k, val)
	}
	return nil
}

// client sends one request and follows no redirect, so a 3xx is the response, as stdlib.md
// says of any status.
// Its transport asks for no compression and decodes none, and request sends no User-Agent of
// its own, so the request carries only the headers the program gave, Host and Content-Length,
// and the response's body and headers are as received.
var client = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	Transport:     plainTransport(),
}

func plainTransport() http.RoundTripper {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableCompression = true
	return t
}

// httpReason is the innermost reason of a failed request, without Go's layers: the text of a
// system error alone, as File errors give it ("connection refused"), and no "net/http:" or
// "http:" prefix.
func httpReason(err error) string {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno.Error()
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		return dns.Err
	}
	for {
		inner := errors.Unwrap(err)
		if inner == nil {
			break
		}
		err = inner
	}
	text := strings.TrimPrefix(err.Error(), "net/http: ")
	return strings.TrimPrefix(text, "http: ")
}

// isToken reports whether s is an HTTP token (RFC 9110), as a method must be.
func isToken(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || strings.IndexByte("!#$%&'*+-.^_`|~", c) >= 0) {
			return false
		}
	}
	return true
}

// sortedHeaders fills a map from Go's headers in order of name, after first, so the order is
// the same on every run; a repeated header is joined with ", ".
func sortedHeaders(out *value.Map, h http.Header) *value.Map {
	for _, k := range slices.Sorted(maps.Keys(h)) {
		out.Set(value.Text(k), value.Text(strings.Join(h[k], ", ")))
	}
	return out
}

// request is Http.request and, with the method filled in, Http.get and Http.post: one request
// under the run's context, the given context and the timeout, the whole response read. A
// response of any status is Ok; a failure to get one is Err, with the context's reason, or
// "timeout", when a bound ended it, and otherwise as `<method> "<url>": <reason>`, the method
// lowercased and the URL as given.
func request(g *interp.Goroutine, who, method, url, body string, hdrs, limit, c value.Value) value.Value {
	if !isToken(method) {
		return errResult(fmt.Errorf("request %q: the method %q is not a token", url, method))
	}
	fail := func(err error) value.Value {
		return errResult(fmt.Errorf("%s %q: %s", strings.ToLower(method), url, httpReason(err)))
	}
	wait, err := timeout(g, who, limit)
	if err != nil {
		return fail(err)
	}
	ctx := waitContext(g, who, c)
	if wait > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeoutCause(ctx, wait, value.ErrTimeout)
		defer cancel()
	}
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return fail(err)
	}
	req.Header["User-Agent"] = []string{""} // which makes Go's client send none of its own
	if err := headers(g, who, hdrs, req); err != nil {
		return fail(err)
	}
	resp, err := client.Do(req)
	g.AfterWait()
	if err != nil {
		if ctx.Err() != nil {
			return gaveUp(ctx)
		}
		return fail(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if ctx.Err() != nil {
			return gaveUp(ctx)
		}
		return fail(err)
	}
	out := sortedHeaders(value.NewMap(), resp.Header)
	return value.Ok(value.NewRecord(ResponseTag, []value.Value{int64(resp.StatusCode), out, bytesValue(data)}))
}

// optionalBody is a body of Http.request, a string or Bytes, or none.
func optionalBody(g *interp.Goroutine, who string, v value.Value) string {
	if !given(v) {
		return ""
	}
	return textOrBytes(g, who, v)
}

// RequestTag is Http.Request: what a handler of Http.serve receives, the body held once, as
// received; Http.text decodes it.
var RequestTag = value.NewTag("Http.Request", value.RecordTag, []string{"method", "path", "query", "headers", "bytes", "remote"})

// serve is Http.serve: an http.Server on the listener until ctx is done, then a shutdown that
// waits up to five seconds for the requests in flight. Each request's handler runs on an Ant
// goroutine of its own; a panic there is reported as any goroutine's is and answered with 500.
func serve(g *interp.Goroutine, who string, a []value.Value) value.Value {
	l := listenerArg(g, who, a[0])
	handler := funcArg(g, who, a[1])
	ctx := waitContext(g, who, a[2])
	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serveRequest(g, handler, w, r)
		}),
	}
	done := make(chan error, 1)
	go func() { done <- srv.Serve(newServeListener(l, g.Context().Done())) }()
	select {
	case err := <-done:
		g.AfterWait()
		if l.isClosed() {
			return value.Err(value.ErrorValue(errListenerClosed))
		}
		return errResult(err)
	case <-ctx.Done():
		// The shutdown closes the listener and waits for the requests in flight; past the
		// five seconds, Close ends their connections. Either way the listener is closed now.
		stop, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(stop); err != nil {
			srv.Close()
		}
		l.markClosed()
		<-done
		g.AfterWait()
		return value.Ok(value.Void)
	}
}

// serveRequest answers one request: the Http.Request record goes to the handler on a new Ant
// goroutine, and its Http.Reply is written back, or 500 when the handler panicked.
func serveRequest(g *interp.Goroutine, handler value.Value, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	headers := value.NewMap()
	headers.Set("Host", value.Text(r.Host)) // which Go keeps apart from the other headers
	sortedHeaders(headers, r.Header)
	req := value.NewRecord(RequestTag, []value.Value{value.Text(r.Method), value.Text(r.URL.Path), value.Text(r.URL.RawQuery), headers, bytesValue(body), r.RemoteAddr})
	type reply struct {
		resp *value.Record
		body string // the Reply's body, a string or Bytes, as the bytes to send
	}
	result := make(chan reply, 1)
	g.Spawn(func(ng *interp.Goroutine) {
		defer close(result)
		resp, ok := ng.Call(handler, req).(*value.Record)
		if !ok || resp.Tag != ReplyTag {
			ng.Panic("Http.serve: the handler must return an Http.Reply")
		}
		status, _ := resp.Field("status")
		if n, ok := status.(int64); !ok || n < 200 || n > 599 {
			ng.Panic("Http.serve: the response status %s is outside 200 to 599", value.Render(status))
		}
		// Two names that differ only in case name one header, as in Http.request.
		if hdrs, ok := resp.Field("headers"); ok {
			if m, ok := hdrs.(*value.Map); ok {
				seen := map[string]string{}
				for _, e := range m.Entries() {
					if k, ok := e.Key.(string); ok {
						lower := strings.ToLower(k)
						if first, ok := seen[lower]; ok {
							ng.Panic("Http.serve: the headers %q and %q differ only in case", first, k)
						}
						seen[lower] = k
					}
				}
			}
		}
		body, _ := resp.Field("body")
		result <- reply{resp, textOrBytes(ng, "Http.serve", body)}
	})
	answer, ok := <-result
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	status, _ := answer.resp.Field("status")
	hdrs, _ := answer.resp.Field("headers")
	if m, ok := hdrs.(*value.Map); ok {
		for _, e := range m.Entries() {
			k, kok := e.Key.(string)
			v, vok := e.Value.(string)
			if kok && vok {
				w.Header().Set(k, v)
			}
		}
	}
	if _, set := w.Header()["Content-Type"]; !set {
		w.Header()["Content-Type"] = nil // sent as given: no type sniffed from the body
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(answer.body)))
	w.WriteHeader(int(status.(int64)))
	io.WriteString(w, answer.body)
}

func init() {
	register("Http", func() value.Value {
		return namespace("Http",
			constant("Response", ResponseTag),
			fn("Http.request", "method, url, body?, headers?, timeout?, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return request(g, who, stringArg(g, who, a[0]), stringArg(g, who, a[1]), optionalBody(g, who, a[2]), a[3], a[4], a[5])
			}),
			fn("Http.get", "url, headers?, timeout?, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return request(g, who, "GET", stringArg(g, who, a[0]), "", a[1], a[2], a[3])
			}),
			fn("Http.post", "url, body, headers?, timeout?, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return request(g, who, "POST", stringArg(g, who, a[0]), textOrBytes(g, who, a[1]), a[2], a[3], a[4])
			}),
			constant("Request", RequestTag),
			constant("Reply", ReplyTag),
			fn("Http.text", "r", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if r, ok := a[0].(*value.Record); ok && (r.Tag == ResponseTag || r.Tag == RequestTag) {
					b, _ := r.Field("bytes")
					if b, ok := b.(*value.Bytes); ok {
						return value.Text(string(b.View()))
					}
				}
				g.Panic("%s needs an Http.Response or an Http.Request, got %s", who, value.Describe(a[0]))
				return nil
			}),
			fn("Http.serve", "l, handler, ctx?", serve),
		)
	})
}
