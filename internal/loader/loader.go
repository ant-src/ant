// Package loader reads a program's modules, parses and resolves them, and orders them for
// initialization (section 10).
package loader

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/resolve"
	"github.com/ant-src/ant/parser"
)

// Diagnostic is a load error with its position.
type Diagnostic struct {
	Path string
	Pos  ast.Pos
	Msg  string
}

func (d Diagnostic) String() string {
	if d.Pos.Line == 0 {
		return d.Path + ": " + d.Msg
	}
	return fmt.Sprintf("%s:%d:%d: %s", d.Path, d.Pos.Line, d.Pos.Col, d.Msg)
}

// Config says where sources come from.
type Config struct {
	// FS supplies sources by slash-separated path; nil means the operating system.
	FS fs.FS
	// Sources overrides the text of modules by path.
	Sources map[string]string
	// Mode is the parser mode for the root module; imports always parse as programs.
	Mode parser.Mode
	// RequireMain rejects a root module without func main.
	RequireMain bool
	// Extra names the prelude binds in this program, beyond the standard ones.
	Extra []string
}

// Module is one loaded module.
type Module struct {
	Path  string // canonical, the module's identity
	File  *ast.File
	Index int
}

// Program is a loaded module graph.
type Program struct {
	// Modules in initialization order: every module after the modules it imports, the root last.
	Modules []*Module
	Root    *Module
}

type loader struct {
	cfg     Config
	diags   []Diagnostic
	modules []*Module
	state   map[string]int // 1: in progress, 2: done
	index   map[string]*Module
}

// Load loads the module at root and everything it imports.
func Load(root string, cfg Config) (*Program, []Diagnostic) {
	l := &loader{cfg: cfg, state: map[string]int{}, index: map[string]*Module{}}
	key, display := l.canonical(root, "")
	m := l.visit(key, display, root, root, Diagnostic{}, true)
	if len(l.diags) > 0 {
		sort.SliceStable(l.diags, func(i, j int) bool {
			if l.diags[i].Path != l.diags[j].Path {
				return false
			}
			if l.diags[i].Pos.Line != l.diags[j].Pos.Line {
				return l.diags[i].Pos.Line < l.diags[j].Pos.Line
			}
			return l.diags[i].Pos.Col < l.diags[j].Pos.Col
		})
		return nil, l.diags
	}
	return &Program{Modules: l.modules, Root: m}, nil
}

// canonical returns a module's identity key and the path used in messages. raw is the path as
// written; importer is the raw path of the importing module, or "" for the root.
func (l *loader) canonical(raw, importer string) (key, display string) {
	joined := raw
	if importer != "" {
		if l.cfg.FS != nil || l.isSource(importer) {
			joined = path.Join(path.Dir(importer), raw)
		} else {
			joined = filepath.Join(filepath.Dir(importer), raw)
		}
	}
	if l.cfg.FS != nil || l.isSource(joined) {
		return path.Clean(joined), path.Clean(joined)
	}
	abs, err := filepath.Abs(joined)
	if err != nil {
		return joined, joined
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		abs = real
	}
	return abs, joined
}

func (l *loader) isSource(p string) bool {
	_, ok := l.cfg.Sources[path.Clean(p)]
	return ok
}

func (l *loader) read(key, raw string) (string, bool) {
	if src, ok := l.cfg.Sources[path.Clean(raw)]; ok {
		return src, true
	}
	if src, ok := l.cfg.Sources[key]; ok {
		return src, true
	}
	var data []byte
	var err error
	if l.cfg.FS != nil {
		data, err = fs.ReadFile(l.cfg.FS, key)
	} else {
		data, err = os.ReadFile(key)
	}
	if err != nil {
		return "", false
	}
	return string(data), true
}

func (l *loader) visit(key, display, raw, written string, at Diagnostic, root bool) *Module {
	switch l.state[key] {
	case 1:
		l.diags = append(l.diags, Diagnostic{Path: at.Path, Pos: at.Pos, Msg: fmt.Sprintf("import cycle through %s", display)})
		return nil
	case 2:
		return l.index[key]
	}
	src, ok := l.read(key, raw)
	if !ok {
		if root {
			l.diags = append(l.diags, Diagnostic{Path: display, Msg: "cannot read the module"})
		} else {
			l.diags = append(l.diags, Diagnostic{Path: at.Path, Pos: at.Pos, Msg: fmt.Sprintf("cannot find module %q", written)})
		}
		return nil
	}
	mode := parser.Program
	if root {
		mode = l.cfg.Mode
	}
	f, err := parser.Parse(display, src, mode)
	if err != nil {
		pe := err.(*parser.Error)
		l.diags = append(l.diags, Diagnostic{Path: display, Pos: pe.Pos, Msg: pe.Msg})
		return nil
	}
	for _, e := range resolve.Resolve(f, resolve.Options{Extra: l.cfg.Extra}) {
		l.diags = append(l.diags, Diagnostic{Path: display, Pos: e.Pos, Msg: e.Msg})
	}
	if root && l.cfg.RequireMain && !hasMain(f) {
		l.diags = append(l.diags, Diagnostic{Path: display, Msg: "no func main in the module"})
	}
	l.state[key] = 1
	m := &Module{Path: key, File: f}
	for _, imp := range f.Imports {
		childKey, childDisplay := l.canonical(imp.Path, raw)
		child := l.visit(childKey, childDisplay, path.Join(path.Dir(raw), imp.Path), imp.Path, Diagnostic{Path: display, Pos: imp.Pos}, false)
		if child != nil {
			imp.Module = child.Index
		}
	}
	l.state[key] = 2
	m.Index = len(l.modules)
	l.modules = append(l.modules, m)
	l.index[key] = m
	return m
}

func hasMain(f *ast.File) bool {
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Owner == "" && fd.Name == "main" {
			return true
		}
	}
	return false
}
