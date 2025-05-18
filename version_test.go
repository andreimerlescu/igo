package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestVersion_String(t *testing.T) {
	v := Version{
		Version:      "1.20.0",
		DownloadName: "go1.20.0.linux-amd64.tar.gz",
		ExtractPath:  "/tmp/igo/versions/1.20.0",
		TarPath:      "/tmp/igo/downloads/go1.20.0.linux-amd64.tar.gz",
	}
	result := v.String()
	assert.Contains(t, result, "Version 1.20.0")
	assert.Contains(t, result, "go1.20.0.linux-amd64.tar.gz")
}

func TestVersion_downloadURL(t *testing.T) {
	testDir := t.TempDir()
	downloadsDir := filepath.Join(testDir, "downloads")
	err := os.MkdirAll(downloadsDir, 0755)
	assert.NoError(t, err)
	v := Version{
		Version:      "1.20.0",
		DownloadName: "go1.20.0.linux-amd64.tar.gz",
		TarPath:      filepath.Join(downloadsDir, "go1.20.0.linux-amd64.tar.gz"),
	}
	os.Args = []string{os.Args[0], "-" + kVerbose, "-" + kDebug}
	app := NewApp()
	app.figs.StoreBool(kVerbose, true)
	app.figs.StoreBool(kDebug, true)
	originalHTTPGet := httpGet
	defer func() { httpGet = originalHTTPGet }()
	httpGet = func(url string) (*http.Response, error) {
		assert.Equal(t, "https://go.dev/dl/go1.20.0.linux-amd64.tar.gz", url)
		mockContent := []byte("mock tarball content")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(mockContent)),
		}, nil
	}
	err = v.downloadURL(app)
	assert.NoError(t, err)
	content, err := os.ReadFile(v.TarPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("mock tarball content"), content)
}

func TestVersion_extractTarGz(t *testing.T) {
	testDir := t.TempDir()
	downloadsDir := filepath.Join(testDir, "downloads")
	extractDir := filepath.Join(testDir, "versions", "1.20.0")
	err := os.MkdirAll(downloadsDir, 0755)
	assert.NoError(t, err)
	tarPath := filepath.Join(downloadsDir, "go1.20.0.linux-amd64.tar.gz")
	err = createMockTarGz(tarPath)
	assert.NoError(t, err)
	v := Version{
		Version:      "1.20.0",
		DownloadName: "go1.20.0.linux-amd64.tar.gz",
		TarPath:      tarPath,
		ExtractPath:  extractDir,
	}
	os.Args = []string{os.Args[0], "-" + kVerbose, "-" + kDebug}
	app := NewApp()
	err = v.extractTarGz(app)
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(extractDir, "go", "bin", "go"))
	assert.FileExists(t, filepath.Join(extractDir, "go", "README.md"))
}

func createMockTarGz(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	mockFiles := map[string]string{
		"go/README.md": "This is a test README",
		"go/bin/go":    "#!/bin/sh\necho Hello, world!",
	}
	for path, content := range mockFiles {
		header := &tar.Header{
			Name:    path,
			Mode:    0644,
			Size:    int64(len(content)),
			ModTime: time.Now(),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			return err
		}
	}
	return nil
}
