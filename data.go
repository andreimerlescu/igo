package main

import (
	"github.com/andreimerlescu/figtree/v2"
)

const (
	PRODUCT    string = "igo"
	AUTHOR     string = "github.com/andreimerlescu/igo"
	XRP        string = "rAparioji3FxAtD7UufS8Hh9XmFn7h6AX"
	kVersion   string = "version"
	kGoDir     string = "godir"
	kCommand   string = "cmd"
	kGoVersion string = "gover"
	kSystem    string = "system"
	kGoos      string = "goos"
	kGoArch    string = "goarch"
)

var packages = map[string]string{
	"gotop":                "github.com/cjbassi/gotop",
	"go-generate-password": "github.com/m1/go-generate-password/cmd/go-generate-password",
	"bombardier":           "github.com/codesenberg/bombardier",
}

var (
	figs        figtree.Fruit
	userHomeDir string
)
