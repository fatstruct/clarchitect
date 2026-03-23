package version

import "runtime/debug"

// Version is the application version. It can be set at build time via ldflags:
//
//	go build -ldflags "-X github.com/fatstruct/clarchitect/internal/version.Version=0.1.0"
//
// If not set via ldflags, it falls back to the module version from
// debug.ReadBuildInfo, or "dev" if neither is available.
var Version = ""

func init() {
	if Version != "" {
		return
	}
	Version = resolveVersion()
}

// resolveVersion attempts to read the version from build info.
// Returns "dev" if no version information is available.
func resolveVersion() string {
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
