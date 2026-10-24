package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const geometry = `// Point is a position on the plane.
export type Point = { x: float64, y: float64 }

// Point.norm returns the distance from p to the origin.
//
//     Point.norm(Point{ x: 3.0, y: 4.0 })   // 5.0
export func Point.norm(p: Point): float64 {
  Math.sqrt(p.x * p.x + p.y * p.y)
}

func hidden() {}
`

func TestDocCommand(t *testing.T) {
	dir := tree(t, map[string]string{"geometry.ant": geometry, "bad.ant": syntaxError})
	mod, bad := filepath.Join(dir, "geometry.ant"), filepath.Join(dir, "bad.ant")
	cases := []struct {
		name   string
		args   []string
		status int
		out    string
	}{
		{"prelude func", []string{"Array.length"}, 0, "func Array.length[T](xs: Array[T]): int\n"},
		{"any case", []string{"array.LENGTH"}, 0, "func Array.length[T](xs: Array[T]): int\n"},
		{"namespace", []string{"Mutex"}, 0, "type Mutex\n\nfunc Mutex.new(): Mutex\nfunc Mutex.lock(m: Mutex): void\n" +
			"func Mutex.unlock(m: Mutex): void\nfunc Mutex.withLock[T](m: Mutex, f: func(): T): T\n"},
		{"constant", []string{"Math.PI"}, 0, "var Math.PI: float64\n"},
		{"unknown", []string{"Array.nope"}, 1, ""},
		{"module", []string{mod}, 0, "type Point = { x: float64, y: float64 }\nfunc Point.norm(p: Point): float64\n"},
		{"module name", []string{mod, "point.norm"}, 0, "func Point.norm(p: Point): float64\n\n" +
			"    Point.norm returns the distance from p to the origin.\n\n        Point.norm(Point{ x: 3.0, y: 4.0 })   // 5.0\n"},
		{"module type", []string{mod, "Point"}, 0, "type Point = { x: float64, y: float64 }\n\n" +
			"    Point is a position on the plane.\n\nfunc Point.norm(p: Point): float64\n"},
		{"private", []string{mod, "hidden"}, 1, ""},
		{"syntax error", []string{bad}, 2, ""},
		{"no args", nil, 2, ""},
		{"too many", []string{mod, "a", "b"}, 2, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errOut strings.Builder
			status := run(append([]string{"doc"}, c.args...), strings.NewReader(""), &out, &errOut)
			if status != c.status {
				t.Errorf("status %d, want %d; stderr: %s", status, c.status, errOut.String())
			}
			if c.status == 0 && out.String() != c.out {
				t.Errorf("stdout %q, want %q", out.String(), c.out)
			}
		})
	}
}
