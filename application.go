package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andreimerlescu/figtree/v2"
	"github.com/fatih/color"
)

type Application struct {
	ctx         context.Context
	figs        figtree.Fruit
	userHomeDir string
}

func NewApp() *Application {
	userHomeDir, err := os.UserHomeDir()
	capture(err)
	return &Application{
		ctx:         context.Background(),
		userHomeDir: userHomeDir,
	}
}

// Workspace provides the path to where igo is installed
func (app *Application) Workspace() string {
	if *app.figs.Bool(kSystem) {
		return filepath.Join("/", "usr", "go")
	}
	return filepath.Join(app.userHomeDir, "go")
}

// runVersionCheck executes "go version" with specified environment variables and returns the output.
// Panics if an error occurs.
func (app *Application) runVersionCheck(envs map[string]string, version string) string {
	goBinPath := filepath.Join(app.Workspace(), "versions", version, "go", "bin", fmt.Sprintf("go.%s", version))

	if _, err := os.Stat(goBinPath); os.IsNotExist(err) {
		capture(fmt.Errorf("go binary does not exist at %s: %v", goBinPath, err))
	}

	cmdEnv := []string{
		fmt.Sprintf("GOROOT=%s", envs["GODIR"]),
		fmt.Sprintf("GOPATH=%s", envs["GODIR"]),
		fmt.Sprintf("GOBIN=%s", envs["GODIR"]),
		fmt.Sprintf("GOOS=%s", envs["GOOS"]),
		fmt.Sprintf("GOARCH=%s", envs["GOARCH"]),
	}

	cmd := exec.Command(goBinPath, "version")
	cmd.Env = append([]string{}, cmdEnv...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		capture(fmt.Errorf("failed to execute 'go version' with %s: %v\nOutput: %s", goBinPath, err, string(output)))
	}

	if *app.figs.Bool(kVerbose) {
		color.Green("Received terminal output: ")
		fmt.Println(string(output))
	}

	return strings.TrimSpace(string(output))
}

// installExtraPackages installs additional Go packages using the specified environment and version.
func (app *Application) installExtraPackages(envs map[string]string, version string) error {
	workspace := app.Workspace()
	goBinPath := filepath.Join(workspace, "versions", version, "go", "bin", fmt.Sprintf("go.%s", version))
	if _, err := os.Stat(goBinPath); os.IsNotExist(err) {
		return fmt.Errorf("go binary does not exist at %s: %w", goBinPath, err)
	}
	cmdEnv := []string{
		fmt.Sprintf("GOROOT=%s", filepath.Join(workspace, "versions", version, "go")),
		fmt.Sprintf("GOPATH=%s", filepath.Join(workspace, "versions", version)),
		fmt.Sprintf("GOBIN=%s", filepath.Join(workspace, "versions", version, "go", "bin")),
		fmt.Sprintf("GOOS=%s", envs["GOOS"]),
		fmt.Sprintf("GOARCH=%s", envs["GOARCH"]),
	}
	for pkg, modulePath := range packages {
		cmd := exec.Command(goBinPath, "install", fmt.Sprintf("%s@latest", modulePath))
		cmd.Env = append(os.Environ(), cmdEnv...) // Include existing env vars plus custom ones
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install %s (%s): %w\nOutput: %s", pkg, modulePath, err, string(output))
		}
		binPath := filepath.Join(envs["GOBIN"], pkg)
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			return fmt.Errorf("installation of %s succeeded but binary not found at %s", pkg, binPath)
		}
		color.Green("Installed %s successfully", pkg)
	}
	return nil
}

// patchShellConfigPath updates the shell config file to ensure PATH includes specific directories.
// envs is a map containing GOSHIMS, GOBIN, and GOSCRIPTS.
// Returns an error if the operation fails.
func (app *Application) patchShellConfigPath(envs map[string]string) error {
	requiredPaths := []string{
		envs["GOSHIMS"],
		envs["GOBIN"],
		envs["GOSCRIPTS"],
	}

	shellFiles := []string{
		filepath.Join(app.userHomeDir, ".bashrc"),
		filepath.Join(app.userHomeDir, ".zshrc"),
	}

	var targetFile string
	for _, shellFile := range shellFiles {
		if _, err := os.Stat(shellFile); err == nil {
			targetFile = shellFile
			break
		}
	}
	if targetFile == "" {
		return fmt.Errorf("no shell config file found (.bashrc or .zshrc)")
	}

	content, err := os.ReadFile(targetFile)
	capture(err)
	lines := strings.Split(string(content), "\n")

	// Look for the export PATH line
	var pathLine string
	pathLineIndex := -1
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "export PATH=") {
			pathLine = trimmedLine
			pathLineIndex = i
			break
		}
	}

	if pathLineIndex != -1 {
		pathValue := strings.TrimPrefix(pathLine, "export PATH=")
		pathParts := strings.Split(pathValue, ":")

		var missingPaths []string
		for _, reqPath := range requiredPaths {
			if reqPath == "" {
				return fmt.Errorf("required PATH component is empty in envs")
			}
			found := false
			for _, part := range pathParts {
				if strings.TrimSpace(part) == reqPath {
					found = true
					break
				}
			}
			if !found {
				missingPaths = append(missingPaths, reqPath)
			}
		}

		if len(missingPaths) > 0 {
			newPathValue := strings.Join(append(missingPaths, pathValue), ":")
			newPathLine := fmt.Sprintf("export PATH=%s", newPathValue)
			lines[pathLineIndex] = newPathLine
			capture(os.WriteFile(targetFile, []byte(strings.Join(lines, "\n")), 0644))
			fmt.Printf("Updated PATH in %s with missing paths: %v\n", targetFile, missingPaths)
		} else {
			fmt.Printf("PATH in %s already contains all required paths\n", targetFile)
		}
		return nil
	}

	newPathLine := fmt.Sprintf("export PATH=%s:%s:%s:$PATH", envs["GOSHIMS"], envs["GOBIN"], envs["GOSCRIPTS"])
	targetHandler := captureOpenFile(targetFile, os.O_APPEND|os.O_WRONLY, 0644)
	defer capture(targetHandler.Close())

	if len(content) > 0 && content[len(content)-1] != '\n' {
		captureInt(targetHandler.WriteString("\n"))
	}

	captureInt(targetHandler.WriteString(newPathLine + "\n"))

	return nil
}

// findGoVersions returns installed versions of Go in the igoWorkspace()
func (app *Application) findGoVersions() ([]string, error) {
	var versions []string
	d := app.Workspace()
	dvs := filepath.Join(d, "versions")
	capture(filepath.WalkDir(d, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			p := filepath.Base(path)
			for _, version := range versions {
				versionDir := filepath.Join(dvs, version)
				if strings.HasPrefix(path, versionDir) {
					return nil // skip over file in this version directory
				}
			}
			var _maj, _min, _pat int
			_, _ = fmt.Sscanf(p, "%d.%d.%d", &_maj, &_min, &_pat)
			if _maj == 0 && _min == 0 && _pat == 0 {
				return nil // silent skip over non-version directory
			}
			versions = append(versions, fmt.Sprintf("%d.%d.%d", _maj, _min, _pat))
		}
		return nil
	}))
	return versions, nil
}

// activatedVersion verifies which version is defined in the igoWorkspace()
func (app *Application) activatedVersion() (string, error) {
	d := app.Workspace()
	b, e := os.ReadFile(filepath.Join(d, "version"))
	if e != nil {
		return "", e
	}
	s := string(b)
	s = strings.TrimSpace(s)
	return s, nil
}

// injectEnvVarsToShellConfig will take the map of envs and add them to the bashrc or zshrc file as export ENV=val
func (app *Application) injectEnvVarsToShellConfig(envs map[string]string) error {
	// Possible shell config files to check
	shellFiles := []string{
		filepath.Join(app.userHomeDir, ".bashrc"),
		filepath.Join(app.userHomeDir, ".zshrc"),
	}

	// Find the first existing shell config file
	var targetFile string
	for _, file := range shellFiles {
		if _, err := os.Stat(file); err == nil {
			targetFile = file
			break
		}
	}
	if targetFile == "" {
		return fmt.Errorf("no shell config file found (.bashrc or .zshrc)")
	}

	// Read the existing content of the target file
	content, err := os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", targetFile, err)
	}
	existingLines := strings.Split(string(content), "\n")

	// Build a map of existing export statements
	existingExports := make(map[string]bool)
	for _, line := range existingLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "export ") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimPrefix(parts[0], "export ")
				existingExports[key] = true
			}
		}
	}

	// Prepare new export lines to append
	var newLines []string
	for key, value := range envs {
		exportLine := fmt.Sprintf("export %s=%s", key, value)
		if !existingExports[key] {
			newLines = append(newLines, exportLine)
		}
	}

	// If there are new lines to append, write them to the file
	if len(newLines) > 0 {
		// Open the file in append mode
		file, err := os.OpenFile(targetFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open %s for appending: %w", targetFile, err)
		}
		defer capture(file.Close())

		// Add a newline before appending if the file doesn't end with one
		if len(content) > 0 && content[len(content)-1] != '\n' {
			if _, err := file.WriteString("\n"); err != nil {
				return fmt.Errorf("failed to write newline to %s: %w", targetFile, err)
			}
		}

		// Write the new export lines
		for _, line := range newLines {
			if _, err := file.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("failed to write to %s: %w", targetFile, err)
			}
		}

		fmt.Printf("Updated %s with %d new environment variables\n", targetFile, len(newLines))
	} else {
		fmt.Printf("No new environment variables to add to %s\n", targetFile)
	}

	return nil
}
