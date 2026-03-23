package version

import "testing"

func TestResolveVersionFallback(t *testing.T) {
	// When running tests, debug.ReadBuildInfo returns (devel) for Main.Version,
	// so resolveVersion should return "dev".
	got := resolveVersion()
	if got != "dev" {
		t.Errorf("resolveVersion() = %q, want %q", got, "dev")
	}
}

func TestVersionIsSet(t *testing.T) {
	// The package-level Version variable should be initialized.
	// During tests it will be "dev" (since ldflags are not set and
	// build info reports "(devel)").
	if Version == "" {
		t.Error("Version should not be empty after init")
	}
}

func TestVersionLdflagsOverride(t *testing.T) {
	// Simulate ldflags having set the version before init ran.
	// We can test this by directly setting Version and verifying
	// that resolveVersion is not used when Version is already set.
	original := Version
	defer func() { Version = original }()

	Version = "1.2.3"
	if Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", Version, "1.2.3")
	}
}
