// Package version provides the current version of gitusr.
// The Version variable is set at build time via ldflags (e.g., -X
// github.com/northwang-lucky/gitusr/internal/version.Version=v1.0.0)
// and defaults to "dev" for local development builds.
package version

// Version holds the current gitusr version string.
// Override with ldflags during release builds.
var Version = "dev"
