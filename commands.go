package main

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/olekukonko/tablewriter"
)

func fix(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, version string) {
	defer wg.Done()
	panic("not implemented")
}

func use(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, version string) {
	defer wg.Done()
	panic("not implemented")
}

func uninstall(ctx context.Context, wg *sync.WaitGroup, errCh chan error, version string) {
	defer wg.Done()
	panic("not implemented")
}

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
	fmt.Println(about())
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Version", "Creation", "Status"})
	table.SetBorder(true) // Set Border to false

	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor})

	table.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlueColor},
		tablewriter.Colors{tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgRedColor})

	table.AppendBulk(data)
	table.Render()

}

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
		capture(os.MkdirAll(workspace, 0755))
		if verbose {
			color.Green("Create workspace directory: %v", workspace)
		}
	}

	shimDir := filepath.Join(workspace, "shims")
	_, shimsErr := os.Stat(shimDir)
	if os.IsNotExist(shimsErr) {
		capture(os.MkdirAll(shimDir, 0755))
		if verbose {
			color.Green("Create shim directory: %v", shimDir)
		}
	}
	binDir := filepath.Join(workspace, "bin")
	pathDir := filepath.Join(workspace, "path")
	rootDir := filepath.Join(workspace, "root")

	versionDir := filepath.Join(workspace, "versions", version)

	// define the environment that igo requires
	envs := map[string]string{
		"GOOS":    *app.figs.String(kGoos),
		"GOARCH":  *app.figs.String(kGoArch),
		"GOSHIMS": shimDir,
		"GOBIN":   binDir,
		"GOROOT":  rootDir,
		"GOPATH":  pathDir,
	}

	installerLockFile := filepath.Join(workspace, "installer.lock")
	versionLockFile := filepath.Join(versionDir, "installer.lock")
	tarball := fmt.Sprintf("go%s.%s-%s.tar.gz", version, envs["GOOS"], envs["GOARCH"])
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
	defer capture(os.Remove(installerLockFile))
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
	capture(os.WriteFile(installerLockFile, []byte(version), 0644))
	if verbose {
		color.Green("Created igo lockfile at %v", installerLockFile)
	}

	// create the downloads directory
	_, err = os.Stat(downloadsDir)
	if os.IsNotExist(err) {
		capture(os.MkdirAll(downloadsDir, 0755))
		if verbose {
			color.Green("Created directory %s", downloadsDir)
		}
	}

	// check if the download exists
	_, tarErr := os.Stat(filepath.Join(downloadsDir, tarball))
	if os.IsNotExist(tarErr) {
		// download the tar.gz
		capture(versionData.downloadURL(app))
		if verbose {
			color.Green("Download file %s to %s", tarball, downloadsDir)
		}
	}

	// create if not exists the version extract destination
	_, err = os.Stat(versionData.ExtractPath)
	if os.IsNotExist(err) {
		capture(os.MkdirAll(versionData.ExtractPath, 0755))
		if verbose {
			color.Green("Created directory %s", versionData.ExtractPath)
		}
	}

	// extract the tar.gz into the destination
	capture(versionData.extractTarGz(app))
	if verbose {
		color.Green("Extracted %s to %s", versionData.DownloadName, versionData.ExtractPath)
	}

	// move go to go.version in the version dir
	_old := filepath.Join(versionDir, "go", "bin", "go")
	_new := filepath.Join(versionDir, "go", "bin", "go."+version)
	capture(os.Rename(_old, _new))
	if verbose {
		color.Green("Renamed %s to %s", _old, _new)
	}

	// move gofmt to gofmt.version in the version dir
	_old = filepath.Join(versionDir, "go", "bin", "gofmt")
	_new = filepath.Join(versionDir, "go", "bin", "gofmt."+version)
	capture(os.Rename(_old, _new))
	if verbose {
		color.Green("Renamed %s to %s", _old, _new)
	}

	// create a symlink in version dir to shim go
	src := filepath.Join(workspace, "shims", "go")
	tar := filepath.Join(versionDir, "go", "bin", "go")
	capture(os.Symlink(src, tar))
	if verbose {
		color.Green("Created symlink %s to %s", src, tar)
	}

	// create a symlink in the version dir to shim gofmt
	src = filepath.Join(workspace, "shims", "gofmt")
	tar = filepath.Join(versionDir, "go", "bin", "gofmt")
	capture(os.Symlink(src, tar))
	if verbose {
		color.Green("Created symlink %s to %s", src, tar)
	}

	// symlink for GOROOT to version go directory
	src = filepath.Join(versionDir, "go")
	tar = filepath.Join(workspace, "root")
	capture(os.Symlink(src, tar))
	if verbose {
		color.Green("Created symlink %s to %s", src, tar)
	}

	// if GOBIN is a directory, move it to bin.bak in the igoWorkspace()
	capture(backupIfNotSymlink(filepath.Join(workspace, "bin")))

	// symlink for GOBIN to version go directory
	src = filepath.Join(versionDir, "go", "bin")
	tar = filepath.Join(workspace, "bin")
	capture(os.Symlink(src, tar))
	if verbose {
		color.Green("Created symlink %s -> %s", src, tar)
	}

	// symlink for GOPATH to version go directory
	src = strings.Clone(versionDir)
	tar = filepath.Join(workspace, "path")
	capture(os.Symlink(src, tar))
	if verbose {
		color.Green("Created symlink %s -> %s", src, tar)
	}

	// add GOBIN/GOROOT/GOOS/GOARCH/GOPATH to ~/.zshrc or ~/.bashrc
	capture(app.injectEnvVarsToShellConfig(envs))
	if verbose {
		color.Green("Patched igo variables in ENV")
	}

	// update PATH in ~/.zshrc and ~/.bashrc to use GOSHIMS and GOBIN directories before PATH
	capture(app.patchShellConfigPath(envs))
	if verbose {
		color.Green("Patched PATH in shell configs!")
	}

	// read the text printed in the "go version" for this version
	dataInVersionFile := app.runVersionCheck(envs, version)
	if verbose {
		color.Green("Found data in version file response: %v", dataInVersionFile)
	}

	// validate the format matches
	if !strings.EqualFold(strings.TrimSpace(dataInVersionFile), version) {
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
	fileHandler := captureOpenFile(installerLockFile, os.O_CREATE|os.O_EXCL|os.O_RDWR|os.O_TRUNC, 0600)
	if verbose {
		color.Green("Opened the %v", installerLockFile)
	}

	// write the current version
	captureInt(fileHandler.Write([]byte(version)))
	if verbose {
		color.Green("Wrote %s to %s", version, installerLockFile)
	}

	// report back to the user
	if verbose {
		color.Green("Assigned the igo version to %v", version)
	}

	// close the current version file handler
	capture(fileHandler.Close())

	// install extra packages on the system
	capture(app.installExtraPackages(envs, dataInVersionFile))
	if verbose {
		color.Green("Installed extra packages successfully!")
	}

	// write a lockfile to the version directory to prevent future changes by this script
	capture(touch(versionLockFile))
	if verbose {
		color.Green("Locked version of go with locker file at %v", versionLockFile)
	}

	// when we're finished, remove the installer.lock file
	capture(os.Remove(installerLockFile))
	if verbose {
		color.Green("Removed the igo runtime locker at %v", installerLockFile)
	}
}
