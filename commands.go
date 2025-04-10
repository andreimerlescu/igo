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

func install(ctx context.Context, wg *sync.WaitGroup, errCh chan error, version string) {
	defer wg.Done()

	workspace := igoWorkspace()
	defer capture(os.Remove(filepath.Join(workspace, "installer.lock")))

	envs := map[string]string{
		"GOOS":    *figs.String(kGoos),
		"GOARCH":  *figs.String(kGoArch),
		"GOSHIMS": filepath.Join(workspace, "shims"),
		"GOBIN":   filepath.Join(workspace, "bin"),
		"GOROOT":  filepath.Join(workspace, "root"),
		"GOPATH":  filepath.Join(workspace, "path"),
	}

	tarball := fmt.Sprintf("go%s.%s-%s.tar.gz", envs["GOOS"], envs["GOARCH"], envs["GOBIN"])

	versionData := Version{
		DownloadToPath: filepath.Join(workspace, "downloads", tarball),
		Version:        version,
		ExtractToPath:  "",
	}

	lockFileHandler := captureOpenFile(filepath.Join(workspace, "installer.lock"), // path
		os.O_CREATE|os.O_EXCL|os.O_RDWR|os.O_TRUNC, // flags
		0600) // mode
	captureInt(lockFileHandler.Write([]byte(*figs.String(kGoVersion))))
	defer capture(lockFileHandler.Close())

	// is it already installed?
	capture(checkfs.Directory(
		filepath.Join(workspace, "versions", *figs.String(kGoVersion), "installed.lock"),
		directory.Options{Exists: true}))

	discard(os.Setenv("GOBIN", envs["GOBIN"]),
		os.Setenv("GOPATH", envs["GOPATH"]),
		os.Setenv("GOROOT", envs["GOROOT"]),
		os.Setenv("GOSHIMS", envs["GOSHIMS"]),
		os.Setenv("GOOS", envs["GOOS"]),
		os.Setenv("GOARCH", envs["GOARCH"]))

	goVersionFileHandler := captureOpenFile(filepath.Join(workspace, "version"), // path
		os.O_CREATE|os.O_EXCL|os.O_RDWR|os.O_TRUNC, // flags
		0600) // mode
	captureInt(goVersionFileHandler.Write([]byte(*figs.String(kGoVersion))))
	capture(goVersionFileHandler.Close())

	capture(os.MkdirAll(filepath.Join(workspace, "downloads"), 0755))

	_, tarErr := os.Stat(filepath.Join(workspace, "downloads", tarball))
	if os.IsNotExist(tarErr) {
		capture(versionData.downloadURL(ctx))
	}

	capture(checkfs.Directory(filepath.Join(workspace, "versions", versionData.Version),
		directory.Options{
			WillCreate: true,
			Create: directory.Create{
				Path:     versionData.ExtractToPath,
				Kind:     directory.IfNotExists,
				FileMode: 0755,
			}}))

	capture(versionData.extractTarGz())

	capture(os.Rename(filepath.Join(workspace, "versions", versionData.Version, "go", "bin", "go"), // old
		filepath.Join(workspace, "versions", versionData.Version, "go", "bin", "go."+versionData.Version))) // new

	capture(os.Rename(filepath.Join(workspace, "versions", versionData.Version, "go", "bin", "gofmt"), // old
		filepath.Join(workspace, "versions", versionData.Version, "go", "bin", "gofmt."+versionData.Version))) // new

	capture(os.Symlink(filepath.Join(workspace, "shims", "go"), // source
		filepath.Join(workspace, "versions", version, "go", "bin", "go"))) // target

	capture(os.Symlink(filepath.Join(workspace, "shims", "gofmt"), // source
		filepath.Join(workspace, "versions", version, "go", "bin", "gofmt"))) // target

	// symlink for GOROOT
	capture(os.Symlink(filepath.Join(workspace, "versions", version, "go"), // source
		filepath.Join(workspace, "root"))) // target

	capture(backupIfNotSymlink(filepath.Join(workspace, "bin")))

	capture(os.Symlink(filepath.Join(workspace, "versions", version, "go", "bin"), // source
		filepath.Join(workspace, "bin"))) // target

	capture(os.Symlink(filepath.Join(workspace, "versions", version), // source
		filepath.Join(workspace, "path"))) // target

	capture(injectEnvVarsToShellConfig(envs))

	capture(patchShellConfigPath(envs))

	dataInVersionFile := getGoVersionOutput(envs, version)
	if !strings.EqualFold(strings.TrimSpace(dataInVersionFile), version) {
		errCh <- fmt.Errorf("failed sanity check - mismatched go versions got = %s ; wanted = %s", dataInVersionFile, version)
		return
	}

	capture(installExtraPackages(envs, dataInVersionFile))

	capture(touch(filepath.Join(workspace, "versions", version, "installer.lock")))

}
