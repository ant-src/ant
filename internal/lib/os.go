package lib

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

func init() {
	register("Os", func() value.Value {
		return namespace("Os", append(processEntries(),
			fn("Os.args", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				args := g.Instance().Args()
				out := make([]value.Value, len(args))
				for i, s := range args {
					out[i] = value.Text(s)
				}
				return value.NewArray(out)
			}),
			fn("Os.program", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.Text(g.Instance().Program())
			}),
			fn("Os.platform", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.Text(runtime.GOOS)
			}),
			fn("Os.arch", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return value.Text(runtime.GOARCH)
			}),
			fn("Os.env", "name", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v, ok := g.Instance().LookupEnv(stringArg(g, who, a[0]))
				if !ok {
					return value.Null
				}
				return value.Text(v)
			}),
			fn("Os.environ", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				m := value.NewMap()
				for _, kv := range g.Instance().Environ() {
					k, v, _ := strings.Cut(kv, "=")
					m.Set(value.Text(k), value.Text(v))
				}
				return m
			}),
			fn("Os.cwd", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				dir, err := os.Getwd()
				if err != nil {
					return errResult(err)
				}
				return value.Ok(value.Text(dir))
			}),
			fn("Os.home", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return pathResult(userDir(g.Instance().LookupEnv, homeDir))
			}),
			fn("Os.configDir", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return pathResult(userDir(g.Instance().LookupEnv, configDir))
			}),
			fn("Os.cacheDir", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return pathResult(userDir(g.Instance().LookupEnv, cacheDir))
			}),
			fn("Os.pid", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return int64(os.Getpid())
			}),
			fn("Os.uid", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				if uid := os.Geteuid(); uid >= 0 {
					return int64(uid)
				}
				return value.Null
			}),
			fn("Os.hostname", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return pathResult(os.Hostname())
			}),
			fn("Os.exit", "status", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				g.Instance().Exit(int(intArg(g, who, a[0])))
				return nil
			}),
			// The standard streams are one handle each per run, so that a program reading its
			// input from two places shares one buffer.
			fn("Os.interrupt", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return g.Instance().Interrupt()
			}),
			fn("Os.stdin", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				in := g.Instance()
				return in.Keep(who, func() any {
					p := &pump{src: in.Stdin(), done: in.Context().Done()}
					h := newHandle(nil, p, nil)
					h.std, h.pump, h.childIn = true, p, in.ChildStdin()
					return h
				}).(*Handle)
			}),
			fn("Os.stdout", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				in := g.Instance()
				return in.Keep(who, func() any {
					h := newHandle(nil, nil, in.Stdout())
					h.std, h.childOut = true, in.ChildStdout()
					return h
				}).(*Handle)
			}),
			fn("Os.stderr", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				in := g.Instance()
				return in.Keep(who, func() any {
					h := newHandle(nil, nil, in.ErrWriter())
					h.std, h.childOut = true, in.ChildStderr()
					return h
				}).(*Handle)
			}),
		)...)
	})
}

// pathResult is Ok of a path, or the Err of err.
func pathResult(p string, err error) value.Value {
	if err != nil {
		return errResult(err)
	}
	return value.Ok(value.Text(p))
}

// The user directories of Os.home, Os.configDir and Os.cacheDir.
const (
	homeDir = iota
	configDir
	cacheDir
)

// userDir is Go's os.UserHomeDir, os.UserConfigDir or os.UserCacheDir, by the same rules, but
// reading the run's environment, which Os.env reads, rather than the process's.
func userDir(env func(string) (string, bool), which int) (string, error) {
	get := func(name string) string {
		v, _ := lookupEnvFold(env, name)
		return v
	}
	if runtime.GOOS == "windows" {
		name := [...]string{"USERPROFILE", "AppData", "LocalAppData"}[which]
		if v := get(name); v != "" {
			return v, nil
		}
		return "", errors.New("%" + name + "% is not defined")
	}
	if which != homeDir && runtime.GOOS != "darwin" {
		name := [...]string{"", "XDG_CONFIG_HOME", "XDG_CACHE_HOME"}[which]
		if v := get(name); v != "" {
			if !filepath.IsAbs(v) {
				return "", errors.New("path in $" + name + " is relative")
			}
			return v, nil
		}
	}
	home := get("HOME")
	if home == "" {
		return "", errors.New("$HOME is not defined")
	}
	switch {
	case which == homeDir:
		return home, nil
	case runtime.GOOS == "darwin":
		return filepath.Join(home, "Library", [...]string{"", "Application Support", "Caches"}[which]), nil
	default:
		return filepath.Join(home, [...]string{"", ".config", ".cache"}[which]), nil
	}
}
