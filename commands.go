package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/andreimerlescu/checkfs"
	"github.com/andreimerlescu/checkfs/directory"
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

func list(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()
	versions, err := findGoVersions()
	if err != nil {
		errCh <- err
		return
	}
	slices.Sort(versions)
	slices.Reverse(versions)
	currentVersion, _ := activatedVersion()
	var data [][]string
	for _, version := range versions {
		info, infoErr := os.Stat(filepath.Join(igoWorkspace(), "versions", version))
		if os.IsNotExist(infoErr) {
			errCh <- infoErr
			return
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

	// igo can run in the system or user profile, so igoWorkspace() provides where iGo is installed
	workspace := igoWorkspace()

	// this file protects an installed version of go from getting replaced by this func
	versionLockFile := filepath.Join(workspace, "versions", version, "installer.lock")

	// does an installer.lock file exist in the version path?
	if _, err := os.Stat(versionLockFile); err == nil {
		// File exists
		errCh <- fmt.Errorf("%s is already installed", version)
		return
	} else if !os.IsNotExist(err) {
		// Some other error occurred while checking
		errCh <- fmt.Errorf("failed to check if lock file exists: %w", err)
		return
	}

	// when we're finished, remove the installer.lock file
	defer capture(os.Remove(filepath.Join(workspace, "installer.lock")))

	// define the environment that igo requires
	envs := map[string]string{
		"GOOS":    *app.figs.String(kGoos),
		"GOARCH":  *app.figs.String(kGoArch),
		"GOSHIMS": filepath.Join(workspace, "shims"),
		"GOBIN":   filepath.Join(workspace, "bin"),
		"GOROOT":  filepath.Join(workspace, "root"),
		"GOPATH":  filepath.Join(workspace, "path"),
	}

	// the tarball path
	tarball := fmt.Sprintf("go%s.%s-%s.tar.gz", envs["GOOS"], envs["GOARCH"], envs["GOBIN"])

	//
	versionData := Version{
		DownloadName: filepath.Join(workspace, "downloads", tarball),
		Version:      version,
		ExtractPath:  filepath.Join(workspace, "versions", version),
	}

	lockFileHandler := captureOpenFile(filepath.Join(workspace, "installer.lock"), // path
		os.O_CREATE|os.O_EXCL|os.O_RDWR|os.O_TRUNC, // flags
		0600) // mode
	captureInt(lockFileHandler.Write([]byte(*app.figs.String(kGoVersion))))
	defer capture(lockFileHandler.Close())

	// is it already installed?
	capture(checkfs.Directory(
		filepath.Join(workspace, "versions", *app.figs.String(kGoVersion), "installed.lock"),
		directory.Options{Exists: true}))

	// set the env and discard any warnings
	discard(os.Setenv("GOBIN", envs["GOBIN"]),
		os.Setenv("GOPATH", envs["GOPATH"]),
		os.Setenv("GOROOT", envs["GOROOT"]),
		os.Setenv("GOSHIMS", envs["GOSHIMS"]),
		os.Setenv("GOOS", envs["GOOS"]),
		os.Setenv("GOARCH", envs["GOARCH"]))

	// open the version file
	fileHandler := captureOpenFile(filepath.Join(workspace, "version"),
		os.O_CREATE|os.O_EXCL|os.O_RDWR|os.O_TRUNC, 0600)

	// write the current version
	captureInt(fileHandler.Write([]byte(version)))

	// close the current version file handler
	capture(fileHandler.Close())

	// create the downloads directory
	capture(os.MkdirAll(filepath.Join(workspace, "downloads"), 0755))

	// check if the download exists
	_, tarErr := os.Stat(filepath.Join(workspace, "downloads", tarball))
	if os.IsNotExist(tarErr) {
		// download the tar.gz
		capture(versionData.downloadURL(app.ctx))
	}

	// create if not exists the version extract destination
	capture(checkfs.Directory(filepath.Join(workspace, "versions", versionData.Version),
		directory.Options{
			WillCreate: true,
			Create: directory.Create{
				Path:     versionData.ExtractPath,
				Kind:     directory.IfNotExists,
				FileMode: 0755,
			}}))

	// extract the tar.gz into the destination
	capture(versionData.extractTarGz())

	// move go to go.version in the version dir
	capture(os.Rename(filepath.Join(workspace, "versions", versionData.Version, "go", "bin", "go"), // old
		filepath.Join(workspace, "versions", versionData.Version, "go", "bin", "go."+versionData.Version))) // new

	// move gofmt to gofmt.version in the version dir
	capture(os.Rename(filepath.Join(workspace, "versions", versionData.Version, "go", "bin", "gofmt"), // old
		filepath.Join(workspace, "versions", versionData.Version, "go", "bin", "gofmt."+versionData.Version))) // new

	// create a symlink in version dir to shim go
	capture(os.Symlink(filepath.Join(workspace, "shims", "go"), // source
		filepath.Join(workspace, "versions", version, "go", "bin", "go"))) // target

	// create a symlink in the version dir to shim gofmt
	capture(os.Symlink(filepath.Join(workspace, "shims", "gofmt"), // source
		filepath.Join(workspace, "versions", version, "go", "bin", "gofmt"))) // target

	// symlink for GOROOT to version go directory
	capture(os.Symlink(filepath.Join(workspace, "versions", version, "go"), // source
		filepath.Join(workspace, "root"))) // target

	// if GOBIN is a directory, move it to bin.bak in the igoWorkspace()
	capture(backupIfNotSymlink(filepath.Join(workspace, "bin")))

	// symlink for GOBIN to version go directory
	capture(os.Symlink(filepath.Join(workspace, "versions", version, "go", "bin"), // source
		filepath.Join(workspace, "bin"))) // target

	// symlink for GOPATH to version go directory
	capture(os.Symlink(filepath.Join(workspace, "versions", version), // source
		filepath.Join(workspace, "path"))) // target

	// add GOBIN/GOROOT/GOOS/GOARCH/GOPATH to ~/.zshrc or ~/.bashrc
	capture(injectEnvVarsToShellConfig(envs))

	// update PATH in ~/.zshrc and ~/.bashrc to use GOSHIMS and GOBIN directories before PATH
	capture(patchShellConfigPath(envs))

	// read the text printed in the "go version" for this version
	dataInVersionFile := getGoVersionOutput(envs, version)
	// validate the format matches
	if !strings.EqualFold(strings.TrimSpace(dataInVersionFile), version) {
		errCh <- fmt.Errorf("failed sanity check - mismatched go versions got = %s ; wanted = %s", dataInVersionFile, version)
		return
	}

	// install extra packages on the system
	capture(installExtraPackages(envs, dataInVersionFile))

	// write a lockfile to the version directory to prevent future changes by this script
	capture(touch(versionLockFile))
}
