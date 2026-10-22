package lib

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// ResponseTag is Http.Response: { status: int, headers: Map[string, string], body: string }.
var ResponseTag = value.NewTag("Http.Response", value.RecordTag, []string{"status", "headers", "body"})

// timeoutError is the Err of a request whose timeout passed.
var timeoutError = errors.New("timeout")

// timeout converts an optional Time.Duration to a Go duration by the rules of stdlib.md: hours
// and smaller units are exact and a day is 24 hours; years, months or weeks, which have no
// fixed length, and a zero or negative duration are errors; a duration beyond Go's range is
// taken as that range. Zero means no deadline.
func timeout(g *interp.Goroutine, who string, v value.Value) (time.Duration, error) {
	if !given(v) {
		return 0, nil
	}
	d := durationArg(g, who, v)
	if d.Years != 0 || d.Months != 0 || d.Weeks != 0 {
		return 0, errors.New("the timeout has years, months or weeks, which have no fixed length")
	}
	if d.Sign() <= 0 {
		return 0, errors.New("the timeout is zero or negative")
	}
	seconds := float64(d.Days)*86400 + float64(d.Hours)*3600 + float64(d.Minutes)*60 + float64(d.Seconds) +
		float64(d.Milliseconds)/1e3 + float64(d.Microseconds)/1e6 + float64(d.Nanoseconds)/1e9
	ns := seconds * 1e9
	if ns >= math.MaxInt64 {
		return time.Duration(math.MaxInt64), nil
	}
	return time.Duration(ns), nil
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
// under the run's context and the timeout, the whole response read. A response of any status
// is Ok; a failure to get one is Err, with the message "timeout" when the timeout passed.
func request(g *interp.Goroutine, who, method, url, body string, hdrs, limit value.Value) value.Value {
	wait, err := timeout(g, who, limit)
	if err != nil {
		return errResult(err)
	}
	ctx := g.Context()
	if wait > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, wait)
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
		return errResult(timedOut(ctx, err))
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return errResult(timedOut(ctx, err))
	}
	out := value.NewMap()
	for k, vs := range resp.Header {
		out.Set(k, strings.Join(vs, ", "))
	}
	return value.Ok(value.NewRecord(ResponseTag, []value.Value{int64(resp.StatusCode), out, string(data)}))
}

// timedOut replaces err with the timeout error when the request's own deadline passed.
func timedOut(ctx context.Context, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return timeoutError
	}
	return err
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
			fn("Http.request", "method, url, body?, headers?, timeout?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Http.request"
				return request(g, who, stringArg(g, who, a[0]), stringArg(g, who, a[1]), optionalString(g, who, a[2]), a[3], a[4])
			}),
			fn("Http.get", "url, headers?, timeout?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Http.get"
				return request(g, who, "GET", stringArg(g, who, a[0]), "", a[1], a[2])
			}),
			fn("Http.post", "url, body, headers?, timeout?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Http.post"
				return request(g, who, "POST", stringArg(g, who, a[0]), stringArg(g, who, a[1]), a[2], a[3])
			}),
		)
	})
}
