package lib

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ant-src/ant/value"
)

// TestModeRoundTrip: an Ant mode in 0 ..= 0o7777 survives the trip through Go's FileMode, setuid,
// setgid and sticky included (stdlib.md, File: Modes).
func TestModeRoundTrip(t *testing.T) {
	for _, m := range []int64{0, 0o644, 0o755, 0o4755, 0o2750, 0o1777, 0o7777} {
		if got := antMode(goMode(m)); got != m {
			t.Errorf("antMode(goMode(%O)) = %O", m, got)
		}
	}
	if got := antMode(fs.ModeDir | fs.ModeSymlink | 0o700); got != 0o700 {
		t.Errorf("type bits leak into the mode: %O", got)
	}
}

// TestUserDir: Os.home, Os.configDir and Os.cacheDir follow Go's rules, read from the run's
// environment (stdlib.md, Os).
func TestUserDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the Unix rules")
	}
	env := func(vars map[string]string) func(string) (string, bool) {
		return func(name string) (string, bool) { v, ok := vars[name]; return v, ok }
	}
	home := env(map[string]string{"HOME": "/home/ant"})
	if got, err := userDir(home, homeDir); err != nil || got != "/home/ant" {
		t.Errorf("home = %q, %v", got, err)
	}
	wantConfig, wantCache := "/home/ant/.config", "/home/ant/.cache"
	if runtime.GOOS == "darwin" {
		wantConfig, wantCache = "/home/ant/Library/Application Support", "/home/ant/Library/Caches"
	}
	if got, err := userDir(home, configDir); err != nil || got != wantConfig {
		t.Errorf("configDir = %q, %v; want %q", got, err, wantConfig)
	}
	if got, err := userDir(home, cacheDir); err != nil || got != wantCache {
		t.Errorf("cacheDir = %q, %v; want %q", got, err, wantCache)
	}
	if _, err := userDir(env(nil), homeDir); err == nil {
		t.Error("home without $HOME is Ok")
	}
	if runtime.GOOS != "darwin" {
		xdg := env(map[string]string{"HOME": "/home/ant", "XDG_CONFIG_HOME": "/xdg/config"})
		if got, _ := userDir(xdg, configDir); got != "/xdg/config" {
			t.Errorf("configDir with XDG_CONFIG_HOME = %q", got)
		}
		relative := env(map[string]string{"HOME": "/home/ant", "XDG_CACHE_HOME": "cache"})
		if _, err := userDir(relative, cacheDir); err == nil {
			t.Error("a relative XDG_CACHE_HOME is Ok")
		}
	}
}

// TestCopyAllReadOnlyDirectory: a directory whose mode forbids writing is filled before its mode
// is set (stdlib.md, File: Copies).
func TestCopyAllReadOnlyDirectory(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src")
	if err := os.MkdirAll(filepath.Join(src, "ro"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "ro", "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(src, "ro"), 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(filepath.Join(src, "ro"), 0o755) })
	dst := filepath.Join(t.TempDir(), "dst")
	t.Cleanup(func() { os.Chmod(filepath.Join(dst, "ro"), 0o755) })
	if r := value.Render(copyAll(src, dst)); r != "Ok{ value: void }" {
		t.Fatalf("copyAll = %s", r)
	}
	fi, err := os.Stat(filepath.Join(dst, "ro"))
	if err != nil || fi.Mode().Perm() != 0o555 {
		t.Fatalf("mode = %v, %v; want 0o555", fi.Mode().Perm(), err)
	}
	if data, err := os.ReadFile(filepath.Join(dst, "ro", "f")); err != nil || string(data) != "x" {
		t.Fatalf("copied file = %q, %v", data, err)
	}
}

// TestCopyAllUnderLinkedWorkingDirectory: a to inside from is caught when the working directory is
// reached through a link, so from and to, both relative, resolve it alike, and nothing is created
// (stdlib.md, File: Copies).
func TestCopyAllUnderLinkedWorkingDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need privileges")
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "real", "site"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "real", "site", "index.html"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real", filepath.Join(root, "via")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(root, "via"))
	want := `Err{ error: Error{ message: "copy site to site/backup: site/backup lies inside site" } }`
	if r := value.Render(copyAll("site", "site/backup")); r != want {
		t.Fatalf("copyAll = %s; want %s", r, want)
	}
	if _, err := os.Lstat(filepath.Join(root, "real", "site", "backup")); err == nil {
		t.Fatal("copyAll created site/backup")
	}
}

// TestDefaultTempDir: the temporary directory comes from the run's environment, which Os.env
// reads, not the process's (stdlib.md, File: Temporary files).
func TestDefaultTempDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the Unix rule")
	}
	env := func(vars map[string]string) func(string) (string, bool) {
		return func(name string) (string, bool) { v, ok := vars[name]; return v, ok }
	}
	if got := defaultTempDir(env(map[string]string{"TMPDIR": "/sandbox/tmp"})); got != "/sandbox/tmp" {
		t.Errorf("with TMPDIR = %q", got)
	}
	if got := defaultTempDir(env(nil)); got != "/tmp" {
		t.Errorf("without TMPDIR = %q", got)
	}
}
