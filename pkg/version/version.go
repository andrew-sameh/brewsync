package version

// Version information set via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Info returns a formatted version string
func Info() string {
	return Version
}

// Full returns complete version information
func Full() string {
	return Version + " (" + Commit + ") built " + BuildDate
}
