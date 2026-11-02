//go:build unix

package lib

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestGlobUnreadable: a directory the pattern leads into that cannot be read is Err, as for
// File.walk, while one the pattern does not lead into is never read (stdlib.md, File: Globs).
func TestGlobUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root reads every directory")
	}
	dir := t.TempDir()
	locked := filepath.Join(dir, "locked")
	for _, p := range []string{filepath.Join(locked, "a.txt"), filepath.Join(dir, "open", "b.txt")} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(locked, 0o755)

	if _, err := glob("**/*.txt", dir, false); err == nil {
		t.Error("glob ** over an unreadable directory = Ok, want Err")
	}
	if _, err := glob("*/*.txt", dir, false); err == nil {
		t.Error("glob */*.txt over an unreadable directory = Ok, want Err")
	}
	got, err := glob("open/*.txt", dir, false)
	if err != nil || !slices.Equal(got, []string{"open/b.txt"}) {
		t.Errorf("glob open/*.txt = %v, %v; want [open/b.txt]", got, err)
	}
}

// TestGlobLinkLoop: ** does not descend through a link, so a link to an ancestor ends.
func TestGlobLinkLoop(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink("..", filepath.Join(dir, "up")); err != nil {
		t.Fatal(err)
	}
	got, err := glob("**", dir, false)
	if err != nil || !slices.Equal(got, []string{"up"}) {
		t.Errorf("glob ** = %v, %v; want [up]", got, err)
	}
}
