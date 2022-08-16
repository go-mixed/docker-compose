package mod

import (
	"fmt"
	"go/build"
	"golang.org/x/mod/modfile"
	"strings"
)

// modFileGoVersion returns the (non-empty) Go version at which the requirements
// in modFile are interpreted, or the latest Go version if modFile is nil.
func modFileGoVersion(modFile *modfile.File) (string, error) {
	if modFile == nil {
		return latestGoVersion()
	}
	if modFile.Go == nil || modFile.Go.Version == "" {
		// The main module necessarily has a go.mod file, and that file lacks a
		// 'go' directive. The 'go' command has been adding that directive
		// automatically since Go 1.12, so this module either dates to Go 1.11 or
		// has been erroneously hand-edited.
		//
		// The semantics of the go.mod file are more-or-less the same from Go 1.11
		// through Go 1.16, changing at 1.17 to support module graph pruning.
		// So even though a go.mod file without a 'go' directive is theoretically a
		// Go 1.11 file, scripts may assume that it ends up as a Go 1.16 module.
		return "1.16", nil
	}
	return modFile.Go.Version, nil
}

// LatestGoVersion returns the latest version of the Go language supported by
// this toolchain, like "1.17".
func latestGoVersion() (string, error) {
	tags := build.Default.ReleaseTags
	version := tags[len(tags)-1]
	if !strings.HasPrefix(version, "go") || !modfile.GoVersionRE.MatchString(version[2:]) {
		return "", fmt.Errorf("go: internal error: unrecognized default version %q", version)
	}
	return version[2:], nil
}
