//go:build windows

package temporal

import (
	"syscall"
	"unsafe"
)

// systemZoneName is the IANA zone CLDR gives the Windows zone named in the registry.
func systemZoneName() string {
	path, err := syscall.UTF16PtrFromString(`SYSTEM\CurrentControlSet\Control\TimeZoneInformation`)
	if err != nil {
		return ""
	}
	var key syscall.Handle
	if err := syscall.RegOpenKeyEx(syscall.HKEY_LOCAL_MACHINE, path, 0, syscall.KEY_READ, &key); err != nil {
		return ""
	}
	defer syscall.RegCloseKey(key)
	name, err := syscall.UTF16PtrFromString("TimeZoneKeyName")
	if err != nil {
		return ""
	}
	var buf [256]uint16
	n := uint32(len(buf) * 2)
	var typ uint32
	if err := syscall.RegQueryValueEx(key, name, nil, &typ, (*byte)(unsafe.Pointer(&buf[0])), &n); err != nil {
		return ""
	}
	return windowsZones[syscall.UTF16ToString(buf[:])]
}
