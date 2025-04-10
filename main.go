package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/andreimerlescu/figtree/v2"
)

//go:embed VERSION
var binaryVersionBytes embed.FS

var binaryCurrentVersion string

func BinaryVersion() string {
	if len(binaryCurrentVersion) == 0 {
		versionBytes, err := binaryVersionBytes.ReadFile("VERSION")
		capture(err)
		binaryCurrentVersion = strings.TrimSpace(string(versionBytes))
	}
	return binaryCurrentVersion
}

func init() {
	if runtime.GOOS == "windows" {
		panic("windows not supported, please use Go's MSI installers instead")
	}
	home, homeErr := os.UserHomeDir()
	capture(homeErr)
	userHomeDir = home
	figs = figtree.New()
	figs.NewBool(kVersion, false, "Display current version")
	figs.NewBool(kSystem, false, "Install for system-wide usage (ignore USER HOME directory)")
	figs.NewString(kGoDir, filepath.Join(home, "go"), "Path where you want multiple go versions installed")
	figs.NewString(kCommand, "", "Command to run: install uninstall use list")
	figs.NewString(kGoVersion, "1.24.2", "Go Version")
	figs.NewString(kGoos, runtime.GOOS, "Go OS")
	figs.NewString(kGoArch, runtime.GOARCH, "Go Architecture")
	capture(figs.Parse())
	if *figs.Bool(kVersion) {
		fmt.Println(BinaryVersion())
		os.Exit(0)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}
	errCh := make(chan error)
	go func() {
		wg.Wait()
		for err := range errCh {
			capture(err)
		}
	}()
	wg.Add(1)
	switch *figs.String(kCommand) {
	case "ins":
		go install(ctx, wg, errCh, *figs.String(kGoVersion))
	case "install":
		go install(ctx, wg, errCh, *figs.String(kGoVersion))
	case "uni":
		go uninstall(ctx, wg, errCh, *figs.String(kGoVersion))
	case "uninstall":
		go uninstall(ctx, wg, errCh, *figs.String(kGoVersion))
	case "l":
		go list(ctx, wg, errCh)
	case "list":
		go list(ctx, wg, errCh)
	case "u":
		go use(ctx, wg, errCh, *figs.String(kGoVersion))
	case "use":
		go use(ctx, wg, errCh, *figs.String(kGoVersion))
	case "f":
		go fix(ctx, wg, errCh, *figs.String(kGoVersion))
	case "fix":
		go fix(ctx, wg, errCh, *figs.String(kGoVersion))
	}
	wg.Wait()
	close(errCh)

}
