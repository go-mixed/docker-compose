package mod

import (
	"errors"
	"fmt"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type VendorList struct {
	vendorList      []module.Version          // modules that contribute packages to the build, in order of appearance
	vendorReplaced  []module.Version          // all replaced modules; may or may not also contribute packages
	vendorVersion   map[string]string         // module path → selected version (if known)
	vendorPkgModule map[string]module.Version // package → containing module
	vendorMeta      map[module.Version]vendorMetadata
	rawGoVersion    *sync.Map
}

type vendorMetadata struct {
	Explicit    bool
	Replacement module.Version
	GoVersion   string
}

// readVendorList reads the list of vendored modules from vendor/modules.txt.
func readVendorList(vendorPath string) (*VendorList, error) {

	var (
		rawGoVersion    sync.Map                          // map[module.Version]string
		vendorList      []module.Version                  // modules that contribute packages to the build, in order of appearance
		vendorReplaced  []module.Version                  // all replaced modules; may or may not also contribute packages
		vendorVersion   = make(map[string]string)         // module path → selected version (if known)
		vendorPkgModule = make(map[string]module.Version) // package → containing module
		vendorMeta      = make(map[module.Version]vendorMetadata)
	)

	data, err := os.ReadFile(filepath.Join(vendorPath, "modules.txt"))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("go: %s", err)
		}
		return nil, err
	}

	var mod module.Version
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "# ") {
			f := strings.Fields(line)

			if len(f) < 3 {
				continue
			}
			if semver.IsValid(f[2]) {
				// A module, but we don't yet know whether it is in the build list or
				// only included to indicate a replacement.
				mod = module.Version{Path: f[1], Version: f[2]}
				f = f[3:]
			} else if f[2] == "=>" {
				// A wildcard replacement found in the main module's go.mod file.
				mod = module.Version{Path: f[1]}
				f = f[2:]
			} else {
				// Not a version or a wildcard replacement.
				// We don't know how to interpret this module line, so ignore it.
				mod = module.Version{}
				continue
			}

			if len(f) >= 2 && f[0] == "=>" {
				meta := vendorMeta[mod]
				if len(f) == 2 {
					// File replacement.
					meta.Replacement = module.Version{Path: f[1]}
					vendorReplaced = append(vendorReplaced, mod)
				} else if len(f) == 3 && semver.IsValid(f[2]) {
					// Path and version replacement.
					meta.Replacement = module.Version{Path: f[1], Version: f[2]}
					vendorReplaced = append(vendorReplaced, mod)
				} else {
					// We don't understand this replacement. Ignore it.
				}
				vendorMeta[mod] = meta
			}
			continue
		}

		// Not a module line. Must be a package within a module or a metadata
		// directive, either of which requires a preceding module line.
		if mod.Path == "" {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			// Metadata. Take the union of annotations across multiple lines, if present.
			meta := vendorMeta[mod]
			for _, entry := range strings.Split(strings.TrimPrefix(line, "## "), ";") {
				entry = strings.TrimSpace(entry)
				if entry == "explicit" {
					meta.Explicit = true
				}
				if strings.HasPrefix(entry, "go ") {
					meta.GoVersion = strings.TrimPrefix(entry, "go ")
					rawGoVersion.Store(mod, meta.GoVersion)
				}
				// All other tokens are reserved for future use.
			}
			vendorMeta[mod] = meta
			continue
		}

		if f := strings.Fields(line); len(f) == 1 && module.CheckImportPath(f[0]) == nil {
			// A package within the current module.
			vendorPkgModule[f[0]] = mod

			// Since this module provides a package for the build, we know that it
			// is in the build list and is the selected version of its path.
			// If this information is new, record it.
			if v, ok := vendorVersion[mod.Path]; !ok || semver.Compare(v, mod.Version) < 0 {
				vendorList = append(vendorList, mod)
				vendorVersion[mod.Path] = mod.Version
			}
		}
	}

	return &VendorList{
		vendorList:      vendorList,
		vendorReplaced:  vendorReplaced,
		vendorVersion:   vendorVersion,
		vendorPkgModule: vendorPkgModule,
		vendorMeta:      vendorMeta,
		rawGoVersion:    &rawGoVersion,
	}, nil

}
