package lib

import (
	"errors"
	"io/fs"
	"net"
	"net/url"
	"os"
	"syscall"
	"testing"

	"ant-lang.org/value"
)

// TestKindSurvivesRewording: Net and Http word their own messages from a Go error, and the Err
// keeps the kind of that error (stdlib.md), which a port the user may not bind shows.
func TestKindSurvivesRewording(t *testing.T) {
	denied := &net.OpError{Op: "listen", Net: "tcp", Err: os.NewSyscallError("bind", syscall.EACCES)}
	got := value.Render(errResult(netError("listen", "127.0.0.1:80", denied)))
	want := `Err{ error: Error{ message: "listen \"127.0.0.1:80\": permission denied", kind: Error.Permission{} } }`
	if got != want {
		t.Errorf("netError: %s, want %s", got, want)
	}
	refused := &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}
	if got := value.Render(errResult(netError("dial", "127.0.0.1:1", refused))); got != `Err{ error: Error{ message: "dial \"127.0.0.1:1\": connection refused", kind: null } }` {
		t.Errorf("netError: %s, want no kind", got)
	}
	blocked := &url.Error{Op: "Get", URL: "http://10.0.0.1/", Err: &net.OpError{Op: "dial", Err: os.NewSyscallError("connect", syscall.EACCES)}}
	if err := httpReason(blocked); !errors.Is(err, fs.ErrPermission) || err.Error() != "permission denied" {
		t.Errorf("httpReason: %v, want permission denied as fs.ErrPermission", err)
	}
	if err := httpReason(&url.Error{Op: "Get", URL: "http://x/", Err: errors.New("net/http: timeout awaiting response headers")}); err.Error() != "timeout awaiting response headers" {
		t.Errorf("httpReason: %v, want the reason without its prefix", err)
	}
}
