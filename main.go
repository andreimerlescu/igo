package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/andreimerlescu/figtree/v2"
)

func main() {
	if runtime.GOOS == "windows" {
		panic("windows not supported, please use Go's MSI installers instead")
	}
	app := NewApp()
	app.figs = figtree.New()
	app.figs.NewBool(kVersion, false, "Display current version")
	app.figs.NewBool(kSystem, false, "Install for system-wide usage (ignore USER HOME directory)")
	app.figs.NewBool(kDebug, false, "Enable debug mode")
	app.figs.NewBool(kVerbose, false, "Enable verbose mode")
	app.figs.NewString(kGoDir, filepath.Join(app.userHomeDir, "go"), "Path where you want multiple go versions installed")
	app.figs.NewString(kCommand, "", "Command to run: install uninstall use list")
	app.figs.NewString(kGoVersion, "1.24.2", "Go Version")
	app.figs.NewString(kGoos, runtime.GOOS, "Go OS")
	app.figs.NewString(kGoArch, runtime.GOARCH, "Go Architecture")
	capture(app.figs.Parse())
	if *app.figs.Bool(kVersion) {
		fmt.Println(BinaryVersion())
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx
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
	switch *app.figs.String(kCommand) {
	case "ins":
		go install(app, wg, errCh, *app.figs.String(kGoVersion))
	case "install":
		go install(app, wg, errCh, *app.figs.String(kGoVersion))
	case "uni":
		go uninstall(ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "uninstall":
		go uninstall(ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "l":
		go list(app, ctx, wg, errCh)
	case "list":
		go list(app, ctx, wg, errCh)
	case "u":
		go use(ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "use":
		go use(ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "f":
		go fix(ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "fix":
		go fix(ctx, wg, errCh, *app.figs.String(kGoVersion))
	}
	wg.Wait()
	close(errCh)

}
