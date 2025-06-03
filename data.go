package main

import (
	"embed"
	"strings"

	"github.com/andreimerlescu/igo/internal"
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

	VerboseEnabled    string = "VERBOSE MODE ENABLED"
	DebugEnabled      string = "DEBUG MODE ENABLED"
	CreatedSymlinkFmt string = "Created symlink %s -> %s"
	LinkingFmt        string = "Linking %v to %v"
	NoGoMessage       string = "No installed version of Go found"
	IndentList        string = "│   ├── %v=%s"
	IndentItem        string = "│   ├── %v"
	IndentValue       string = "    ├── %v -> %v %s  "
	VersionFmt        string = "%d.%d.%d"

	cmdInstall   string = "i" // mutagenesis = string (version)
	cmdUninstall string = "u" // mutagenesis = string (version)
	cmdActivate  string = "a" // mutagenesis = string (version)
	cmdFix       string = "f" // mutagenesis = string (version) ; empty = current version activated
	cmdList      string = "l" // mutagenesis = bool (true = display list)
	cmdHelp      string = "h" // mutagenesis = bool (true = display help)
	cmdVersion   string = "v" // mutagenesis = bool (true = display version)
	cmdSwitch    string = "s" // mutagenesis = string (version)
	cmdEnv       string = "e" // mutagenesis = bool (true = display env)

	// kGoDir defines -godir in the CLI to assign igoWorkspace()
	kGoDir string = "godir"

	// kSystem defines -system in the CLI that ignores UserHomeDir in igoWorkspace()
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
