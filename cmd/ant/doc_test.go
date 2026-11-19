package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const geometry = `// Copyright 2026. MIT license.

//! Geometry computes distances on the plane.
//!
//!     Point.norm(Point{ x: 3.0, y: 4.0 })   // 5.0

// Point is a position on the plane.
export type Point = { x: float64, y: float64 }

// Point.norm returns the distance from p to the origin.
//
//     Point.norm(Point{ x: 3.0, y: 4.0 })   // 5.0
export func Point.norm(p: Point): float64 {
  Math.sqrt(p.x * p.x + p.y * p.y)
}

// Shape has an area.
export interface Shape {
  area(s: Self): float64
}

func hidden() {}
`

func TestDocCommand(t *testing.T) {
	dir := tree(t, map[string]string{"geometry.ant": geometry, "bad.ant": syntaxError, "empty.ant": "func main() {}\n"})
	mod, bad, empty := filepath.Join(dir, "geometry.ant"), filepath.Join(dir, "bad.ant"), filepath.Join(dir, "empty.ant")
	// The prelude's docs change as they are written, so a case on the prelude checks how its
	// output starts, and a case on a module checks all of it.
	cases := []struct {
		name   string
		args   []string
		status int
		out    string
	}{
		{"prelude func", []string{"Array.length"}, 0, "func Array.length[T](xs: Array[T]): int\n\n    Array.length "},
		{"any case", []string{"array.LENGTH"}, 0, "func Array.length[T](xs: Array[T]): int\n\n    Array.length "},
		{"namespace", []string{"Mutex"}, 0, "type Mutex\n\n    Mutex "},
		{"constant", []string{"Math.PI"}, 0, "let Math.PI: float64\n\n    Math.PI "},
		{"namespace doc", []string{"array"}, 0, "Array\n\n    Array holds "},
		{"unknown", []string{"Array.nope"}, 1, ""},
		{"module", []string{mod}, 0, "    Geometry computes distances on the plane.\n\n        Point.norm(Point{ x: 3.0, y: 4.0 })   // 5.0\n\n" +
			"type Point = { x: float64, y: float64 }\nfunc Point.norm(p: Point): float64\ninterface Shape\n"},
		{"module without docs or exports", []string{empty}, 1, ""},
		{"module name", []string{mod, "point.norm"}, 0, "func Point.norm(p: Point): float64\n\n" +
			"    Point.norm returns the distance from p to the origin.\n\n        Point.norm(Point{ x: 3.0, y: 4.0 })   // 5.0\n"},
		{"module type", []string{mod, "Point"}, 0, "type Point = { x: float64, y: float64 }\n\n" +
			"    Point is a position on the plane.\n\nfunc Point.norm(p: Point): float64\n"},
		{"module interface", []string{mod, "Shape"}, 0, "interface Shape\n\n    Shape has an area.\n\narea(s: Self): float64\n"},
		{"private", []string{mod, "hidden"}, 1, ""},
		{"syntax error", []string{bad}, 2, ""},
		{"no args", nil, 2, ""},
		{"too many", []string{mod, "a", "b"}, 2, ""},
	}
	// a name that names nothing gets a suggestion on standard error (doc.md section 1)
	suggestions := []struct{ name, err string }{
		{"Array.join", "ant doc: no declaration Array.join; did you mean Path.join or String.join?\n"},
		{"join", "ant doc: no declaration join; did you mean Path.join or String.join?\n"},
		{"Json.prase", "ant doc: no declaration Json.prase; did you mean Json.parse?\n"},
		{"Arry", "ant doc: no declaration Arry; did you mean Array?\n"},
		{"Json.decode", "ant doc: no declaration Json.decode; ant doc Json lists its members\n"},
		{"percentile", "ant doc: no declaration percentile\n"},
	}
	for _, c := range suggestions {
		var out, errOut strings.Builder
		if status := run([]string{"doc", c.name}, strings.NewReader(""), &out, &errOut); status != 1 || errOut.String() != c.err {
			t.Errorf("ant doc %s: status %d, stderr %q, want %q", c.name, status, errOut.String(), c.err)
		}
	}
	var out, errOut strings.Builder
	if run([]string{"doc", mod, "Point.nrom"}, strings.NewReader(""), &out, &errOut); errOut.String() != "ant doc: no declaration Point.nrom; did you mean Point.norm?\n" {
		t.Errorf("ant doc %s Point.nrom: stderr %q", mod, errOut.String())
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errOut strings.Builder
			status := run(append([]string{"doc"}, c.args...), strings.NewReader(""), &out, &errOut)
			if status != c.status {
				t.Errorf("status %d, want %d; stderr: %s", status, c.status, errOut.String())
			}
			got := out.String()
			if c.status == 0 && len(c.args) == 1 && !strings.HasSuffix(c.args[0], ".ant") {
				got = got[:min(len(got), len(c.out))]
			}
			if c.status == 0 && got != c.out {
				t.Errorf("stdout %q, want %q", out.String(), c.out)
			}
		})
	}
}
