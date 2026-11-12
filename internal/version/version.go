// Package version exposes build information, set with -ldflags at release time.
package version

import "runtime/debug"

// Set by `-ldflags "-X .../internal/version.Version=..."`.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns a one-line description, falling back to the module version
// embedded by `go install` when no ldflags were given.
func String() string {
	v := Version
	if v == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			v = info.Main.Version
		}
	}
	return v + " (commit " + Commit + ", built " + Date + ")"
}
