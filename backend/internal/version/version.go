// Package version provides version information for the application.
package version

var (
	// Version is the application version. Should be synced with ../../HISTORY.md
	Version = "0.2.17"

	// BuildTime is set during build with -ldflags "-X internal/version.BuildTime=..."
	BuildTime = ""

	// Commit is set during build with -ldflags "-X internal/version.Commit=..."
	Commit = ""
)
