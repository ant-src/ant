package lib

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"unicode/utf8"
)

// Globs (stdlib.md, Path and File: Globs): the syntax of Go's filepath.Match per element, with
// ** across elements, [!...] for [^...], and the shell's rule that a wildcard does not match a
// leading dot. Path.match and File.glob share the parsed pattern.

// escapes reports whether \ escapes in a pattern: on Unix, where it is not a separator.
const escapes = filepath.Separator == '/'

// globPattern is a parsed pattern: its elements, and whether it ended in a separator, which
// asks for directories only.
type globPattern struct {
	elems   []globElem
	dirOnly bool
}

// globElem is one element of a pattern. A literal element has no wildcards and is lit, with its
// escapes removed; ** is star; any other element is its tokens.
type globElem struct {
	literal bool
	lit     string
	star    bool
	toks    []globTok
}

type tokKind uint8

const (
	tokRune tokKind = iota // r itself
	tokAny                 // ?
	tokStar                // *
	tokSet                 // [...]
)

type globTok struct {
	kind   tokKind
	r      rune
	ranges []rune // lo, hi pairs, for a set
	negate bool
}

// splitElems splits p at the host's separators, keeping empty elements, so "/a" gives "", "a".
func splitElems(p string) []string {
	var out []string
	start := 0
	for i := 0; i < len(p); i++ {
		if os.IsPathSeparator(p[i]) {
			out = append(out, p[start:i])
			start = i + 1
		}
	}
	return append(out, p[start:])
}

// parseGlob parses pattern, or says why it is malformed.
func parseGlob(pattern string) (*globPattern, error) {
	parts := splitElems(pattern)
	g := &globPattern{}
	if len(parts) > 1 && parts[len(parts)-1] == "" {
		g.dirOnly = true
		parts = parts[:len(parts)-1]
	}
	for _, p := range parts {
		e, err := parseElem(p)
		if err != nil {
			return nil, fmt.Errorf("malformed pattern %s: %w", pattern, err)
		}
		g.elems = append(g.elems, e)
	}
	return g, nil
}

func parseElem(p string) (globElem, error) {
	if p == "**" {
		return globElem{star: true}, nil
	}
	var toks []globTok
	literal := true
	for i := 0; i < len(p); {
		r, n := utf8.DecodeRuneInString(p[i:])
		switch {
		case r == '*':
			literal = false
			if len(toks) == 0 || toks[len(toks)-1].kind != tokStar {
				toks = append(toks, globTok{kind: tokStar})
			}
			i += n
		case r == '?':
			literal = false
			toks = append(toks, globTok{kind: tokAny})
			i += n
		case r == '[':
			literal = false
			t, w, err := parseSet(p[i+1:])
			if err != nil {
				return globElem{}, err
			}
			toks = append(toks, t)
			i += 1 + w
		case r == '\\' && escapes:
			if i+n >= len(p) {
				return globElem{}, errors.New(`\ at the end of an element`)
			}
			r2, n2 := utf8.DecodeRuneInString(p[i+n:])
			toks = append(toks, globTok{kind: tokRune, r: r2})
			i += n + n2
		default:
			toks = append(toks, globTok{kind: tokRune, r: r})
			i += n
		}
	}
	if literal {
		var b strings.Builder
		for _, t := range toks {
			b.WriteRune(t.r)
		}
		return globElem{literal: true, lit: b.String()}, nil
	}
	return globElem{toks: toks}, nil
}

// parseSet parses a set after its [, returning it and the bytes it took, the ] included.
func parseSet(s string) (globTok, int, error) {
	t := globTok{kind: tokSet}
	i := 0
	if i < len(s) && (s[i] == '^' || s[i] == '!') {
		t.negate = true
		i++
	}
	for {
		if i >= len(s) {
			return t, 0, errors.New("[ without ]")
		}
		if s[i] == ']' && len(t.ranges) > 0 {
			return t, i + 1, nil
		}
		lo, w, err := setRune(s[i:])
		if err != nil {
			return t, 0, err
		}
		i += w
		hi := lo
		if i < len(s) && s[i] == '-' {
			hi, w, err = setRune(s[i+1:])
			if err != nil {
				return t, 0, err
			}
			i += 1 + w
			if hi < lo {
				return t, 0, fmt.Errorf("range %c-%c ends before it starts", lo, hi)
			}
		}
		t.ranges = append(t.ranges, lo, hi)
	}
}

// setRune reads one code point of a set, escaped or not; a bare - or ] is malformed, as in Go.
func setRune(s string) (rune, int, error) {
	if s == "" {
		return 0, 0, errors.New("[ without ]")
	}
	if s[0] == '-' || s[0] == ']' {
		if s[0] == ']' {
			return 0, 0, errors.New("empty set")
		}
		return 0, 0, errors.New("- outside a range; escape it")
	}
	if s[0] == '\\' && escapes {
		if len(s) == 1 {
			return 0, 0, errors.New("[ without ]")
		}
		r, n := utf8.DecodeRuneInString(s[1:])
		return r, 1 + n, nil
	}
	r, n := utf8.DecodeRuneInString(s)
	return r, n, nil
}

// matchName reports whether one element of a path matches e, under the dot rule.
func (e *globElem) matchName(name string, hidden bool) bool {
	if e.literal {
		return e.lit == name
	}
	if e.star {
		return name != "." && name != ".." && (hidden || !strings.HasPrefix(name, "."))
	}
	if name == "." || name == ".." {
		return false
	}
	if strings.HasPrefix(name, ".") && !hidden && (e.toks[0].kind != tokRune || e.toks[0].r != '.') {
		return false
	}
	return matchToks(e.toks, name)
}

// matchToks matches name against toks, backtracking only to the last *, which is enough since
// every other token takes exactly one code point.
func matchToks(toks []globTok, name string) bool {
	ti, ni := 0, 0
	starT, starN := -1, 0
	for ni < len(name) || ti < len(toks) {
		if ti < len(toks) {
			t := toks[ti]
			if t.kind == tokStar {
				starT, starN = ti, ni
				ti++
				continue
			}
			if ni < len(name) {
				r, n := utf8.DecodeRuneInString(name[ni:])
				if t.matches(r) {
					ti++
					ni += n
					continue
				}
			}
		}
		if starT < 0 || starN >= len(name) {
			return false
		}
		_, n := utf8.DecodeRuneInString(name[starN:])
		starN += n
		ti, ni = starT+1, starN
	}
	return true
}

func (t globTok) matches(r rune) bool {
	switch t.kind {
	case tokRune:
		return t.r == r
	case tokAny:
		return true
	case tokSet:
		in := false
		for i := 0; i < len(t.ranges); i += 2 {
			if t.ranges[i] <= r && r <= t.ranges[i+1] {
				in = true
				break
			}
		}
		return in != t.negate
	}
	return false
}

// matchPath is Path.match: the whole of path, element by element, on the text alone.
func (g *globPattern) matchPath(path string, hidden bool) bool {
	parts := splitElems(path)
	if g.dirOnly {
		if len(parts) < 2 || parts[len(parts)-1] != "" {
			return false
		}
		parts = parts[:len(parts)-1]
	} else if len(parts) > 1 && parts[len(parts)-1] == "" {
		return false
	}
	return matchElems(g.elems, parts, hidden)
}

func matchElems(elems []globElem, parts []string, hidden bool) bool {
	if len(elems) == 0 {
		return len(parts) == 0
	}
	e := &elems[0]
	if !e.star {
		return len(parts) > 0 && e.matchName(parts[0], hidden) && matchElems(elems[1:], parts[1:], hidden)
	}
	// ** takes zero or more elements, one or more as the last; never one that begins with a dot.
	min := 0
	if len(elems) == 1 {
		min = 1
	}
	for k := 0; k <= len(parts); k++ {
		if k >= min && matchElems(elems[1:], parts[k:], hidden) {
			return true
		}
		if k < len(parts) && !e.matchName(parts[k], hidden) {
			return false
		}
	}
	return false
}

// glob is File.glob: the paths below dir that match pattern.
func glob(pattern, dir string, hidden bool) ([]string, error) {
	g, err := parseGlob(pattern)
	if err != nil {
		return nil, err
	}
	if pattern == "" {
		return nil, nil
	}
	w := &globWalk{g: g, hidden: hidden, seen: map[string]bool{}}
	if filepath.IsAbs(pattern) {
		// The volume and the root are where the walk starts, and the rest is matched below them.
		root := pattern[:len(filepath.VolumeName(pattern))+1]
		if w.g, err = parseGlob(pattern[len(root):]); err != nil {
			return nil, err
		}
		if err := w.match(root, root, w.g.elems); err != nil {
			return nil, err
		}
	} else {
		fi, err := os.Stat(dir)
		if err != nil {
			return nil, err
		}
		if !fi.IsDir() {
			return nil, errors.New(dir + " is not a directory")
		}
		if err := w.match(dir, "", g.elems); err != nil {
			return nil, err
		}
	}
	slices.SortFunc(w.out, compareElems)
	return w.out, nil
}

// globWalk reads the directories a pattern leads into. fs is the path on disk, out the path to
// report: relative to dir, or absolute.
type globWalk struct {
	g      *globPattern
	hidden bool
	out    []string
	seen   map[string]bool
}

func (w *globWalk) emit(fsPath, out string) error {
	if w.g.dirOnly {
		fi, err := os.Stat(fsPath)
		if err != nil || !fi.IsDir() {
			return nil
		}
	}
	out = filepath.Clean(out)
	if w.g.dirOnly && !strings.HasSuffix(out, string(filepath.Separator)) {
		out += string(filepath.Separator)
	}
	if !w.seen[out] {
		w.seen[out] = true
		w.out = append(w.out, out)
	}
	return nil
}

func join(a, b string) string {
	if a == "" {
		return b
	}
	return filepath.Join(a, b)
}

// missing reports an error that means nothing is there: no match, not a failure.
func missing(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR)
}

// readDir lists dir, or nothing when it is missing or not a directory.
func readDir(dir string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil && missing(err) {
		return nil, nil
	}
	return entries, err
}

// match matches elems below fsPath, whose reported path is out.
func (w *globWalk) match(fsPath, out string, elems []globElem) error {
	if len(elems) == 0 {
		return w.emit(fsPath, out)
	}
	e := &elems[0]
	switch {
	case e.literal:
		p := join(fsPath, e.lit)
		if _, err := os.Lstat(p); err != nil {
			if missing(err) {
				return nil
			}
			return err
		}
		return w.match(p, join(out, e.lit), elems[1:])
	case e.star:
		return w.star(fsPath, out, elems)
	}
	entries, err := readDir(fsPath)
	if err != nil {
		return err
	}
	for _, d := range entries {
		if !e.matchName(d.Name(), w.hidden) {
			continue
		}
		// Only a directory, or a link that may lead to one, has anything below it.
		if len(elems) > 1 && !d.IsDir() && d.Type()&fs.ModeSymlink == 0 {
			continue
		}
		if err := w.match(join(fsPath, d.Name()), join(out, d.Name()), elems[1:]); err != nil {
			return err
		}
	}
	return nil
}

// star matches ** at elems[0] below fsPath: zero or more elements, one or more as the last,
// descending only into directories, never through links.
func (w *globWalk) star(fsPath, out string, elems []globElem) error {
	last := len(elems) == 1
	if !last {
		if err := w.match(fsPath, out, elems[1:]); err != nil {
			return err
		}
	}
	entries, err := readDir(fsPath)
	if err != nil {
		return err
	}
	for _, d := range entries {
		if !elems[0].matchName(d.Name(), w.hidden) {
			continue
		}
		p, o := join(fsPath, d.Name()), join(out, d.Name())
		if last {
			if err := w.emit(p, o); err != nil {
				return err
			}
		}
		if d.IsDir() {
			if err := w.star(p, o, elems); err != nil {
				return err
			}
		}
	}
	return nil
}

// compareElems orders paths element by element, as File.walk does.
func compareElems(a, b string) int {
	return slices.Compare(splitElems(a), splitElems(b))
}
