package main

import (
	"embed"
	"github.com/andreimerlescu/igo/internal"
	"strings"
)

// binaryVersionBytes contains the embedded VERSION file's contents
//
//go:embed VERSION
var binaryVersionBytes embed.FS

// binaryCurrentVersion is defined by BinaryVersion() and contains the contents of
// the VERSION file
var binaryCurrentVersion string

// BinaryVersion returns the embedded VERSION file of the igo repository as a string
// and cache that value into binaryCurrentVersion once os.ReadFile is complete
func BinaryVersion() string {
	if len(binaryCurrentVersion) == 0 {
		versionBytes, err := binaryVersionBytes.ReadFile("VERSION")
		internal.Capture(err)
		binaryCurrentVersion = strings.TrimSpace(string(versionBytes))
	}
	return binaryCurrentVersion
}

const (
	// Register the go environment variables as constants

	GOBIN          string = "GOBIN"
	GOPATH         string = "GOPATH"
	GOROOT         string = "GOROOT"
	GOMODCACHE     string = "GOMODCACHE"
	GOOS           string = "GOOS"
	GOARCH         string = "GOARCH"
	GOSCRIPTS      string = "GOSCRIPTS"
	GOSHIMS        string = "GOSHIMS"
	GOTELEMETRYDIR string = "GOTELEMETRYDIR"
	GOCACHE        string = "GOCACHE"

	// kVersion defines -version in the CLI to print the BinaryVersion()
	kVersion string = "version"

	// kGoDir defines -godir in the CLI to assign igoWorkspace()
	kGoDir string = "godir"

	// kCommand defines -cmd in the CLI to run commands by names and aliases
	kCommand string = "cmd"

	// kGoVersion defines the -gover in the CLI that targets a Version
	kGoVersion string = "gover"

	// kSystem defines -system in the CLI that ignores userHomeDir in igoWorkspace()
	kSystem string = "system"

	// kGoos defines -goos in the CLI that allows you to define these values without
	// requiring you to set ENV variables first
	kGoos string = "goos"

	// kExtras defines -extras in the CLI that allows you to install extra packages
	kExtras string = "extras"

	// kExtraPackages defines -extra-packages in the CLI that allows you to install extra packages
	kExtraPackages string = "extra-packages"

	// kGoArch defines -goarch in the CLI that allows you to define these values
	// without requiring you to set ENV variables first
	kGoArch string = "goarch"

	kDebug   string = "debug"
	kVerbose string = "verbose"
)

// packages are installed after a new version of Go is installed
var packages = map[string]string{
	"genwordpass":          "github.com/ProjectApario/genwordpass",
	"summarize":            "github.com/andreimerlescu/summarize",
	"counter":              "github.com/andreimerlescu/counter",
	"govulncheck":          "golang.org/x/vuln/cmd/govulncheck",
	"go-generate-password": "github.com/m1/go-generate-password/cmd/go-generate-password",
	"cli-gematria":         "github.com/andreimerlescu/cli-gematria",
}
