package ant_test

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)

// Running a program from a string, with its output captured.
func Example_run() {
	src := `
for name in Os.args() { println("hello, ${name}") }
`
	prog, err := ant.Load("main.ant", ant.WithSource("main.ant", src))
	if err != nil {
		fmt.Println(err)
		return
	}
	var out strings.Builder
	err = prog.Run(context.Background(), ant.WithStdout(&out), ant.WithArgs([]string{"ada", "linus"}))
	fmt.Print(out.String())
	fmt.Println("exit status", ant.ExitStatus(err))
	// Output:
	// hello, ada
	// hello, linus
	// exit status 0
}

// Calling an Ant function from Go.
func Example_call() {
	src := `
type Point = { x: float64, y: float64 }

func scale(p: Point, k: float64): Point { Point{ x: p.x * k, y: p.y * k } }
`
	prog, err := ant.Load("geometry.ant", ant.WithSource("geometry.ant", src))
	if err != nil {
		fmt.Println(err)
		return
	}
	in, err := prog.Start(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()
	pointTag, _ := in.Value("Point")
	scale, _ := in.Func("scale")
	p := value.NewRecord(pointTag.(*value.Tag), []value.Value{1.5, 2.0})
	result, err := scale.Call(p, 2.0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(value.Render(result))
	// Output:
	// Point{ x: 3.0, y: 4.0 }
}

// Giving a program a host namespace, declared in Ant notation and implemented in Go.
func Example_namespace() {
	store := map[string]string{}
	kv := ant.Namespace{
		Name: "Kv",
		Signatures: `
func Kv.get(key: string): string | null
func Kv.set(key: string, v: string): void
`,
		Funcs: map[string]ant.HostFunc{
			"get": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				if v, ok := store[args[0].(string)]; ok {
					return v, nil
				}
				return value.Null, nil
			},
			"set": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				store[args[0].(string)] = args[1].(string)
				return value.Void, nil
			},
		},
	}
	src := `
Kv.set("greeting", "hi")
println(Kv.get("greeting") ?? "missing")
println(Kv.get("other") ?? "missing")
`
	prog, err := ant.Load("main.ant", ant.WithSource("main.ant", src), ant.WithNamespace(kv), ant.WithoutNamespaces("File", "Http"))
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := prog.Run(context.Background(), ant.WithStdout(os.Stdout)); err != nil {
		fmt.Println(err)
	}
	// Output:
	// hi
	// missing
}
