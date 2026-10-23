package tests

import (
	"bytes"
	"context"
	"testing"
	"testing/fstest"

	"github.com/ant-src/ant"
)

// TestFarDeadline checks that Context.deadline is right for a deadline past 2262, where
// nanoseconds since the epoch overflow; the run has no deadline of its own, since the host's
// would be the earlier one.
func TestFarDeadline(t *testing.T) {
	prog, err := ant.Load("main.ant", ant.WithSource("main.ant", `
func main(): void {
  var far = Context.new(Time.duration(days: 100000))
  var deadline = Context.deadline(far)
  print(deadline != null && Time.compare(deadline, Time.fromEpochMillis(4102444800000)) > 0)
  Context.cancel(far)
}
`), ant.WithFS(fstest.MapFS{}))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := prog.Run(context.Background(), ant.WithStdout(&out)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "true\n" {
		t.Fatalf("printed %q", out.String())
	}
}
