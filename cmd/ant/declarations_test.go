package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDeclarationFileCommands: a .d.ant file is a declaration file (section 10, Declaration
// files) to every command: check checks its signatures, run refuses it, doc lists every
// declaration, and fmt lays it out.
func TestDeclarationFileCommands(t *testing.T) {
	dir := tree(t, map[string]string{
		"db.d.ant":    "// Db.Row is a row.\ntype Db.Row = Map[string, string]\nfunc Db.query(sql: string): Array[Db.Row]\n",
		"bad.d.ant":   "func Db.count(): Count\n",
		"messy.d.ant": "type  Db.Row =  { a: int,b: int }\nfunc Db.query( sql: string ) : Array[Db.Row]\n",
	})
	db, bad, messy := filepath.Join(dir, "db.d.ant"), filepath.Join(dir, "bad.d.ant"), filepath.Join(dir, "messy.d.ant")
	cases := []struct {
		name   string
		args   []string
		status int
		out    string
		err    string
	}{
		{"check clean", []string{"check", db}, 0, "", ""},
		{"check errors", []string{"check", bad}, 1, "", "bad.d.ant:1:18: unknown type Count"},
		{"run", []string{"run", db}, 2, "", "db.d.ant: a declaration file declares names for the prelude and cannot be run"},
		{"run short form", []string{db}, 2, "", "cannot be run"},
		{"doc", []string{"doc", db}, 0, "type Db.Row = Map[string, string]\nfunc Db.query(sql: string): Array[Db.Row]\n", ""},
		{"doc name", []string{"doc", db, "db.row"}, 0, "type Db.Row = Map[string, string]\n\n    Db.Row is a row.\n", ""},
		{"doc owner", []string{"doc", db, "Db"}, 0, "type Db.Row = Map[string, string]\nfunc Db.query(sql: string): Array[Db.Row]\n", ""},
		{"fmt check", []string{"fmt", "--check", messy}, 1, messy + "\n", ""},
		{"fmt", []string{"fmt", messy}, 0, "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errOut strings.Builder
			status := run(c.args, strings.NewReader(""), &out, &errOut)
			if status != c.status {
				t.Errorf("status %d, want %d; stderr: %s", status, c.status, errOut.String())
			}
			if out.String() != c.out {
				t.Errorf("stdout %q, want %q", out.String(), c.out)
			}
			if !strings.Contains(errOut.String(), c.err) {
				t.Errorf("stderr %q, want one containing %q", errOut.String(), c.err)
			}
		})
	}
	got, _ := os.ReadFile(messy)
	if want := "type Db.Row = { a: int, b: int }\nfunc Db.query(sql: string): Array[Db.Row]\n"; string(got) != want {
		t.Errorf("fmt wrote %q, want %q", got, want)
	}
}
