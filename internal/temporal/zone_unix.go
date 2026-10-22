//go:build !windows

package temporal

import (
	"os"
	"strings"
)

// systemZoneName is the zone /etc/localtime links to, or the one /etc/timezone names.
func systemZoneName() string {
	if target, err := os.Readlink("/etc/localtime"); err == nil {
		if i := strings.LastIndex(target, "zoneinfo/"); i >= 0 {
			return target[i+len("zoneinfo/"):]
		}
	}
	if data, err := os.ReadFile("/etc/timezone"); err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}
