package version

import "fmt"

const (
	// Major is the major version of the tool
	Major = 0
	// Minor is the minor version of the tool
	Minor = 1
	// Patch is the patch version of the tool
	Patch = 0
)

// GetVersionString generates a version string based on the constants in this package
func GetVersionString() string {
	return fmt.Sprintf("v%d%d%d", Major, Minor, Patch)
}
