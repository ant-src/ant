package lib

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// ResponseTag is Http.Response: { status: int, headers: Map[string, string], body: string }.
var ResponseTag = value.NewTag("Http.Response", value.RecordTag, []string{"status", "headers", "body"})

// timeout converts an optional Time.Duration to a Go duration by goDuration's rules; zero means
// no deadline, and an invalid duration is the Err of the request.
func timeout(g *interp.Goroutine, who string, v value.Value) (time.Duration, error) {
	if !given(v) {
		return 0, nil
	}
	return goDuration(durationArg(g, who, v))
}

// headers reads an optional Map[string, string] of request headers.
func headers(g *interp.Goroutine, who string, v value.Value, req *http.Request) {
	if !given(v) {
		return
	}
	for _, e := range mapArg(g, who, v).Entries() {
		k, ok := e.Key.(string)
		if !ok {
			g.Panic("%s needs string header names, got %s", who, value.Describe(e.Key))
		}
		val, ok := e.Value.(string)
		if !ok {
			g.Panic("%s needs string header values, got %s", who, value.Describe(e.Value))
		}
		if http.CanonicalHeaderKey(k) == "Host" {
			req.Host = val // which Go sends from the request, ignoring a Host in the headers
			continue
		}
		req.Header.Set(k, val)
	}
}

// client sends one request and follows no redirect, so a 3xx is the response, as stdlib.md
// says of any status.
var client = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

// request is Http.request and, with the method filled in, Http.get and Http.post: one request
// under the run's context, the given context and the timeout, the whole response read. A
// response of any status is Ok; a failure to get one is Err, with the context's reason, or
// "timeout", when a bound ended it.
func request(g *interp.Goroutine, who, method, url, body string, hdrs, limit, c value.Value) value.Value {
	if method == "" {
		return errResult(errors.New("the method is empty"))
	}
	wait, err := timeout(g, who, limit)
	if err != nil {
		return errResult(err)
	}
	ctx := waitContext(g, optionalContext(g, who, c))
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
		return errResult(err)
	}
	headers(g, who, hdrs, req)
	g.Instance().Flush()
	resp, err := client.Do(req)
	g.AfterWait()
	if err != nil {
		return failed(ctx, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return failed(ctx, err)
	}
	out := value.NewMap()
	for k, vs := range resp.Header {
		out.Set(k, strings.Join(vs, ", "))
	}
	return value.Ok(value.NewRecord(ResponseTag, []value.Value{int64(resp.StatusCode), out, string(data)}))
}

// failed is the Err of a request that got no response: the context's reason when a bound
// ended it, the Go error otherwise.
func failed(ctx context.Context, err error) value.Value {
	if ctx.Err() != nil {
		return gaveUp(ctx)
	}
	return errResult(err)
}

func optionalString(g *interp.Goroutine, who string, v value.Value) string {
	if !given(v) {
		return ""
	}
	return stringArg(g, who, v)
}

// RequestTag is Http.Request: what a handler of Http.serve receives.
var RequestTag = value.NewTag("Http.Request", value.RecordTag, []string{"method", "path", "query", "headers", "body", "remote"})

// serve is Http.serve: an http.Server on the listener until ctx is done, then a shutdown that
// waits up to five seconds for the requests in flight. Each request's handler runs on an Ant
// goroutine of its own; a panic there is reported as any goroutine's is and answered with 500.
func serve(g *interp.Goroutine, a []value.Value) value.Value {
	const who = "Http.serve"
	l := listenerArg(g, who, a[0])
	handler := funcArg(g, who, a[1])
	ctx := waitContext(g, optionalContext(g, who, a[2]))
	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serveRequest(g, handler, w, r)
		}),
	}
	g.Instance().Flush()
	done := make(chan error, 1)
	go func() { done <- srv.Serve(l.l) }()
	select {
	case err := <-done:
		g.AfterWait()
		if l.isClosed() {
			return value.Err(value.ErrorValue("the listener is closed"))
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
// goroutine, and its Http.Response is written back, or 500 when the handler panicked.
func serveRequest(g *interp.Goroutine, handler value.Value, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	headers := value.NewMap()
	headers.Set("Host", r.Host) // which Go keeps apart from the other headers
	for k, vs := range r.Header {
		headers.Set(k, strings.Join(vs, ", "))
	}
	req := value.NewRecord(RequestTag, []value.Value{r.Method, r.URL.Path, r.URL.RawQuery, headers, string(body), r.RemoteAddr})
	result := make(chan *value.Record, 1)
	g.Spawn(func(ng *interp.Goroutine) {
		defer close(result)
		resp, ok := ng.Call(handler, req).(*value.Record)
		if !ok || resp.Tag != ResponseTag {
			ng.Panic("Http.serve: the handler must return an Http.Response")
		}
		status, _ := resp.Field("status")
		if n, ok := status.(int64); !ok || n < 200 || n > 599 {
			ng.Panic("Http.serve: the response status %s is outside 200 to 599", value.Render(status))
		}
		result <- resp
	})
	resp, ok := <-result
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	status, _ := resp.Field("status")
	hdrs, _ := resp.Field("headers")
	text, _ := resp.Field("body")
	if m, ok := hdrs.(*value.Map); ok {
		for _, e := range m.Entries() {
			k, kok := e.Key.(string)
			v, vok := e.Value.(string)
			if kok && vok {
				w.Header().Set(k, v)
			}
		}
	}
	out, _ := text.(string)
	w.Header().Set("Content-Length", strconv.Itoa(len(out)))
	w.WriteHeader(int(status.(int64)))
	io.WriteString(w, out)
}

func init() {
	register("Http", func() value.Value {
		return namespace("Http",
			constant("Response", ResponseTag),
			fn("Http.request", "method, url, body?, headers?, timeout?, ctx?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Http.request"
				return request(g, who, stringArg(g, who, a[0]), stringArg(g, who, a[1]), optionalString(g, who, a[2]), a[3], a[4], a[5])
			}),
			fn("Http.get", "url, headers?, timeout?, ctx?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Http.get"
				return request(g, who, "GET", stringArg(g, who, a[0]), "", a[1], a[2], a[3])
			}),
			fn("Http.post", "url, body, headers?, timeout?, ctx?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Http.post"
				return request(g, who, "POST", stringArg(g, who, a[0]), stringArg(g, who, a[1]), a[2], a[3], a[4])
			}),
			constant("Request", RequestTag),
			fn("Http.serve", "l, handler, ctx?", serve),
		)
	})
}
