package version

import (
	"fmt"
	"runtime"
)

// Version defines the current application semantic version.
const Version = "0.0.1"

// GitCommit defines the application's Git short SHA-1 revision set during the
// build.
var GitCommit = ""

// ClientVersion returns the application's full version string.
func ClientVersion() string {
	if GitCommit != "" {
		return fmt.Sprintf("%s-%s+%s/%s", Version, GitCommit, runtime.GOOS, runtime.Version())
	}

	return fmt.Sprintf("%s+%s/%s", Version, runtime.GOOS, runtime.Version())
}
