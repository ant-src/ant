package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/ant-src/ant/docs"
	"github.com/ant-src/ant/internal/codes"
)

// explain writes the entry of code from ## Codes of the ant check page: its rule and examples,
// with the code of each example indented by four spaces, as ant doc prints examples. It reports
// whether code has an entry.
func explain(w io.Writer, code string) bool {
	if !codes.Known(codes.Code(code)) {
		return false
	}
	entry, ok := codeEntry(docs.Check, code)
	if !ok {
		return false
	}
	out := []string{code, ""}
	inCode := false
	for _, line := range entry {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "```"):
			if name, ok := strings.CutPrefix(trimmed, "```ant ["); ok && !inCode {
				out = append(out, "    // "+strings.TrimSuffix(name, "]"))
			}
			inCode = !inCode
		case strings.HasPrefix(trimmed, "<!--"):
		case trimmed == "":
			out = append(out, "")
		case inCode:
			out = append(out, "    "+line)
		default:
			out = append(out, unlink(line))
		}
	}
	// one blank line between paragraphs and examples, and none at the end
	prev := ""
	for i, l := range out {
		if l == "" && (prev == "" || i == len(out)-1) && i > 1 {
			continue
		}
		fmt.Fprintln(w, l)
		prev = l
	}
	return true
}

// codeEntry returns the lines of the entry ### `code` under ## Codes of page, without its heading.
func codeEntry(page, code string) ([]string, bool) {
	lines := strings.Split(page, "\n")
	inCodes := false
	for i, l := range lines {
		switch {
		case strings.HasPrefix(l, "## "):
			inCodes = l == "## Codes"
		case inCodes && l == "### `"+code+"`":
			end := i + 1
			for end < len(lines) && !strings.HasPrefix(lines[end], "## ") && !strings.HasPrefix(lines[end], "### ") {
				end++
			}
			return lines[i+1 : end], true
		}
	}
	return nil, false
}

// unlink drops the targets of the Markdown links in line, keeping their text.
func unlink(line string) string {
	var b strings.Builder
	for {
		open := strings.Index(line, "](")
		start := strings.LastIndex(line[:max(open, 0)], "[")
		end := strings.Index(line[max(open, 0):], ")")
		if open < 0 || start < 0 || end < 0 {
			b.WriteString(line)
			return b.String()
		}
		b.WriteString(line[:start])
		b.WriteString(line[start+1 : open])
		line = line[open+end+1:]
	}
}
