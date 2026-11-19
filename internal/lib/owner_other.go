//go:build !unix

package lib

import (
	"io/fs"

	"github.com/ant-src/ant/value"
)

// userIDs reports whether the host has user and group ids: Windows and WebAssembly do not, so Os.uid,
// Os.gid, and File.Info's uid and gid are null there.
const userIDs = false

// owner is File.Info's uid and gid, which a host without user ids, as Windows, does not have.
func owner(fs.FileInfo) (uid, gid value.Value) { return value.Null, value.Null }
