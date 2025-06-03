package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/andreimerlescu/figtree/v2"
	"github.com/andreimerlescu/igo/internal"
	"github.com/fatih/color"
)

//go:embed bundled/shim.go.sh
var bundledShimsGoBytes embed.FS

//go:embed bundled/shim.gofmt.sh
var bundledShimsGofmtBytes embed.FS

type Application struct {
	ctx         context.Context
	Figs        figtree.Plant
	UserHomeDir string
	Workspace   func() string
}

var UserHomeDir = os.UserHomeDir

func NewApp() *Application {
	userHomeDir, err := UserHomeDir()
	internal.Capture(err)
	app := &Application{
		ctx:         context.Background(),
		UserHomeDir: userHomeDir,
	}
	app.Workspace = func() string {
		if *app.Figs.Bool(kSystem) {
			return filepath.Join("/", "usr", "go")
		}
		return filepath.Join(app.UserHomeDir, "go")
	}
	app.Figs = figtree.With(figtree.Options{
		ConfigFile: filepath.Join(app.UserHomeDir, ".igo.config.yml"),
		Germinate:  true,
		Harvest:    0,
	})
	// 5 string 3 bool
	app.Figs.NewBool(cmdHelp, false, "Display help")
	app.Figs.NewBool(cmdVersion, false, "Display version")
	app.Figs.NewBool(cmdList, false, "Display installed versions")
	app.Figs.NewBool(cmdEnv, false, "Display env")
	app.Figs.NewString(cmdInstall, "", "Install a specific version of Go")
	app.Figs.NewString(cmdUninstall, "", "Uninstall a specific version of Go")
	app.Figs.NewString(cmdActivate, "", "Activate a specific version of Go")
	app.Figs.NewString(cmdFix, "", "Fix a specific version of Go")
	app.Figs.NewString(cmdSwitch, "", "Switch to a specific version of Go")
	app.Figs.NewBool(kSystem, false, "Install in system mode /usr/bin/go")
	app.Figs.NewBool(kDebug, false, "Enable debug mode")
	app.Figs.NewBool(kVerbose, false, "Enable verbose mode")
	app.Figs.NewString(kGoDir, filepath.Join(app.UserHomeDir, "go"), "Path where you want multiple go versions installed")
	app.Figs.NewString(kGoos, runtime.GOOS, "Go OS")
	app.Figs.NewString(kGoArch, runtime.GOARCH, "Go Architecture")
	app.Figs.NewBool(kExtras, true, "Install extra packages")
	app.Figs.NewMap(kExtraPackages, packages, "Extra packages to install")
	_, err = os.Lstat(figtree.ConfigFilePath)
	if os.IsNotExist(err) || os.IsPermission(err) {
		internal.Capture(app.Figs.Parse())
	} else {
		internal.Capture(app.Figs.Load())
	}
	return app
}

// Add this function to application.go to validate Go version formats
func (app *Application) validateVersion(version string) error {
	// Basic format check with regex
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(version) {
		return fmt.Errorf("invalid go version format: %s (expected format: X.Y.Z)", version)
	}

	// Parse version parts to check they're valid numbers
	var major, minor, patch int
	_, err := fmt.Sscanf(version, VersionFmt, &major, &minor, &patch)
	if err != nil {
		return fmt.Errorf("error parsing version components: %w", err)
	}

	// Optional: Add constraints on minimum supported versions
	if major < 1 || (major == 1 && minor < 16) {
		return fmt.Errorf("go version %s is not supported (minimum: 1.16.0)", version)
	}

	// Verify the version exists on Go's download server before proceeding
	// Using a HEAD request to check if the URL exists
	resp, err := httpHead(fmt.Sprintf("https://go.dev/dl/go%s.%s-%s.tar.gz",
		version, *app.Figs.String(kGoos), *app.Figs.String(kGoArch)))
	if err != nil {
		return fmt.Errorf("error checking if version exists: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("go version %s not found on download server (status: %d)",
			version, resp.StatusCode)
	}

	return nil
}

// CreateShims creates the shims for go and gofmt
func (app *Application) CreateShims() error {
	workspace := app.Workspace()
	shimsDir := filepath.Join(workspace, "shims")
	goShim := filepath.Join(shimsDir, "go")
	gofmtShim := filepath.Join(shimsDir, "gofmt")
	shimGoBytes, err := bundledShimsGoBytes.ReadFile("bundled/shim.go.sh")
	if err != nil {
		return fmt.Errorf("failed to read bundled shim.go.sh: %v", err)
	}
	err = os.WriteFile(goShim, shimGoBytes, 0755)
	if err != nil {
		return fmt.Errorf("failed to write shim.go.sh: %v", err)
	}
	shimGofmtBytes, err := bundledShimsGofmtBytes.ReadFile("bundled/shim.gofmt.sh")
	if err != nil {
		return fmt.Errorf("failed to read bundled shim.go.sh: %v", err)
	}
	err = os.WriteFile(gofmtShim, shimGofmtBytes, 0755)
	if err != nil {
		return fmt.Errorf("failed to write shim.gofmt.sh: %v", err)
	}
	internal.Capture(os.Chmod(goShim, 0755))
	internal.Capture(os.Chmod(gofmtShim, 0755))
	return nil
}

// runVersionCheck executes "go version" with specified environment variables and returns the output.
// Panics if an error occurs.
func (app *Application) runVersionCheck(envs map[string]string, version string) string {
	goBinPath := filepath.Join(app.Workspace(), "versions", version, "go", "bin", fmt.Sprintf("go.%s", version))

	if _, err := os.Stat(goBinPath); os.IsNotExist(err) {
		internal.Capture(fmt.Errorf("go binary does not exist at %s: %v", goBinPath, err))
	}

	cmdEnv := []string{
		fmt.Sprintf("GOROOT=%s", envs[GOROOT]),
		fmt.Sprintf("GOPATH=%s", envs[GOPATH]),
		fmt.Sprintf("GOBIN=%s", envs[GOBIN]),
		fmt.Sprintf("GOOS=%s", envs[GOOS]),
		fmt.Sprintf("GOARCH=%s", envs[GOARCH]),
		fmt.Sprintf("GOMODCACHE=%s", envs[GOMODCACHE]),
	}

	cmd := exec.Command(goBinPath, "version")
	cmd.Env = append([]string{}, cmdEnv...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		internal.Capture(fmt.Errorf("failed to execute 'go version' with %s: %v\nOutput: %s", goBinPath, err, string(output)))
	}
	gover := strings.TrimSpace(string(output))
	if *app.Figs.Bool(kVerbose) {
		color.Green("Received terminal output: %s", gover)
	}

	return gover
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
		fmt.Sprintf("GOOS=%s", envs[GOOS]),
		fmt.Sprintf("GOARCH=%s", envs[GOARCH]),
	}
	p := app.Figs.Fig(kExtraPackages).ToString()
	color.Green("Installing extra packages: %s", p)

	for pkg, modulePath := range packages {
		cmd := exec.Command(goBinPath, "install", fmt.Sprintf("%s@latest", modulePath))
		cmd.Env = append(os.Environ(), cmdEnv...) // Include existing env vars plus custom ones
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install %s (%s): %w\nOutput: %s", pkg, modulePath, err, string(output))
		}
		binPath := filepath.Join(envs[GOBIN], pkg)
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
		envs[GOSHIMS],
		envs[GOBIN],
		envs[GOSCRIPTS],
	}

	bashrc := filepath.Join(app.UserHomeDir, ".profile")
	zshrc := filepath.Join(app.UserHomeDir, ".zshrc.local")
	shellFiles := []string{bashrc, zshrc}

	var targetFile string
	for _, shellFile := range shellFiles {
		if _, err := os.Stat(shellFile); !os.IsNotExist(err) && !os.IsPermission(err) {
			color.Green("Found %s", shellFile)
			targetFile = shellFile
			break
		}
	}
	if targetFile == "" {
		contents := fmt.Sprintf("export PATH=%s:%s:%s:%s\n",
			envs[GOSHIMS], envs[GOSCRIPTS], envs[GOBIN], os.Getenv("PATH"))
		internal.Capture(os.WriteFile(zshrc, []byte(contents), 0644))
		return os.WriteFile(bashrc, []byte(contents), 0644)
	}

	content, err := os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("168 failed to read file %s: %w", targetFile, err)
	}
	if *app.Figs.Bool(kVerbose) {
		color.Green("Contents of %s is: \n%s\n", targetFile, content)
	}
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
				return fmt.Errorf("194 required PATH component is empty in envs")
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
			err := os.WriteFile(targetFile, []byte(strings.Join(lines, "\n")), 0644)
			if err != nil {
				return fmt.Errorf("214 failed to write file %s: %w", targetFile, err)
			}
			fmt.Printf("Updated PATH in %s with missing paths: %v\n", targetFile, missingPaths)
		} else {
			fmt.Printf("PATH in %s already contains all required paths\n", targetFile)
		}
		return nil
	}

	newPathLine := fmt.Sprintf("export PATH=%s:%s:%s:%s", envs[GOSHIMS], envs[GOBIN], envs[GOSCRIPTS], os.Getenv("PATH"))
	targetHandler, err := os.OpenFile(targetFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("226 could not open target file: %w", err)
	}

	if len(content) > 0 && content[len(content)-1] != '\n' {
		_, err = targetHandler.WriteString("\n")
		if err != nil {
			return fmt.Errorf("232 could not write to target file: %w", err)
		}
	}

	_, err = targetHandler.WriteString(newPathLine + "\n")
	if err != nil {
		return fmt.Errorf("238 could not write to target file: %w", err)
	}

	return targetHandler.Close()
}

// findGoVersions returns installed versions of Go in the igoWorkspace()
func (app *Application) findGoVersions() ([]string, error) {
	var versions []string
	d := app.Workspace()
	dvs := filepath.Join(d, "versions")
	internal.Capture(filepath.WalkDir(d, func(path string, d fs.DirEntry, err error) error {
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
			var vMaj, vMin, vPat int
			_, _ = fmt.Sscanf(p, VersionFmt, &vMaj, &vMin, &vPat)
			if vMaj == 0 && vMin == 0 && vPat == 0 {
				return nil // silent skip over non-version directory
			}
			versions = append(versions, fmt.Sprintf(VersionFmt, vMaj, vMin, vPat))
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
		filepath.Join(app.UserHomeDir, ".profile"),
		filepath.Join(app.UserHomeDir, ".bash_profile"),
		filepath.Join(app.UserHomeDir, ".zshrc.local"),
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
		targetFile = filepath.Join(app.UserHomeDir, ".profile")
		err := os.WriteFile(targetFile, []byte(""), 0644)
		if err != nil {
			return fmt.Errorf("305 failed to write %s: %w", targetFile, err)
		}
	}

	// Read the existing content of the target file
	content, err := os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("312 failed to read %s: %w", targetFile, err)
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
		shellProfileFile, err := os.OpenFile(targetFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("343 failed to open %s for appending: %w", targetFile, err)
		}

		// Add a newline before appending if the file doesn't end with one
		if len(content) > 0 && content[len(content)-1] != '\n' {
			if _, err := shellProfileFile.WriteString("\n"); err != nil {
				return fmt.Errorf("349 failed to write newline to %s: %w", targetFile, err)
			}
		}

		// Write the new export lines
		for _, line := range newLines {
			if _, err := shellProfileFile.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("356 failed to write to %s: %w", targetFile, err)
			}
		}

		fmt.Printf("Updated %s with %d new environment variables\n", targetFile, len(newLines))

		return shellProfileFile.Close()
	} else {
		fmt.Printf("No new environment variables to add to %s\n", targetFile)
	}

	return nil
}
