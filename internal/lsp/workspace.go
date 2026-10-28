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
// file under the workspace folders, skipping entries that start with a dot and node_modules.
// Files on disk are parsed again only when their size or modification time changed.
func (s *Server) workspaceFiles() []*workspaceFile {
	if s.files == nil {
		s.files = map[string]*workspaceFile{}
	}
	seen := map[string]bool{}
	var out []*workspaceFile
	for path, d := range s.byPath() {
		seen[path] = true
		f := s.files[path]
		if f == nil || f.doc != d || f.version != d.version || f.version == 0 {
			parsed, _, _ := parser.ParseRecover(path, d.text)
			f = &workspaceFile{doc: d, file: parsed, version: d.version}
			s.files[path] = f
		}
		out = append(out, f)
	}
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
			f := s.files[p]
			if f == nil || f.doc.version != 0 || f.modTime != info.ModTime().UnixNano() || f.size != info.Size() {
				src, err := os.ReadFile(p)
				if err != nil {
					return nil
				}
				parsed, _, _ := parser.ParseRecover(p, string(src))
				f = &workspaceFile{doc: newDocument(fileURI(p), 0, string(src)), file: parsed, modTime: info.ModTime().UnixNano(), size: info.Size()}
				s.files[p] = f
			}
			out = append(out, f)
			return nil
		})
	}
	for path := range s.files {
		if !seen[path] {
			delete(s.files, path)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].doc.path < out[j].doc.path })
	return out
}
