package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ant-src/ant/internal/format"
)

// fmtCommand is ant fmt (spec/fmt.md section 1): it rewrites each named file and each .ant file
// under each named directory, or with --check only lists those that would change. The status is
// 2 when any file could not be read, parsed, or written, else 1 when --check found a file to
// change, else 0.
func fmtCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	check := len(args) > 0 && args[0] == "--check"
	if check {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && arg != "-" {
			fmt.Fprintln(stderr, usage)
			return 2
		}
	}
	f := &formatter{check: check, stdout: stdout, stderr: stderr}
	for _, arg := range args {
		switch info, err := os.Stat(arg); {
		case arg == "-":
			f.stdin(stdin)
		case err != nil:
			f.fail(err)
		case info.IsDir():
			f.dir(arg)
		default:
			f.file(arg)
		}
	}
	return f.status
}

type formatter struct {
	check          bool
	stdout, stderr io.Writer
	status         int
}

func (f *formatter) fail(err error) {
	fmt.Fprintln(f.stderr, err)
	f.status = 2
}

// unformatted reports a file --check would change.
func (f *formatter) unformatted(name string) {
	fmt.Fprintln(f.stdout, name)
	if f.status == 0 {
		f.status = 1
	}
}

// format formats src, reporting a syntax error under name.
func (f *formatter) format(name string, src []byte) ([]byte, bool) {
	out, err := format.Source(name, src)
	if err != nil {
		f.fail(fmt.Errorf("%s:%w", name, err))
		return nil, false
	}
	return out, true
}

func (f *formatter) stdin(r io.Reader) {
	src, err := io.ReadAll(r)
	if err != nil {
		f.fail(err)
		return
	}
	out, ok := f.format("<stdin>", src)
	switch {
	case !ok:
	case f.check:
		if !bytes.Equal(out, src) {
			f.unformatted("<stdin>")
		}
	default:
		if _, err := f.stdout.Write(out); err != nil {
			f.fail(err)
		}
	}
}

// dir formats the .ant files under root in lexical order, skipping entries whose names start
// with a dot and not following symbolic links. A root that is itself a link is walked through
// its target, which WalkDir would not enter, and its files are named under root as given.
func (f *formatter) dir(root string) {
	target, err := filepath.EvalSymlinks(root)
	if err != nil {
		f.fail(err)
		return
	}
	err = filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			f.fail(err)
			return nil
		}
		rel, err := filepath.Rel(target, path)
		if err != nil {
			return err
		}
		switch name := filepath.Join(root, rel); {
		case rel != "." && strings.HasPrefix(d.Name(), "."):
			if d.IsDir() {
				return filepath.SkipDir
			}
		case d.Type().IsRegular() && strings.HasSuffix(name, ".ant"):
			f.file(name)
		}
		return nil
	})
	if err != nil {
		f.fail(err)
	}
}

// file formats one file in place, through a symbolic link to its target, writing only when the
// contents change and replacing the old contents in one rename.
func (f *formatter) file(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		f.fail(err)
		return
	}
	out, ok := f.format(path, src)
	if !ok || bytes.Equal(out, src) {
		return
	}
	if f.check {
		f.unformatted(path)
		return
	}
	if err := replace(path, out); err != nil {
		f.fail(err)
	}
}

func replace(path string, contents []byte) error {
	target, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	// The rename would replace a file the user may not write, so ask for write access first.
	w, err := os.OpenFile(target, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	w.Close()
	tmp, err := os.CreateTemp(filepath.Dir(target), ".ant-fmt-*")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(contents)
	cerr := tmp.Close()
	if err := errors.Join(werr, cerr, os.Chmod(tmp.Name(), info.Mode().Perm())); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), target); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}
