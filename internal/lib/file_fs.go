package lib

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/temporal"
	"github.com/ant-src/ant/value"
)

// The File functions of the file system beyond reads and writes (stdlib.md, File: Modes,
// Temporary files, Not replacing, Copies, Links, Times, Locks).

// New files and directories are created as the shell creates them: 0o666 and 0o777, less the
// umask.
const (
	newFileMode fs.FileMode = 0o666
	newDirMode  fs.FileMode = 0o777
)

// maxMode is the largest mode File.chmod takes: the permission bits and setuid, setgid and sticky.
const maxMode = 0o7777

// boolOr is an optional bool argument, def when the call omitted it.
func boolOr(v value.Value, def bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

// antMode is File.Info.mode: Go's permission bits and its setuid, setgid and sticky flags, as the
// Unix numbers 0o4000, 0o2000 and 0o1000.
func antMode(m fs.FileMode) int64 {
	v := int64(m.Perm())
	if m&fs.ModeSetuid != 0 {
		v |= 0o4000
	}
	if m&fs.ModeSetgid != 0 {
		v |= 0o2000
	}
	if m&fs.ModeSticky != 0 {
		v |= 0o1000
	}
	return v
}

// goMode is the Go mode of an Ant mode in 0 ..= 0o7777.
func goMode(v int64) fs.FileMode {
	m := fs.FileMode(v & 0o777)
	if v&0o4000 != 0 {
		m |= fs.ModeSetuid
	}
	if v&0o2000 != 0 {
		m |= fs.ModeSetgid
	}
	if v&0o1000 != 0 {
		m |= fs.ModeSticky
	}
	return m
}

// statFile is File.stat: Info for path, of the link itself when follow is false.
func statFile(path string, follow bool) value.Value {
	stat := os.Stat
	if !follow {
		stat = os.Lstat
	}
	fi, err := stat(path)
	if err != nil {
		return errResult(err)
	}
	modified, err := temporal.NewInstant(fi.ModTime().Unix(), int64(fi.ModTime().Nanosecond()))
	if err != nil {
		return errResult(err)
	}
	return value.Ok(value.NewRecord(InfoTag, []value.Value{
		value.Text(fi.Name()), fi.Size(), fi.IsDir(), fi.Mode()&fs.ModeSymlink != 0, modified, antMode(fi.Mode()),
	}))
}

// chmod is File.chmod: the mode set exactly, following a link.
func chmod(path string, mode int64) value.Value {
	if mode < 0 || mode > maxMode {
		return errResult(fmt.Errorf("chmod %s: mode %O is outside 0o0 ..= 0o7777", path, mode))
	}
	return okVoid(os.Chmod(path, goMode(mode)))
}

// createFlags are the flags of a file opened for writing from its start: truncated when it
// exists, or with replace false created only when nothing is at the path, a link included.
func createFlags(replace bool) int {
	if replace {
		return os.O_CREATE | os.O_TRUNC
	}
	return os.O_CREATE | os.O_EXCL
}

// writeFile is File.write and File.writeBytes: os.WriteFile, which keeps an existing file's mode,
// with replace false creating the file only when nothing is at path.
func writeFile(path string, data []byte, replace bool) value.Value {
	f, err := os.OpenFile(path, os.O_WRONLY|createFlags(replace), newFileMode)
	if err != nil {
		return errResult(err)
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return okVoid(err)
}

// touch is File.touch: an empty file created when nothing is at path, never truncated, and its
// modified and access times set to t.
func touch(path string, t time.Time) value.Value {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, newFileMode)
	if err != nil {
		// A directory, or a file that cannot be opened for writing, still has times to set.
		if _, serr := os.Stat(path); serr != nil {
			return errResult(err)
		}
	} else if err := f.Close(); err != nil {
		return errResult(err)
	}
	return okVoid(os.Chtimes(path, t, t))
}

// tempFile is File.tempFile: os.CreateTemp, with the mode set to 0o600 exactly, whatever the
// umask.
func tempFile(pattern, dir string) value.Value {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return errResult(err)
	}
	name := f.Name()
	err = f.Chmod(0o600)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.Remove(name)
		return errResult(err)
	}
	return value.Ok(value.Text(name))
}

// tempDir is File.tempDir: os.MkdirTemp, with the mode set to 0o700 exactly.
func tempDir(pattern, dir string) value.Value {
	name, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return errResult(err)
	}
	if err := os.Chmod(name, 0o700); err != nil {
		os.Remove(name)
		return errResult(err)
	}
	return value.Ok(value.Text(name))
}

// copyFile is File.copy: from's contents into to, then from's mode, exactly. It is Err before
// changing anything when to is a directory or the same file as from.
func copyFile(from, to string, replace bool) value.Value {
	src, err := os.Open(from)
	if err != nil {
		return errResult(err)
	}
	defer src.Close()
	sfi, err := src.Stat()
	if err != nil {
		return errResult(err)
	}
	if sfi.IsDir() {
		return errResult(fmt.Errorf("copy %s: is a directory; File.copyAll copies a tree", from))
	}
	if dfi, err := os.Stat(to); err == nil {
		if dfi.IsDir() {
			return errResult(fmt.Errorf("copy %s to %s: %s is a directory", from, to, to))
		}
		if os.SameFile(sfi, dfi) {
			return errResult(fmt.Errorf("copy %s to %s: the same file", from, to))
		}
	}
	return okVoid(copyContents(src, sfi, to, createFlags(replace)))
}

// copyContents writes src to the file at to, opened with flags, and gives it src's mode.
func copyContents(src *os.File, sfi fs.FileInfo, to string, flags int) error {
	dst, err := os.OpenFile(to, os.O_WRONLY|flags, newFileMode)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	if err == nil {
		err = dst.Chmod(sfi.Mode() & (fs.ModePerm | fs.ModeSetuid | fs.ModeSetgid | fs.ModeSticky))
	}
	if err1 := dst.Close(); err == nil {
		err = err1
	}
	return err
}

// resolved is path with every link resolved, for a path that may not exist yet: its nearest
// existing ancestor resolved, and the rest joined on.
func resolved(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var rest []string
	for {
		real, err := filepath.EvalSymlinks(abs)
		if err == nil {
			slices.Reverse(rest)
			return filepath.Join(append([]string{real}, rest...)...), nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", err
		}
		rest = append(rest, filepath.Base(abs))
		abs = parent
	}
}

// within reports whether path is dir or lies below it; both are resolved.
func within(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// copyAll is File.copyAll: a file, or a tree with its modes and its links as links, to a to that
// does not exist or is an empty directory. Every check comes before the first change.
func copyAll(from, to string) value.Value {
	sfi, err := os.Stat(from)
	if err != nil {
		return errResult(err)
	}
	existing := false
	if dfi, err := os.Lstat(to); err == nil {
		empty := false
		if sfi.IsDir() && dfi.IsDir() {
			entries, err := os.ReadDir(to)
			if err != nil {
				return errResult(err)
			}
			empty = len(entries) == 0
		}
		if !empty {
			return errResult(fmt.Errorf("copy %s to %s: %s exists", from, to, to))
		}
		existing = true
	} else if !errors.Is(err, fs.ErrNotExist) {
		return errResult(err)
	}
	if !sfi.IsDir() {
		src, err := os.Open(from)
		if err != nil {
			return errResult(err)
		}
		defer src.Close()
		return okVoid(copyContents(src, sfi, to, os.O_CREATE|os.O_EXCL))
	}
	root, err := filepath.EvalSymlinks(from)
	if err != nil {
		return errResult(err)
	}
	if root, err = filepath.Abs(root); err != nil {
		return errResult(err)
	}
	dest, err := resolved(to)
	if err != nil {
		return errResult(err)
	}
	if within(dest, root) {
		return errResult(fmt.Errorf("copy %s to %s: %s lies inside %s", from, to, to, from))
	}
	// A directory's mode is set after its entries are copied, so a read-only one can be filled.
	type dirMode struct {
		path string
		mode fs.FileMode
	}
	var modes []dirMode
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		target := filepath.Join(to, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			if rel == "." && existing {
				return nil
			}
			if err := os.Mkdir(target, 0o700); err != nil {
				return err
			}
			modes = append(modes, dirMode{target, info.Mode() & (fs.ModePerm | fs.ModeSetuid | fs.ModeSetgid | fs.ModeSticky)})
		case info.Mode()&fs.ModeSymlink != 0:
			link, err := os.Readlink(p)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		case info.Mode().IsRegular():
			src, err := os.Open(p)
			if err != nil {
				return err
			}
			defer src.Close()
			return copyContents(src, info, target, os.O_CREATE|os.O_EXCL)
		default:
			return fmt.Errorf("copy %s: not a file, a directory or a link", p)
		}
		return nil
	})
	for i := len(modes) - 1; i >= 0; i-- {
		if err1 := os.Chmod(modes[i].path, modes[i].mode); err == nil {
			err = err1
		}
	}
	return okVoid(err)
}

// realPath is File.realPath: absolute, clean, every link resolved.
func realPath(path string) value.Value {
	abs, err := filepath.Abs(path)
	if err != nil {
		return errResult(err)
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return errResult(err)
	}
	return value.Ok(value.Text(real))
}

// lockRetry is how long File.lock waits between tries of a lock another handle holds: flock
// cannot be interrupted, so the wait is a loop of non-blocking tries that a context can end.
const (
	lockRetryMin = time.Millisecond
	lockRetryMax = 50 * time.Millisecond
)

// openLock opens the file at path for File.lock and File.tryLock: for reading and writing, created
// when missing, never truncated.
func openLock(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE, newFileMode)
}

// lockFile is File.lock with wait true and File.tryLock with wait false: a handle holding the
// file's exclusive lock, or null from tryLock when another handle holds it.
func lockFile(ctx context.Context, path string, wait bool) value.Value {
	f, err := openLock(path)
	if err != nil {
		return errResult(err)
	}
	delay := lockRetryMin
	for {
		held, err := tryLockFile(f)
		if err != nil {
			f.Close()
			return errResult(fmt.Errorf("lock %s: %w", path, err))
		}
		if !held {
			return value.Ok(newHandle(f, f, f))
		}
		if !wait {
			f.Close()
			return value.Ok(value.Null)
		}
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			f.Close()
			return gaveUp(ctx)
		}
		delay = min(2*delay, lockRetryMax)
	}
}

// handleIsTerminal is File.isTerminal: whether the file under h, or the stream a standard handle
// stands for, is a terminal.
func handleIsTerminal(h *Handle) bool {
	h = h.sink()
	var f *os.File
	switch {
	case h.std && h.childIn != nil:
		f, _ = h.childIn.(*os.File)
	case h.std:
		f, _ = h.childOut.(*os.File)
	default:
		f = h.file()
	}
	return f != nil && fileIsTerminal(f)
}

func init() {
	fileFS = []*interp.Builtin{
		fn("File.chmod", "path, mode", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			return chmod(stringArg(g, who, a[0]), intArg(g, who, a[1]))
		}),
		fn("File.touch", "path, time?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			t := time.Now()
			if given(a[1]) {
				t = instantArg(g, who, a[1]).GoTime()
			}
			return touch(stringArg(g, who, a[0]), t)
		}),
		fn("File.copy", "from, to, replace?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			return copyFile(stringArg(g, who, a[0]), stringArg(g, who, a[1]), boolOr(a[2], true))
		}),
		fn("File.copyAll", "from, to", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			return copyAll(stringArg(g, who, a[0]), stringArg(g, who, a[1]))
		}),
		fn("File.tempFile", "pattern, dir?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			dir := ""
			if given(a[1]) {
				dir = stringArg(g, who, a[1])
			}
			return tempFile(stringArg(g, who, a[0]), dir)
		}),
		fn("File.symlink", "target, link", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			return okVoid(os.Symlink(stringArg(g, who, a[0]), stringArg(g, who, a[1])))
		}),
		fn("File.readLink", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			target, err := os.Readlink(stringArg(g, who, a[0]))
			if err != nil {
				return errResult(err)
			}
			return value.Ok(value.Text(target))
		}),
		fn("File.realPath", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			return realPath(stringArg(g, who, a[0]))
		}),
		fn("File.lock", "path, ctx?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			ctx := waitContext(g, who, a[1])
			defer g.AfterWait()
			return lockFile(ctx, stringArg(g, who, a[0]), true)
		}),
		fn("File.tryLock", "path", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			return lockFile(g.Context(), stringArg(g, who, a[0]), false)
		}),
		fn("File.isTerminal", "f", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
			return handleIsTerminal(handleArg(g, who, a[0]))
		}),
	}
}

// fileFS are the builtins above, which the File namespace appends.
var fileFS []*interp.Builtin
