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
	p := app.Figs.FigFlesh(kExtraPackages).ToString()
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
// patchShellConfigPath is now mostly informational / first-time setup helper
func (app *Application) patchShellConfigPath(envs map[string]string) error {
	verbose, _ := *app.Figs.Bool(kVerbose), *app.Figs.Bool(kDebug)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	activatorPath := filepath.Join(home, ".igo")

	if _, err := os.Stat(activatorPath); os.IsNotExist(err) {
		if err := app.installActivatorScript(activatorPath); err != nil {
			return err
		}

		app.printActivatorInstruction(activatorPath)
		return nil
	}

	if verbose {
		color.Green("Activator script already present at %s", activatorPath)
	}
	return nil
}

// installActivatorScript copies the embedded template to the target location
func (app *Application) installActivatorScript(dst string) error {
	content, err := activatorTemplate.ReadFile("bundled/.igosh.tpl")
	if err != nil {
		return fmt.Errorf("failed to read embedded .igosh.tpl: %w", err)
	}

	finalContent := string(content)
	finalContent = strings.ReplaceAll(finalContent, "{{ .Dir }}", app.Workspace())

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(dst, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write activator script: %w", err)
	}

	color.Green("Created igo activator script → %s", dst)
	return nil
}

// printActivatorInstruction shows clear, copy-paste friendly message
func (app *Application) printActivatorInstruction(activatorPath string) {
	shell := detectUserShell()

	msg := fmt.Sprintf(`
To finish setting up igo, add the following line to your shell configuration file:

    source %q

Depending on your shell, common locations are:

    • zsh:   ~/.zshrc
    • bash:  ~/.bashrc  or  ~/.bash_profile
    • other: ~/.profile

After adding the line, either restart your terminal or run:

    source ~/.zshrc    # (or appropriate file)

`, activatorPath)

	color.Cyan(msg)

	if shell != "" {
		color.Yellow("Detected shell: %s\n", shell)
	}
}

func detectUserShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shellNameFromPath(shell)
	}

	u := internal.User()

	if u.Shell != "" && u.Shell != "/bin/false" && !strings.HasSuffix(u.Shell, "nologin") {
		return shellNameFromPath(u.Shell)
	}

	return fallbackByOS()
}

func fallbackByOS() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS default since ~2019
		return "zsh"
	default:
		// linux, freebsd, openbsd, etc.
		// no strong default — better not to guess
		return ""
	}
}

func shellNameFromPath(path string) string {
	name := filepath.Base(path)
	name = strings.ToLower(name)
	name = strings.TrimSuffix(name, ".sh")
	name = strings.TrimRight(name, "0123456789.-")

	switch name {
	case "bash", "zsh", "fish", "ksh", "csh", "tcsh", "sh":
		return name
	default:
		return ""
	}
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
