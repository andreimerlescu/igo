package main

import (
	"github.com/andreimerlescu/go-common/version"
	"github.com/andreimerlescu/igo/internal"
	"github.com/fatih/color"
	"log"
	"os"
	"runtime"
)

func main() {
	if runtime.GOOS == "windows" {
		panic("windows not supported, please use Go's MSI installers instead")
	}
	app := NewApp()

	if *app.Figs.Bool(cmdVersion) {
		color.Magenta(BinaryVersion() + " - " + internal.About())
		os.Exit(0)
	}
	if *app.Figs.Bool(cmdList) {
		list(app)
		return
	}
	if *app.Figs.Bool(cmdEnv) {
		env(app)
		return
	}
	maybeVersions := map[string]string{
		"install":   *app.Figs.String(cmdInstall),
		"uninstall": *app.Figs.String(cmdUninstall),
		"fix":       *app.Figs.String(cmdFix),
		"activate":  *app.Figs.String(cmdActivate),
		"switch":    *app.Figs.String(cmdSwitch),
	}
	for command, maybeVersion := range maybeVersions {
		if len(maybeVersion) == 0 {
			continue
		}
		if err := app.validateVersion(maybeVersion); err != nil {
			if *app.Figs.Bool(kVerbose) || *app.Figs.Bool(kDebug) {
				color.Red("ErrBadVersion(%T %s): %w", maybeVersion, maybeVersion, err)
			}
			log.Fatalf("ErrBadVersion(%T %s): %s", maybeVersion, maybeVersion, err.Error())
		}
		version := version.FromString(maybeVersion)
		if version.String() == "v0.0.1" {
			log.Fatalf("failed to parse the version: %s", version.String())
		}
		switch command {
		case "install":
			if len(maybeVersion) > 0 {
				install(app, maybeVersion)
			}
		case "uninstall":
			if len(maybeVersion) > 0 {
				uninstall(app, maybeVersion)
			}
		case "fix":
			if len(maybeVersion) > 0 {
				fix(app, maybeVersion)
			}
		default:
			if len(maybeVersion) > 0 {
				use(app, maybeVersion)
			}
		}
	}

}
