package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// sources says which files a command takes from the paths on its command line.
type sources struct {
	// keep accepts the name of a file under a directory.
	keep func(name string) bool
	// named, when set, refuses a file named directly.
	named func(path string) error
	// what names the files for the error about a directory with none, as in "no .ant files
	// under dir"; when empty, such a directory is no error.
	what string
}

// files lists the files the paths name: a file itself, or every file under a directory that keep
// accepts, in lexical order, leaving out directories whose names begin with a dot. A directory
// named through a symbolic link is walked through its target, and its files are named under the
// link as given. With no path, it is ".".
func (s sources) files(paths []string) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if s.named != nil {
				if err := s.named(p); err != nil {
					return nil, err
				}
			}
			files = append(files, p)
			continue
		}
		found, err := s.dir(p)
		if err != nil {
			return nil, err
		}
		if len(found) == 0 && s.what != "" {
			return nil, fmt.Errorf("no %s under %s", s.what, p)
		}
		files = append(files, found...)
	}
	return files, nil
}

func (s sources) dir(root string) ([]string, error) {
	target, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(target, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			if rel != "." && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if s.keep(d.Name()) {
			files = append(files, filepath.Join(root, rel))
		}
		return nil
	})
	return files, err
}
