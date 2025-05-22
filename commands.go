package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/andreimerlescu/igo/internal"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// fix fixes the go version
func fix(app *Application, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, version string) {
	defer wg.Done()
	verbose, debug := *app.figs.Bool(kVerbose), *app.figs.Bool(kDebug)
	onlyVerbose := verbose && !debug

	if verbose {
		color.Green("VERBOSE MODE ENABLED")
	}
	if debug {
		color.Red("DEBUG MODE ENABLED")
	}
	workspace := app.Workspace()
	symlinks, err := internal.FindSymlinks(workspace)
	if err != nil {
		if debug || onlyVerbose {
			color.Red(err.Error())
		}
		errCh <- err
		return
	}
	if len(symlinks) == 0 {
		color.Red("No go versions installed.")
		return
	}
	activeVersion, err := app.activatedVersion()
	if err != nil {
		if debug || onlyVerbose {
			color.Red(err.Error())
		}
		errCh <- fmt.Errorf("no active Go version found")
		return
	}
	if debug || onlyVerbose {
		color.Green("Active Go version: %v", activeVersion)
	}
	files, err := os.ReadDir(workspace)
	if err != nil {
		if debug || onlyVerbose {
			color.Red(err.Error())
		}
		errCh <- err
		return
	}
	results := map[string]string{
		GOBIN:      "",
		GOPATH:     "",
		GOMODCACHE: "",
		GOROOT:     "",
		"version":  "",
	}
	for _, dirEntry := range files {
		if dirEntry.Name() == "bin" {
			results[GOBIN] = filepath.Join(workspace, dirEntry.Name())
		}
		if dirEntry.Name() == "path" {
			results[GOPATH] = filepath.Join(workspace, dirEntry.Name())
		}
		if dirEntry.Name() == "root" {
			results[GOROOT] = filepath.Join(workspace, dirEntry.Name())
		}
		if dirEntry.Name() == "version" {
			results["version"] = filepath.Join(workspace, dirEntry.Name())
		}
		if dirEntry.Name() == "versions" {
			results[GOMODCACHE] = filepath.Join(workspace, "versions", version, "pkg", "mod")
		}
	}
	for _, dirEntry := range files {
		n := fmt.Sprintf("GO%s", strings.ToUpper(dirEntry.Name()))
		if x, exists := results[n]; exists && len(x) > 0 {
			if dirEntry.Type()&os.ModeSymlink == 0 {
				color.Red("ERROR: file %s is NOT a symlink", dirEntry.Name())
			}
		}
	}
	patched := false
	for name, path := range results {
		if len(path) > 0 {
			color.Green("%s: %s", name, path)
		} else {
			color.Red("!!! MISSING %s...", name)
			switch name {
			case GOPATH:
				src := filepath.Join(workspace, "versions", version, "pkg", "mod")
				tar := filepath.Join(workspace, "path")
				err := os.Symlink(src, tar)
				if err != nil {
					if debug || onlyVerbose {
						color.Red(err.Error())
					}
					errCh <- err
					return
				}
				color.Green("Created symlink %s -> %s", src, tar)
				patched = true
			case GOROOT:
				src := filepath.Join(workspace, "versions", version, "go")
				tar := filepath.Join(workspace, "root")
				err := os.Symlink(src, tar)
				if err != nil {
					if debug || onlyVerbose {
						color.Red(err.Error())
					}
					errCh <- err
					return
				}
				color.Green("Created symlink %s -> %s", src, tar)
				patched = true
			case GOBIN:
				src := filepath.Join(workspace, "versions", version, "go", "bin")
				tar := filepath.Join(workspace, "bin")
				err := os.Symlink(src, tar)
				if err != nil {
					if debug || onlyVerbose {
						color.Red(err.Error())
					}
					errCh <- err
					return
				}
				color.Green("Created symlink %s -> %s", src, tar)
				patched = true
			}
		}
	}

	if patched {
		color.Green("Fixed go %s!", version)
	} else {
		color.Green("Nothing to fix!")
	}
}

// env prints the environment variables for the current go version
func env(app *Application, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()
	verbose, debug := *app.figs.Bool(kVerbose), *app.figs.Bool(kDebug)
	onlyVerbose := verbose && !debug

	if verbose {
		color.Green("VERBOSE MODE ENABLED")
	}

	if debug {
		color.Red("DEBUG MODE ENABLED")
	}

	workspace := app.Workspace()
	_, dirErr := os.Stat(workspace)
	if os.IsNotExist(dirErr) {
		color.Red("No go versions installed.")
		return
	}
	currentVersion, err := app.activatedVersion()
	if err != nil {
		if debug || onlyVerbose {
			color.Red(err.Error())
		}
		if !os.IsNotExist(err) {
			errCh <- err
			return
		}
	}
	have := map[string]bool{
		"GOBIN":      false,
		"GOPATH":     false,
		"GOMODCACHE": false,
		"GOROOT":     false,
	}
	color.Green("Current version: %v", currentVersion)
	color.Green("│   ENV:")
	env := os.Environ()
	slices.Sort(env)
	for _, val := range env {
		parts := strings.SplitN(val, "=", 2)
		if len(parts) != 2 || !strings.HasPrefix(parts[0], "GO") {
			continue
		}
		color.Green("│   ├── %v=%v", parts[0], parts[1])
		if _, exists := have[parts[0]]; exists {
			have[parts[0]] = true
		}
	}
	var (
		binDir  = filepath.Join(workspace, "bin")
		pathDir = filepath.Join(workspace, "path")
		rootDir = filepath.Join(workspace, "root")
		shimDir = filepath.Join(workspace, "shims")
		modDir  = filepath.Join(workspace, "versions", currentVersion, "pkg", "mod")
	)
	path := os.Getenv("PATH")
	path = strings.TrimSpace(path)
	requiredPaths := map[string]bool{shimDir: false, binDir: false}
	paths := strings.Split(path, string(filepath.ListSeparator))
	pathVersions := make(map[string]bool)
	for _, p := range paths {
		if _, exists := requiredPaths[p]; exists {
			requiredPaths[p] = true
		}
		if strings.HasPrefix(p, filepath.Join(workspace, "versions")) {
			pathVersions[p] = true
		}
	}
	// writePaths := slices.Clone(paths)
	writePaths := make(map[string]bool)
	for _, p := range paths {
		if _, exists := requiredPaths[p]; exists {
			writePaths[p] = true
		}
	}
	for p, _ := range pathVersions {
		v := strings.TrimPrefix(p, filepath.Join(workspace, "versions"))
		v = strings.TrimSuffix(v, filepath.Join("go", "bin"))
		v = strings.ReplaceAll(v, string(filepath.Separator), "")
		if v != currentVersion {
			for wp, _ := range writePaths {
				if strings.HasPrefix(wp, p) {
					delete(writePaths, wp)
				}
			}
		}
	}
	for k, v := range have {
		if !v {
			switch k {
			case "GOBIN":
				internal.Capture(os.Setenv(k, binDir))
				color.Green("│   ├── %v=%s", k, binDir)
			case "GOPATH":
				internal.Capture(os.Setenv(k, pathDir))
				color.Green("│   ├── %v=%s", k, pathDir)
			case "GOMODCACHE":
				internal.Capture(os.Setenv(k, modDir))
				color.Green("│   ├── %v=%s", k, modDir)
			case "GOROOT":
				internal.Capture(os.Setenv(k, rootDir))
				color.Green("│   ├── %v=%s", k, rootDir)
			}
		}
	}
	color.Green("│   PATH: ")
	for k, v := range requiredPaths {
		if !v {
			writePaths[k] = true
		}
	}
	newPaths := make([]string, 0, len(writePaths))
	for p, _ := range writePaths {
		color.Green("│   ├── %v", p)
		newPaths = append(newPaths, p)
	}
	slices.Sort(newPaths)
	internal.Capture(os.Setenv("PATH", strings.Join(newPaths, string(filepath.ListSeparator))))

	links, err := internal.FindSymlinks(workspace)
	slices.Sort(links)
	if err != nil {
		if debug || onlyVerbose {
			color.Red(err.Error())
		}
		errCh <- err
		return
	}
	color.Green("└── LINKS:")
	for _, link := range links {
		to, err := internal.ReadSymlink(link)
		if err != nil {
			if debug || onlyVerbose {
				color.Red(err.Error())
			}
			continue
		}
		color.Green("    ├── %v -> %v %s  ", link, to, internal.VerifyLink(link, to))
	}

	return
}

// uninstall removes a version of go.
func uninstall(app *Application, ctx context.Context, wg *sync.WaitGroup, errCh chan error, version string) {
	defer wg.Done()
	verbose, debug := *app.figs.Bool(kVerbose), *app.figs.Bool(kDebug)
	onlyVerbose := verbose && !debug

	if verbose {
		color.Green("VERBOSE MODE ENABLED")
	}

	if debug {
		color.Red("DEBUG MODE ENABLED")
	}

	workspace := app.Workspace()
	_, dirErr := os.Stat(workspace)
	if os.IsNotExist(dirErr) {
		color.Red("No go versions installed.")
		return
	}
	currentVersion, err := app.activatedVersion()
	if err != nil {
		if debug || onlyVerbose {
			color.Red(err.Error())
		}
		errCh <- err
		return
	}
	binDir := filepath.Join(workspace, "bin")
	pathDir := filepath.Join(workspace, "path")
	rootDir := filepath.Join(workspace, "root")
	versionDir := filepath.Join(workspace, "versions", version)
	versionFile := filepath.Join(workspace, "version")

	internal.Capture(internal.RemoveStickyBit(versionDir))
	internal.Capture(internal.RemoveSetuidSetgidBits(versionDir))

	if strings.Contains(currentVersion, version) {
		for _, path := range []string{binDir, pathDir, rootDir} {
			internal.Capture(os.RemoveAll(path))
		}
		internal.Capture(os.Remove(versionFile))
	}

	internal.Capture(internal.MakeDirsWritable(versionDir))
	internal.Capture(os.RemoveAll(versionDir))

	color.Green("Uninstalled version: %s", version)
}

// use sets the version of go to use.
func use(app *Application, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, version string) {
	defer wg.Done()
	verbose, debug := *app.figs.Bool(kVerbose), *app.figs.Bool(kDebug)
	onlyVerbose := verbose && !debug

	if verbose {
		color.Green("VERBOSE MODE ENABLED")
	}

	if debug {
		color.Red("DEBUG MODE ENABLED")
	}

	workspace := app.Workspace()
	_, dirErr := os.Stat(workspace)
	if os.IsNotExist(dirErr) {
		color.Red("No go versions installed.")
		return
	}
	currentVersion, err := app.activatedVersion()
	if err != nil {
		if debug || onlyVerbose {
			color.Red(err.Error())
		}
		errCh <- err
		return
	}
	if strings.Contains(currentVersion, version) {
		color.Green("Already using version %v", currentVersion)
		return
	}

	var (
		binDir       = filepath.Join(workspace, "bin")
		pathDir      = filepath.Join(workspace, "path")
		rootDir      = filepath.Join(workspace, "root")
		shimDir      = filepath.Join(workspace, "shims")
		telemetryDir = filepath.Join(workspace, "telemetry")
		cacheDir     = filepath.Join(workspace, "cache")
		scriptsDir   = filepath.Join(workspace, "scripts")
		versionDir   = filepath.Join(workspace, "versions", version)
		modCacheDir  = filepath.Join(workspace, "versions", version, "go", "pkg", "mod")
		versionFile  = filepath.Join(workspace, "version")
	)

	// define the environment that igo requires
	envs := map[string]string{
		GOOS:           *app.figs.String(kGoos),
		GOARCH:         *app.figs.String(kGoArch),
		GOSCRIPTS:      scriptsDir,
		GOSHIMS:        shimDir,
		GOBIN:          binDir,
		GOROOT:         rootDir,
		GOPATH:         pathDir,
		GOMODCACHE:     modCacheDir,
		GOCACHE:        cacheDir,
		GOTELEMETRYDIR: telemetryDir,
	}

	_, err = os.Stat(versionDir)
	if os.IsNotExist(err) {
		if debug || onlyVerbose {
			color.Red("No go version installed.")
		}
		errCh <- err
		return
	}

	var paths = map[string]string{
		filepath.Join(versionDir):              pathDir,
		filepath.Join(versionDir, "go"):        rootDir,
		filepath.Join(versionDir, "go", "bin"): binDir,
	}
	for source, target := range paths {
		if debug {
			color.Green("Linking %v -> %v", target, source)
		}
		err = internal.RemoveSymlinkOrBackupPath(target)
		if err != nil {
			if debug || onlyVerbose {
				color.Red("Failed to link %v -> %v due to err: %s", target, source, err)
			}
			errCh <- err
			return
		}

		if err := os.Symlink(source, target); err != nil {
			if debug || verbose {
				color.Red("Failed to link %v -> %v due to err: %s", source, target, err)
			}
			errCh <- err
			return
		}
	}

	// replace VERSION file of go
	err = os.Remove(versionFile)
	if err != nil {
		if debug || onlyVerbose {
			color.Red("Failed to remove version file due to err: %s", err)
		}
		errCh <- err
		return
	}
	err = os.WriteFile(versionFile, []byte(version), 0644)
	if err != nil {
		if debug || onlyVerbose {
			color.Red("Failed to write version file due to err: %s", err)
		}
		errCh <- err
		return
	}

	if debug {
		versionFound := app.runVersionCheck(envs, version)
		if !strings.Contains(versionFound, version) {
			color.Red("Mismatched go version %v and found %v", version, versionFound)
			errCh <- errors.New("mismatching go versions for switch indicates failure")
			return
		}
	}

	if debug || onlyVerbose {
		color.Green("Set go version %v", version)
	}
}

// list lists all installed go versions
func list(app *Application, ctx context.Context, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()
	verbose, debug := *app.figs.Bool(kVerbose), *app.figs.Bool(kDebug)
	onlyVerbose, onlyDebug := verbose && !debug, !verbose && debug

	if verbose {
		color.Green("VERBOSE MODE ENABLED")
	}

	if debug {
		color.Red("DEBUG MODE ENABLED")
	}

	workspace := app.Workspace()
	_, dirErr := os.Stat(workspace)
	if os.IsNotExist(dirErr) {
		color.Red("No go versions installed.")
		return
	}
	versions, err := app.findGoVersions()
	if err != nil {
		if onlyVerbose {
			color.Red(err.Error())
		}
		if debug {
			color.Red(err.Error())
		}
		errCh <- err
		return
	}
	slices.Sort(versions)
	slices.Reverse(versions)
	currentVersion, _ := app.activatedVersion()
	var data [][]string
	for _, version := range versions {
		info, infoErr := os.Stat(filepath.Join(workspace, "versions", version))
		if os.IsNotExist(infoErr) {
			if onlyDebug {
				errCh <- infoErr
				return
			}
			continue
		}
		a := ""
		if strings.EqualFold(version, currentVersion) {
			a = " * ACTIVE "
		}
		data = append(data, []string{
			version,
			info.ModTime().Format("2006-01-02 15:04"),
			a,
		})
	}
	color.Magenta(internal.About())

	symbols := tw.NewSymbolCustom("Nature").
		WithRow("~").
		WithColumn("|").
		WithTopLeft("🌱").
		WithTopMid("🌿").
		WithTopRight("🌱").
		WithMidLeft("🍃").
		WithCenter("❀").
		WithMidRight("🍃").
		WithBottomLeft("🌻").
		WithBottomMid("🌾").
		WithBottomRight("🌻")

	colorCfg := renderer.ColorizedConfig{
		Header: renderer.Tint{
			FG: renderer.Colors{color.FgGreen, color.Bold}, // Green bold headers
			Columns: []renderer.Tint{
				{FG: renderer.Colors{color.FgHiRed, color.Bold}},
				{FG: renderer.Colors{color.FgHiWhite, color.Bold}},
				{FG: renderer.Colors{color.FgHiBlue, color.Bold}},
			},
			BG: renderer.Colors{color.BgHiWhite},
		},
		Column: renderer.Tint{
			FG: renderer.Colors{color.FgWhite},
			Columns: []renderer.Tint{
				{FG: renderer.Colors{color.FgHiRed}},
				{FG: renderer.Colors{color.FgHiWhite}},
				{FG: renderer.Colors{color.FgHiBlue}},
			},
		},
		Footer: renderer.Tint{
			FG: renderer.Colors{color.FgHiMagenta}, // Yellow bold footer
			Columns: []renderer.Tint{
				{},                                      // Inherit default
				{FG: renderer.Colors{color.FgHiYellow}}, // High-intensity yellow for column 1
				{},                                      // Inherit default
			},
		},
		Border:    renderer.Tint{FG: renderer.Colors{color.FgWhite}},
		Separator: renderer.Tint{FG: renderer.Colors{color.FgWhite}},
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: symbols})),
		tablewriter.WithRenderer(renderer.NewColorized(colorCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoWrap:  tw.WrapNormal,
					Alignment: tw.AlignLeft,
				},
			},
			Footer: tw.CellConfig{
				Formatting: tw.CellFormatting{Alignment: tw.AlignCenter},
			},
		}),
	)

	table.Header([]string{"Version", "Creation", "Status"})
	table.Footer([]string{"I ❤ YOU!", "Made In America", "Be Inspired"})
	err = table.Bulk(data)
	if err != nil {
		if debug || onlyVerbose {
			color.Red(err.Error())
		}
		errCh <- err
		return
	}
	err = table.Render()
	if err != nil {
		if debug || onlyVerbose {
			color.Red(err.Error())
		}
		errCh <- err
		return
	}

}

// install installs a go version
func install(app *Application, wg *sync.WaitGroup, errCh chan error, version string) {
	defer wg.Done()

	verbose, debug := *app.figs.Bool(kVerbose), *app.figs.Bool(kDebug)
	// onlyVerbose, onlyDebug := verbose && !debug, !verbose && debug

	if verbose {
		color.Green("VERBOSE MODE ENABLED")
	}

	if debug {
		color.Red("DEBUG MODE ENABLED")
	}

	workspace := app.Workspace()
	if verbose {
		color.Green("Using workspace: %v", workspace)
	}

	_, workspaceErr := os.Stat(workspace)
	if os.IsNotExist(workspaceErr) {
		internal.Capture(os.MkdirAll(workspace, 0755))
		if verbose {
			color.Green("Create workspace directory: %v", workspace)
		}
	}

	var (
		binDir       = filepath.Join(workspace, "bin")
		pathDir      = filepath.Join(workspace, "path")
		rootDir      = filepath.Join(workspace, "root")
		cacheDir     = filepath.Join(workspace, "cache")
		telemetryDir = filepath.Join(workspace, "telemetry")
		shimDir      = filepath.Join(workspace, "shims")
		scriptsDir   = filepath.Join(workspace, "scripts")
		versionDir   = filepath.Join(workspace, "versions", version)
		modCacheDir  = filepath.Join(workspace, "versions", version, "go", "pkg", "mod")
	)
	_, shimsErr := os.Stat(shimDir)
	if os.IsNotExist(shimsErr) {
		internal.Capture(os.MkdirAll(shimDir, 0755))
		if verbose {
			color.Green("Create shim directory: %v", shimDir)
		}
	}

	// define the environment that igo requires
	envs := map[string]string{
		GOOS:           *app.figs.String(kGoos),
		GOARCH:         *app.figs.String(kGoArch),
		GOSCRIPTS:      scriptsDir,
		GOSHIMS:        shimDir,
		GOBIN:          binDir,
		GOROOT:         rootDir,
		GOPATH:         pathDir,
		GOMODCACHE:     modCacheDir,
		GOTELEMETRYDIR: telemetryDir,
		GOCACHE:        cacheDir,
	}

	installerLockFile := filepath.Join(workspace, "installer.lock")
	versionLockFile := filepath.Join(versionDir, "installer.lock")
	tarball := fmt.Sprintf("go%s.%s-%s.tar.gz", version, envs[GOOS], envs[GOARCH])
	downloadsDir := filepath.Join(workspace, "downloads")
	versionsDir := filepath.Join(workspace, "versions")

	// create a new version struct to download the assets into the location needed
	versionData := Version{
		Version:      version,
		DownloadName: tarball,
		TarPath:      filepath.Join(downloadsDir, tarball),
		ExtractPath:  filepath.Join(versionsDir, version),
	}

	// this file protects the runtime of the igo install func - when its present, the script aborts
	_, err := os.Stat(installerLockFile) // check igo runtime installer.lock
	defer func() {
		_, err = os.Stat(installerLockFile)
		if os.IsNotExist(err) {
			return
		}
		if os.IsPermission(err) {
			internal.Capture(err)
		}
		err = os.Remove(installerLockFile)
		if err != nil {
			internal.Capture(err)
		}
	}()
	if os.IsExist(err) { // installer.lock exists
		errCh <- fmt.Errorf("huh igo is already running")
		return
	}

	_, err = os.Stat(versionLockFile)
	if os.IsExist(err) {
		errCh <- fmt.Errorf("version is already installed")
		return
	}

	// write the current version to the lockFile
	internal.Capture(os.WriteFile(installerLockFile, []byte(version), 0644))
	if verbose {
		color.Green("Created igo lockfile at %v", installerLockFile)
	}

	// create the downloads directory
	_, err = os.Stat(downloadsDir)
	if os.IsNotExist(err) {
		internal.Capture(os.MkdirAll(downloadsDir, 0755))
		if verbose {
			color.Green("Created directory %s", downloadsDir)
		}
	}

	// check if the download exists
	_, tarErr := os.Stat(filepath.Join(downloadsDir, tarball))
	if os.IsNotExist(tarErr) {
		// download the tar.gz
		internal.Capture(versionData.downloadURL(app))
		if verbose {
			color.Green("Download file %s to %s", tarball, downloadsDir)
		}
	}

	// create if not exists the version extract destination
	_, err = os.Stat(versionData.ExtractPath)
	if os.IsNotExist(err) {
		internal.Capture(os.MkdirAll(versionData.ExtractPath, 0755))
		if verbose {
			color.Green("Created directory %s", versionData.ExtractPath)
		}
	}

	// extract the tar.gz into the destination
	internal.Capture(versionData.extractTarGz(app))
	if verbose {
		color.Green("Extracted %s to %s", versionData.DownloadName, versionData.ExtractPath)
	}

	// move go to go.version in the version dir
	_old := filepath.Join(versionDir, "go", "bin", "go")
	_new := filepath.Join(versionDir, "go", "bin", "go."+version)
	internal.Capture(os.Rename(_old, _new))
	if verbose {
		color.Green("Renamed %s to %s", _old, _new)
	}

	// move gofmt to gofmt.version in the version dir
	_old = filepath.Join(versionDir, "go", "bin", "gofmt")
	_new = filepath.Join(versionDir, "go", "bin", "gofmt."+version)
	internal.Capture(os.Rename(_old, _new))
	if verbose {
		color.Green("Renamed %s to %s", _old, _new)
	}

	// create a symlink in version dir to shim go
	err = app.CreateShims()
	if err != nil {
		errCh <- err
		return
	}

	// if GOROOT is a directory, move it to root.bak in the app.Workspace()
	internal.Capture(internal.RemoveSymlinkOrBackupPath(rootDir))

	// symlink for GOROOT to version go directory
	src := filepath.Join(versionDir, "go")
	tar := rootDir
	internal.Capture(os.Symlink(src, tar))
	if verbose {
		color.Green("Created symlink %s to %s", src, tar)
	}

	// if GOBIN is a directory, move it to bin.bak in the app.Workspace()
	internal.Capture(internal.RemoveSymlinkOrBackupPath(binDir))

	// symlink for GOBIN to version go directory
	src = filepath.Join(versionDir, "go", "bin")
	tar = binDir
	internal.Capture(os.Symlink(src, tar))
	if verbose {
		color.Green("Created symlink %s -> %s", src, tar)
	}

	internal.Capture(internal.RemoveSymlinkOrBackupPath(pathDir))

	// symlink for GOPATH to version go directory
	src = strings.Clone(versionDir)
	tar = pathDir
	internal.Capture(os.Symlink(src, tar))
	if verbose {
		color.Green("Created symlink %s -> %s", src, tar)
	}

	// add GOBIN/GOROOT/GOOS/GOARCH/GOPATH to ~/.zshrc or ~/.bashrc
	internal.Capture(app.injectEnvVarsToShellConfig(envs))
	if verbose || debug {
		color.Green("Patched igo variables in ENV")
		for name, value := range envs {
			color.Green("   %s=%s\n", name, value)
		}
	}

	// update PATH in ~/.zshrc and ~/.bashrc to use GOSHIMS and GOBIN directories before PATH
	internal.Capture(app.patchShellConfigPath(envs))
	if verbose || debug {
		color.Green("Patched PATH in shell configs!")
	}

	// read the text printed in the "go version" for this version
	dataInVersionFile := app.runVersionCheck(envs, version)
	if verbose || debug {
		color.Green("Found data in version file response: %v", dataInVersionFile)
	}

	// validate the format matches
	if !strings.Contains(strings.TrimSpace(dataInVersionFile), version) {
		e := fmt.Errorf("failed check - mismatched versions got = %s ; wanted = %s", dataInVersionFile, version)
		if verbose || debug {
			color.Red("Received Err: %v", e)
		}
		errCh <- e
		return
	}
	if verbose {
		color.Green("Verified that the correct version of Go was just installed and it works!")
	}

	// open the version file
	versionFile := filepath.Join(workspace, "version")
	fileHandler := internal.CaptureOpenFile(versionFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if verbose {
		color.Green("Opened the %v", versionFile)
	}

	// write the current version
	internal.CaptureInt(fileHandler.Write([]byte(version)))
	if verbose {
		color.Green("Wrote '%s' to %s", version, versionFile)
	}

	internal.Capture(os.MkdirAll(telemetryDir, 0755))
	internal.Capture(os.MkdirAll(cacheDir, 0755))

	// report back to the user
	if verbose {
		color.Green("Assigned the igo version to %v", version)
	}

	// close the current version file handler
	internal.Capture(fileHandler.Close())

	// install extra packages on the system
	internal.Capture(app.installExtraPackages(envs, version))
	if verbose {
		color.Green("Installed extra packages successfully!")
	}

	internal.Capture(internal.SetStickyBit(versionDir))
	internal.Capture(internal.SetSetuidSetgidBits(versionDir))

	// write a lockfile to the version directory to prevent future changes by this script
	internal.Capture(internal.Touch(versionLockFile))
	if verbose {
		color.Green("Locked version of go with locker file at %v", versionLockFile)
	}

	// when we're finished, remove the installer.lock file
	internal.Capture(os.Remove(installerLockFile))
	if verbose {
		color.Green("Removed the igo runtime locker at %v", installerLockFile)
	}
}
