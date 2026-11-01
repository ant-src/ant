package lsp

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/parser"
)

// workspaceFile is a .ant file of the workspace, parsed: an open document as it is, or a file on
// disk as it was when last modified.
type workspaceFile struct {
	doc     *document
	file    *ast.File
	modTime int64
	size    int64
	version int
}

// workspaceFiles lists the parsed .ant files of the workspace: every open document, and every
// other file under the workspace folders (walk).
func (s *Server) workspaceFiles() []*workspaceFile {
	if s.files == nil {
		s.files = map[string]*workspaceFile{}
	}
	open := s.byPath()
	var out []*workspaceFile
	for path, d := range open {
		f := s.files[path]
		if f == nil || f.doc != d || f.version != d.version || f.version == 0 {
			parsed, _, _ := parser.ParseRecover(path, d.text)
			f = &workspaceFile{doc: d, file: parsed, version: d.version}
			s.files[path] = f
		}
		out = append(out, f)
	}
	for path := range s.files {
		if open[path] == nil {
			delete(s.files, path)
		}
	}
	for _, f := range s.walk() {
		if open[f.doc.path] == nil {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].doc.path < out[j].doc.path })
	return out
}

// walk lists the parsed .ant files under the workspace folders, skipping entries that start
// with a dot and node_modules. While the client watches the files, the last walk stands until
// it reports a change, so that completion, which asks again with each character typed, does not
// walk the folders each time.
func (s *Server) walk() []*workspaceFile {
	if s.watching && s.walked != nil && s.walkedGen == s.diskGen {
		return s.walked
	}
	if s.disk == nil {
		s.disk = map[string]*workspaceFile{}
	}
	seen := map[string]bool{}
	out := []*workspaceFile{}
	for _, root := range s.roots {
		filepath.WalkDir(root, func(p string, e os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if e.IsDir() {
				if p != root && (strings.HasPrefix(e.Name(), ".") || e.Name() == "node_modules") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(p, ".ant") || seen[p] {
				return nil
			}
			seen[p] = true
			info, err := e.Info()
			if err != nil {
				return nil
			}
			// parsed again only when its size or modification time changed
			f := s.disk[p]
			if f == nil || f.modTime != info.ModTime().UnixNano() || f.size != info.Size() {
				src, err := os.ReadFile(p)
				if err != nil {
					return nil
				}
				parsed, _, _ := parser.ParseRecover(p, string(src))
				f = &workspaceFile{doc: newDocument(fileURI(p), 0, string(src)), file: parsed, modTime: info.ModTime().UnixNano(), size: info.Size()}
				s.disk[p] = f
			}
			out = append(out, f)
			return nil
		})
	}
	for path := range s.disk {
		if !seen[path] {
			delete(s.disk, path)
		}
	}
	s.walked, s.walkedGen = out, s.diskGen
	return out
}
