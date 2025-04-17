package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Version stores the paths of the tarball and extract paths for a given version
type Version struct {
	// DownloadName is the tar.gz file of the Version
	DownloadName string
	// ExtractPath is the output destination of extracting the tar.gz file
	ExtractPath string
	// TarPath is the locally saved tarball
	TarPath string
	// Version captures the version of go in the Major.Minor.Patch format
	Version string
}

func (v *Version) String() string {
	out := strings.Builder{}
	out.WriteString(color.GreenString("Version %s {", v.Version) + "\n")
	out.WriteString(color.GreenString("   Download Name: %s", v.DownloadName) + "\n")
	out.WriteString(color.GreenString("    Extract Path: %s", v.ExtractPath) + "\n")
	out.WriteString(color.GreenString("        Tar Path: %s", v.TarPath) + "\n")
	out.WriteString(color.GreenString("}"))
	return out.String()
}

// downloadURL will take the DownloadName and acquire the tar.gz file
func (v *Version) downloadURL(app *Application) (err error) {
	color.Blue("Starting download of %s", v.DownloadName)
	verbose, debug := *app.figs.Bool(kVerbose), *app.figs.Bool(kDebug)
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if err != nil {
			color.Red("Failed to download %s in %v: %v", v.DownloadName, duration, err)
		} else {
			color.Green("Downloaded %s in %v", v.DownloadName, duration)
		}
	}()

	if verbose {
		fmt.Println(v)
	}

	_, err = os.Stat(v.TarPath)
	if os.IsExist(err) {
		color.Yellow("Skipping %s: file already exists", v.DownloadName)
		return nil
	}

	fullURL := "https://go.dev/dl/" + strings.Clone(v.DownloadName)
	if verbose {
		color.Green("Downloading %s", fullURL)
	}

	resp, err := http.Get(fullURL)
	if err != nil {
		color.Red("Failed to download %s: %v", v.DownloadName, err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	out, err := os.Create(v.TarPath)
	if err != nil {
		return err
	}
	defer out.Close()

	total := int64(0)
	total, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	if debug {
		color.Green("Downloaded %d bytes of %s in %v", total, v.ExtractPath, time.Since(startTime))
	}
	return nil
}

// extractTarGz will take the ExtractPath and expand the DownloadName there
func (v *Version) extractTarGz(app *Application) error {
	verbose, debug := *app.figs.Bool(kVerbose), *app.figs.Bool(kDebug)
	if verbose {
		fmt.Println(v)
	}
	// Open the .tar.gz tarFile
	tarFile, err := os.Open(v.TarPath)
	if err != nil {
		return fmt.Errorf("error opening tar.gz tarFile: %v", err)
	}
	defer tarFile.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(tarFile)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %v", err)
	}
	defer capture(gzReader.Close())

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Ensure destination directory exists
	_, err = os.Stat(v.ExtractPath)
	if os.IsNotExist(err) {
		err2 := os.MkdirAll(v.ExtractPath, 0755)
		if err2 != nil {
			return fmt.Errorf("error creating extract dir: %v", errors.Join(err, err2))
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error creating extract dir: %v", err)
	}

	// Iterate through the files in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			if verbose {
				color.Green("Reached end of the .tar.gz file!")
			}
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("error reading tar: %v", err)
		}

		// Get the target path for this tarFile
		target := filepath.Join(v.ExtractPath, header.Name)
		if debug || verbose {
			color.Green("Extracting %s to %s", target, v.ExtractPath)
		}

		// Check the tarFile type
		switch header.Typeflag {
		case tar.TypeDir:
			if debug || verbose {
				color.Green("Creating directory %s", target)
			}
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("error creating directory %s: %v", target, err)
			}

		case tar.TypeReg:
			if debug || verbose {
				color.Green("Extracting %s to %s", target, v.ExtractPath)
			}
			// Create directories for the tarFile if they don't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("error creating directory for tarFile %s: %v", target, err)
			}

			// Create and write to the tarFile
			outFile := captureOpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))

			// Copy the tarFile contents
			if debug || verbose {
				color.Green("Copying tarReader into outFile")
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				if debug {
					color.Red(err.Error())
				}
				_ = outFile.Close()
				return fmt.Errorf("error writing to tarFile %s: %v", target, err)
			}
			capture(outFile.Close())

		default:
			fmt.Printf("Skipping unsupported type %c in %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}
