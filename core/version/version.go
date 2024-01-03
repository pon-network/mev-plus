package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const ourPath = "github.com/pon-network/mev-plus" // Path to our module

const (
	VersionMajor = 0        // Major version component of the current release
	VersionMinor = 0        // Minor version component of the current release
	VersionPatch = 9        // Patch version component of the current release
	VersionMeta  = "stable" // Version metadata to append to the version string
)

// Version holds the textual version string.
var Version = func() string {
	return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
}()

// VersionWithMeta holds the textual version string including the metadata.
var VersionWithMeta = func() string {
	v := Version
	if VersionMeta != "" {
		v += "-" + VersionMeta
	}
	return v
}()

// Info returns version information for the currently executing binary.
func Info() (version string) {
	version = VersionWithMeta
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}
	version = version + "-" + versionInfo(buildInfo)
	return version
}

// versionInfo returns version information for the currently executing
// implementation.
//
// Depending on how the code is instantiated, it returns different amounts of
// information. If it is unable to determine which module is related to our
// package it falls back to the hardcoded values in the params package.
func versionInfo(info *debug.BuildInfo) string {
	// If the main package is from our repo, prefix version with "mevPlus".
	if strings.HasPrefix(info.Path, ourPath) {
		return fmt.Sprintf("mevPlus %s", info.Main.Version)
	}
	// Not our main package, so explicitly print out the module path and
	// version.
	var version string
	if info.Main.Path != "" && info.Main.Version != "" {
		// These can be empty when invoked with "go run".
		version = fmt.Sprintf("%s@%s ", info.Main.Path, info.Main.Version)
	}
	mod := findModule(info, ourPath)
	if mod == nil {
		// If our module path wasn't imported, it's unclear which
		// version of our code they are running. Fallback to hardcoded
		// version.
		return version + fmt.Sprintf("mevPlus")
	}
	// Our package is a dependency for the main module. Return path and
	// version data for both.
	version += fmt.Sprintf("%s@%s", mod.Path, mod.Version)
	if mod.Replace != nil {
		// If our package was replaced by something else, also note that.
		version += fmt.Sprintf(" (replaced by %s@%s)", mod.Replace.Path, mod.Replace.Version)
	}
	return version
}

// findModule returns the module at path.
func findModule(info *debug.BuildInfo, path string) *debug.Module {
	if info.Path == ourPath {
		return &info.Main
	}
	for _, mod := range info.Deps {
		if mod.Path == path {
			return mod
		}
	}
	return nil
}
