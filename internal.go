package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// about prints the product information
func about() string {
	sb := strings.Builder{}
	sb.WriteString(PRODUCT + " ")
	sb.WriteString("[open source at " + AUTHOR + "]")
	return sb.String()
}

// captureInt will discard the integer and process the error using capture()
func captureInt(_ int, err error) {
	if err != nil {
		panic(err)
	}
	return
}

// capture accepts nil or an error or multiple errors
//
// Example:
//
//			capture(errors.New("this is an error to run os.Exit(1) after printing this =D"))
//			// OR
//	     var E1 error
//	     var E2 error
//			capture(E1, E2)
func capture(err ...error) {
	if err == nil || len(err) == 0 || err[0] == nil {
		return
	}
	fmt.Println(err)
	os.Exit(1)
}

// discard err will print the error only if it occurred
func discard(err ...error) {
	if err == nil || len(err) == 0 || err[0] == nil {
		return
	}
	fmt.Println(err)
}

// captureOpenFile is a helper func that accepts a path, opens it or capture() the error
//
// Example:
//
//	handler := captureOpenFile("/opt/app/config.yaml", os.O_RDONLY, 0600)
func captureOpenFile(path string, flag int, perm os.FileMode) *os.File {
	f, e := os.OpenFile(path, flag, perm)
	capture(e)
	return f
}

// removeSymlinkOrBackupPath checks if the given path is a symlink and deletes it if it is not.
// Returns an error if the check or deletion fails.
func removeSymlinkOrBackupPath(path string) error {
	if !PathExists(path) {
		return nil
	}
	if isSymlink(path) {
		err := os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to remove symlink %s: %w", path, err)
		}
		return nil
	}

	err := os.Rename(path, path+".bak") // path isn't symlink, so move it to .bak
	if err != nil {
		return fmt.Errorf("failed to delete non-symlink path %s: %w", path, err)
	}

	return nil
}

func makeDirsWritable(path string) error {
	return filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		mode := info.Mode()
		if mode&0200 == 0 { // Owner write permission is not set
			newMode := mode | 0200
			if err := os.Chmod(p, newMode); err != nil {
				return fmt.Errorf("failed to chmod u+w on %s: %w", p, err)
			}
		}
		return nil
	})
}

func setStickyBit(path string) error {
	if !IsDirectory(path) {
		return fmt.Errorf("setting sticky bits from files do nothing: %s", path)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", fmt.Sprintf("chmod +t %s", path))
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set sticky bit: %w", err)
	}
	return nil
}

func setSetuidSetgidBits(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", fmt.Sprintf("chmod -R a+s %s", path))
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set setuid/setgid bits: %w", err)
	}
	return nil
}

func removeStickyBit(path string) error {
	if !IsDirectory(path) {
		return fmt.Errorf("remove sticky bits from files do nothing: %s", path)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", fmt.Sprintf("chmod -t %s", path))
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove sticky bit: %w", err)
	}
	return nil
}

func removeSetuidSetgidBits(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", fmt.Sprintf("chmod -R a-s %s", path))
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove setuid/setgid bits: %w", err)
	}
	return nil
}

// FindSymlinks scans the provided directory path and returns a slice
// of full paths to any symlinks found in the root directory only.
// It does not recurse into subdirectories.
func FindSymlinks(dirPath string) ([]string, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %w", err)
	}
	defer dir.Close()
	entries, err := dir.ReadDir(-1) // -1 to read all entries
	if err != nil {
		return nil, fmt.Errorf("failed to read directory entries: %w", err)
	}
	var symlinks []string
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			fullPath := filepath.Join(dirPath, entry.Name())
			symlinks = append(symlinks, fullPath)
		}
	}
	return symlinks, nil
}

// ReadSymlink reads the target of a symlink
func ReadSymlink(path string) (string, error) {
	target, err := os.Readlink(path)
	if err != nil {
		return "", fmt.Errorf("failed to read symlink %s: %w", path, err)
	}

	// If the symlink target is relative, make it absolute
	if !filepath.IsAbs(target) {
		symlinkDir := filepath.Dir(path)
		target = filepath.Join(symlinkDir, target)
	}

	return target, nil
}

// VerifyLink checks if a symlink correctly resolves to its expected target
// and returns the appropriate status emoji (✅ for success, ❌ for failure)
func VerifyLink(link, expectedTarget string) string {
	actualTarget, err := ReadSymlink(link)
	if err != nil {
		return "❌" // Error reading the symlink
	}

	// Clean both paths to ensure consistent comparison
	// This normalizes paths by removing redundant separators, dots, etc.
	cleanExpected := filepath.Clean(expectedTarget)
	cleanActual := filepath.Clean(actualTarget)

	// Compare the cleaned paths
	if cleanActual == cleanExpected {
		return "✅" // Link correctly points to expected target
	}

	return "❌" // Link points to a different target
}

func IsDirectory(path string) bool {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func PathExists(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err) && !os.IsPermission(err)
}

func isSymlink(path string) bool {
	fileInfo, err := os.Lstat(path) // Lstat to not follow symlinks
	if err != nil {
		return false
	}
	return fileInfo.Mode()&os.ModeSymlink != 0
}

// touch creates a new empty file or updates the modification time of an existing file at the given path.
// Returns an error if the operation fails.
func touch(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		pathFile, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}
		if err := pathFile.Close(); err != nil {
			return fmt.Errorf("failed to close file %s: %w", path, err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	currentTime := time.Now()
	if err := os.Chtimes(path, currentTime, currentTime); err != nil {
		return fmt.Errorf("failed to update modification time of %s: %w", path, err)
	}

	return nil
}
