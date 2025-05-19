package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreimerlescu/figtree/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "igo-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	origHomeDir := UserHomeDir
	defer func() {
		UserHomeDir = origHomeDir
	}()
	UserHomeDir = func() (string, error) {
		return tempDir, nil
	}

	app := NewApp()
	assert.NotNil(t, app.ctx)
	assert.NotNil(t, app.figs)
	assert.Equal(t, tempDir, app.userHomeDir)

	assert.Equal(t, "1.24.3", *app.figs.String(kGoVersion))
	assert.Equal(t, filepath.Join(tempDir, "go"), *app.figs.String(kGoDir))
}

func TestApplication_validateVersion(t *testing.T) {
	app := &Application{
		ctx:         context.Background(),
		userHomeDir: "/tmp",
		figs:        figtree.With(figtree.Options{}),
	}
	app.figs.NewString(kGoos, "linux", "")
	app.figs.NewString(kGoArch, "amd64", "")

	t.Run("valid version", func(t *testing.T) {
		originalHttpHead := http.Head
		httpHead = func(url string) (resp *http.Response, err error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
		defer func() { httpHead = originalHttpHead }()

		err := app.validateVersion("1.20.0")
		assert.NoError(t, err)
	})

	t.Run("invalid format", func(t *testing.T) {
		err := app.validateVersion("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid go version format")
	})

	t.Run("version too old", func(t *testing.T) {
		err := app.validateVersion("1.15.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("version not found", func(t *testing.T) {
		originalHttpHead := http.Head
		httpHead = func(url string) (resp *http.Response, err error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}
		defer func() { httpHead = originalHttpHead }()

		err := app.validateVersion("9.9.9")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found on download server")
	})
}

func TestApplication_Workspace(t *testing.T) {
	homeDir := "/home/testuser"
	app := NewApp()
	app.Workspace = func() string {
		return filepath.Join(homeDir, "go")
	}
	t.Run("user workspace", func(t *testing.T) {
		expected := filepath.Join(homeDir, "go")
		assert.Equal(t, expected, app.Workspace())
	})
}

func TestApplication_CreateShims(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "igo-shim-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	shimsDir := filepath.Join(tempDir, "shims")
	err = os.MkdirAll(shimsDir, 0755)
	require.NoError(t, err)

	app := &Application{
		ctx:         context.Background(),
		userHomeDir: tempDir,
		figs:        figtree.With(figtree.Options{}),
	}
	app.figs.NewBool(kSystem, false, "")

	origWorkspaceFn := app.Workspace

	app.Workspace = func() string {
		return tempDir
	}
	defer func() { app.Workspace = origWorkspaceFn }()

	err = app.CreateShims()
	assert.NoError(t, err)

	goShimPath := filepath.Join(shimsDir, "go")
	gofmtShimPath := filepath.Join(shimsDir, "gofmt")

	_, err = os.Stat(goShimPath)
	assert.NoError(t, err, "go shim should exist")

	_, err = os.Stat(gofmtShimPath)
	assert.NoError(t, err, "gofmt shim should exist")

	goShimInfo, err := os.Stat(goShimPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), goShimInfo.Mode().Perm())

	gofmtShimInfo, err := os.Stat(gofmtShimPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), gofmtShimInfo.Mode().Perm())
}

func TestApplication_findGoVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "igo-versions-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	versionsDir := filepath.Join(tempDir, "versions")
	require.NoError(t, os.MkdirAll(filepath.Join(versionsDir, "1.20.0"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(versionsDir, "1.21.3"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(versionsDir, "non-version-dir"), 0755))

	app := &Application{
		ctx:         context.Background(),
		userHomeDir: tempDir,
		figs:        figtree.With(figtree.Options{}),
	}

	origWorkspaceFn := app.Workspace
	app.Workspace = func() string {
		return tempDir
	}
	defer func() { app.Workspace = origWorkspaceFn }()

	versions, err := app.findGoVersions()
	assert.NoError(t, err)

	assert.Contains(t, versions, "1.20.0")
	assert.Contains(t, versions, "1.21.3")
	assert.NotContains(t, versions, "non-version-dir")
}

func TestApplication_activatedVersion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "igo-active-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	versionContent := "1.21.3"
	err = os.WriteFile(filepath.Join(tempDir, "version"), []byte(versionContent), 0644)
	require.NoError(t, err)

	app := &Application{
		ctx:         context.Background(),
		userHomeDir: tempDir,
		figs:        figtree.With(figtree.Options{}),
	}

	origWorkspaceFn := app.Workspace
	app.Workspace = func() string {
		return tempDir
	}
	defer func() { app.Workspace = origWorkspaceFn }()

	version, err := app.activatedVersion()
	assert.NoError(t, err)
	assert.Equal(t, "1.21.3", version)

	err = os.Remove(filepath.Join(tempDir, "version"))
	require.NoError(t, err)

	_, err = app.activatedVersion()
	assert.Error(t, err)
}
