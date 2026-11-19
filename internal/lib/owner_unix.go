//go:build unix

package lib

import (
	"io/fs"
	"syscall"

	"github.com/ant-src/ant/value"
)

// userIDs reports whether the host has user and group ids, which Os.uid and Os.gid return.
const userIDs = true

// owner is File.Info's uid and gid: the ids of the user and the group that own the file.
func owner(fi fs.FileInfo) (uid, gid value.Value) {
	if st, ok := fi.Sys().(*syscall.Stat_t); ok {
		return int64(st.Uid), int64(st.Gid)
	}
	return value.Null, value.Null
}
