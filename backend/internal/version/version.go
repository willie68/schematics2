// Package version provides version information for the application.
package version

var (
	// Version is the application version. Should be synced with ../../HISTORY.md
	Version = "0.2.24"

	// BuildTime is set during build with -ldflags "-X internal/version.BuildTime=..."
	BuildTime = ""

	// Commit is set during build with -ldflags "-X internal/version.Commit=..."
	Commit = ""

	// ClientBasePath is the external base path baked in at build time (e.g. /schematics2).
	// Set via -ldflags "-X internal/version.ClientBasePath=/schematics2"
	ClientBasePath = ""
)
