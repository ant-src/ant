package lib

import (
	"io"
	"net/http"
	"strings"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// ResponseTag is Http.Response: { status: int, headers: Map[string, string], body: string }.
var ResponseTag = value.NewTag("Http.Response", value.RecordTag, []string{"status", "headers", "body"})

func response(resp *http.Response) value.Value {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errResult(err)
	}
	headers := value.NewMap()
	for k, vs := range resp.Header {
		headers.Set(k, strings.Join(vs, ", "))
	}
	return value.Ok(value.NewRecord(ResponseTag, []value.Value{int64(resp.StatusCode), headers, string(body)}))
}

func init() {
	register("Http", func() value.Value {
		return namespace("Http",
			constant("Response", ResponseTag),
			fn("Http.get", "url", func(g *interp.Goroutine, a []value.Value) value.Value {
				req, err := http.NewRequestWithContext(g.Context(), "GET", stringArg(g, "Http.get", a[0]), nil)
				if err != nil {
					return errResult(err)
				}
				g.Instance().Flush()
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return errResult(err)
				}
				return response(resp)
			}),
			fn("Http.post", "url, body, headers", func(g *interp.Goroutine, a []value.Value) value.Value {
				body := stringArg(g, "Http.post", a[1])
				req, err := http.NewRequestWithContext(g.Context(), "POST", stringArg(g, "Http.post", a[0]), strings.NewReader(body))
				if err != nil {
					return errResult(err)
				}
				for _, e := range mapArg(g, "Http.post", a[2]).Entries() {
					k, ok := e.Key.(string)
					if !ok {
						g.Panic("Http.post needs string header names, got %s", value.Describe(e.Key))
					}
					v, ok := e.Value.(string)
					if !ok {
						g.Panic("Http.post needs string header values, got %s", value.Describe(e.Value))
					}
					req.Header.Set(k, v)
				}
				g.Instance().Flush()
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return errResult(err)
				}
				return response(resp)
			}),
		)
	})
}
