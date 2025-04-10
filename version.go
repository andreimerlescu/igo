package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andreimerlescu/checkfs"
	"github.com/andreimerlescu/checkfs/directory"
	"github.com/andreimerlescu/checkfs/file"
	"github.com/fatih/color"
)

type Version struct {
	Major, Minor, Patch int
	DownloadToPath      string
	Version             string
	ExtractToPath       string
}

func (v *Version) downloadURL(ctx context.Context) (err error) {
	color.Blue("Starting download of %s", v.DownloadToPath)
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if err != nil {
			color.Red("Failed to download %s in %v: %v", v.DownloadToPath, duration, err)
		} else {
			color.Green("Downloaded %s in %v", v.DownloadToPath, duration)
		}
	}()

	if err = checkfs.File(v.DownloadToPath, file.Options{Exists: true}); err == nil {
		color.Yellow("Skipping %s: file already exists", v.DownloadToPath)
		return nil
	}

	fullURL := strings.Clone(v.DownloadToPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	capture(err)

	client := &http.Client{}
	resp, err := client.Do(req)
	capture(err)
	defer capture(resp.Body.Close())

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	out, err := os.Create(v.ExtractToPath)
	capture(err)
	defer capture(out.Close())

	_, err = io.Copy(out, resp.Body)
	capture(err)
	return nil
}

func (v *Version) extractTarGz() error {
	// Open the .tar.gz tarFile
	tarFile, err := os.Open(v.DownloadToPath)
	if err != nil {
		return fmt.Errorf("error opening tar.gz tarFile: %v", err)
	}
	defer capture(tarFile.Close())

	// Create gzip reader
	gzReader, err := gzip.NewReader(tarFile)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %v", err)
	}
	defer capture(gzReader.Close())

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Ensure destination directory exists
	capture(checkfs.Directory(v.ExtractToPath, directory.Options{
		Exists:     true,
		WillCreate: true,
		Create: directory.Create{
			Kind:     directory.IfNotExists,
			FileMode: 0755,
			Path:     v.ExtractToPath,
		},
	}))

	// Iterate through the files in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("error reading tar: %v", err)
		}

		// Get the target path for this tarFile
		target := filepath.Join(v.ExtractToPath, header.Name)

		// Check the tarFile type
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("error creating directory %s: %v", target, err)
			}

		case tar.TypeReg:
			// Create directories for the tarFile if they don't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("error creating directory for tarFile %s: %v", target, err)
			}

			// Create and write to the tarFile
			outFile := captureOpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))

			// Copy the tarFile contents
			if _, err := io.Copy(outFile, tarReader); err != nil {
				capture(outFile.Close())
				return fmt.Errorf("error writing to tarFile %s: %v", target, err)
			}
			capture(outFile.Close())

		default:
			fmt.Printf("Skipping unsupported type %c in %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}
