package lib

import (
	"context"
	"io"
	"net/http"
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
		req.Header.Set(k, val)
	}
}

// request is Http.request and, with the method filled in, Http.get and Http.post: one request
// under the run's context, the given context and the timeout, the whole response read. A
// response of any status is Ok; a failure to get one is Err, with the context's reason, or
// "timeout", when a bound ended it.
func request(g *interp.Goroutine, who, method, url, body string, hdrs, limit, c value.Value) value.Value {
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
	resp, err := http.DefaultClient.Do(req)
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
		)
	})
}
