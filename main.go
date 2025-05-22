package main

import (
	"context"
	"fmt"
	"github.com/andreimerlescu/igo/internal"
	"os"
	"runtime"
	"sync"
)

func main() {
	if runtime.GOOS == "windows" {
		panic("windows not supported, please use Go's MSI installers instead")
	}
	app := NewApp()

	if *app.figs.Bool(kVersion) {
		fmt.Println(BinaryVersion())
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx
	defer cancel()
	wg := &sync.WaitGroup{}
	errCh := make(chan error)
	wg.Add(1)
	go func() {
		for err := range errCh {
			internal.Capture(err)
		}
	}()
	switch *app.figs.String(kCommand) {
	case "env":
		go env(app, ctx, wg, errCh)
	case "ins":
		go install(app, wg, errCh, *app.figs.String(kGoVersion))
	case "install":
		go install(app, wg, errCh, *app.figs.String(kGoVersion))
	case "uni":
		go uninstall(app, ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "uninstall":
		go uninstall(app, ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "l":
		go list(app, ctx, wg, errCh)
	case "list":
		go list(app, ctx, wg, errCh)
	case "u":
		go use(app, ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "use":
		go use(app, ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "f":
		go fix(app, ctx, wg, errCh, *app.figs.String(kGoVersion))
	case "fix":
		go fix(app, ctx, wg, errCh, *app.figs.String(kGoVersion))
	}
	wg.Wait()
	close(errCh)

}
