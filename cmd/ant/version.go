package main

import "runtime/debug"

// version is the release ant was built from, set by scripts/release.ant with
// -ldflags=-X main.version=vX.Y.Z.
var version string

// antVersion is the version ant version prints: the one the release set, else the module version
// go install recorded (v0.1.0 for go install …@v0.1.0, a pseudo-version or (devel) for a checkout).
func antVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "(devel)"
}
