package lib

import (
	"encoding/csv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// separator reads the optional sep argument, a comma when omitted, and panics unless it is one
// code point other than a quote, a line break, U+0000 or U+FFFD, the rule of Go's reader and
// writer. The separator is part of the program, not of the data, so the check is a panic.
func separator(g *interp.Goroutine, who string, v value.Value) rune {
	if !given(v) {
		return ','
	}
	s := stringArg(g, who, v)
	r, size := utf8.DecodeRuneInString(s)
	if size != len(s) || r == '"' || r == '\r' || r == '\n' || r == 0 || r == utf8.RuneError {
		g.Panic("%s needs a separator of one character other than a quote, a line break, U+0000 or U+FFFD, got %q", who, s)
	}
	return r
}

// csvField quotes a field when it holds the separator, a quote or a line break, or begins with
// white space, doubling a quote inside.
func csvField(sb *strings.Builder, field string, sep rune) {
	needs := strings.ContainsRune(field, sep) || strings.ContainsAny(field, "\"\r\n")
	if r, _ := utf8.DecodeRuneInString(field); field != "" && unicode.IsSpace(r) {
		needs = true
	}
	if !needs {
		sb.WriteString(field)
		return
	}
	sb.WriteByte('"')
	sb.WriteString(strings.ReplaceAll(field, `"`, `""`))
	sb.WriteByte('"')
}

func init() {
	register("Csv", func() value.Value {
		return namespace("Csv",
			fn("Csv.parse", "text, sep?", func(g *interp.Goroutine, a []value.Value) value.Value {
				text := stringArg(g, "Csv.parse", a[0])
				sep := separator(g, "Csv.parse", a[1])
				r := csv.NewReader(strings.NewReader(text))
				r.Comma = sep
				r.FieldsPerRecord = -1
				records, err := r.ReadAll()
				if err != nil {
					return errResult(err)
				}
				rows := make([]value.Value, len(records))
				for i, rec := range records {
					fields := make([]value.Value, len(rec))
					for j, f := range rec {
						fields[j] = f
					}
					rows[i] = value.NewArray(fields)
				}
				return value.Ok(value.NewArray(rows))
			}),
			fn("Csv.stringify", "rows, sep?", func(g *interp.Goroutine, a []value.Value) value.Value {
				rows := arrayArg(g, "Csv.stringify", a[0]).Values()
				sep := separator(g, "Csv.stringify", a[1])
				var sb strings.Builder
				for _, row := range rows {
					for i, field := range stringsArg(g, "Csv.stringify", row) {
						if i > 0 {
							sb.WriteRune(sep)
						}
						csvField(&sb, field, sep)
					}
					sb.WriteByte('\n')
				}
				return sb.String()
			}),
		)
	})
}
