package internal

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// PRODUCT is called igo for Install Go
	PRODUCT string = "igo"

	// AUTHOR is Andrei
	AUTHOR string = "github.com/ProjectApario/igo"

	// XRP is how you can tip the AUTHOR
	XRP string = "rAparioji3FxAtD7UufS8Hh9XmFn7h6AX"
)

var UserCurrent = user.Current

// About prints the product information
func About() string {
	sb := strings.Builder{}
	sb.WriteString(PRODUCT + " ")
	sb.WriteString("open source at " + AUTHOR)
	return sb.String()
}

// CaptureInt will Discard the integer and process the error using Capture()
var CaptureInt = func(_ int, err error) {
	if err != nil {
		panic(err)
	}
	return
}

// Capture accepts nil or an error or multiple errors
//
// Example:
//
//			Capture(errors.New("this is an error to run os.Exit(1) after printing this =D"))
//			// OR
//	     var E1 error
//	     var E2 error
//			Capture(E1, E2)
var Capture = func(err ...error) {
	if err == nil || len(err) == 0 || err[0] == nil {
		return
	}
	fmt.Println(err)
	os.Exit(1)
}

// Discard err will print the error only if it occurred
var Discard = func(err ...error) {
	if err == nil || len(err) == 0 || err[0] == nil {
		return
	}
	fmt.Println(err)
}

// CaptureOpenFile is a helper func that accepts a path, opens it or Capture() the error
//
// Example:
//
//	handler := CaptureOpenFile("/opt/app/config.yaml", os.O_RDONLY, 0600)
var CaptureOpenFile = func(path string, flag int, perm os.FileMode) *os.File {
	f, e := os.OpenFile(path, flag, perm)
	Capture(e)
	return f
}

// RemoveSymlinkOrBackupPath checks if the given path is a symlink and deletes it if it is not.
// Returns an error if the check or deletion fails.
var RemoveSymlinkOrBackupPath = func(path string) error {
	if !PathExists(path) {
		return nil
	}
	if IsSymlink(path) {
		err := os.Remove(path)
		if err != nil {
			return ErrPathFailed{path, err, ""}
		}
		return nil
	}

	err := os.Rename(path, path+".bak") // path isn't symlink, so move it to .bak
	if err != nil {
		return ErrPathFailed{path, err, "non-"}
	}

	return nil
}

var MakeDirsWritable = func(path string) error {
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
				return ErrChmodFailed{p, err}
			}
		}
		return nil
	})
}

var SetStickyBit = func(path string) error {
	if !IsDirectory(path) {
		return ErrStickyBitsOnFile{path}
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", fmt.Sprintf("chmod +t %s", path))
	default:
		return ErrOSNotSupported{runtime.GOOS}
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return ErrStickyBitFailed{path, err, "set"}
	}
	return nil
}

var SetSetuidSetgidBits = func(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", fmt.Sprintf("chmod -R a+s %s", path))
	default:
		return ErrOSNotSupported{runtime.GOOS}
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return ErrBitNotSet{err}
	}
	return nil
}

var RemoveStickyBit = func(path string) error {
	if !IsDirectory(path) {
		return ErrStickyBitsOnFile{path}
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", fmt.Sprintf("chmod -t %s", path))
	default:
		return ErrOSNotSupported{runtime.GOOS}
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return ErrStickyBitFailed{path, err, "remove"}
	}
	return nil
}

var RemoveSetuidSetgidBits = func(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", fmt.Sprintf("chmod -R a-s %s", path))
	default:
		return ErrOSNotSupported{runtime.GOOS}
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return ErrSetUIDGIDBit{path, err, "remove"}
	}
	return nil
}

// FindSymlinks scans the provided directory path and returns a slice
// of full paths to any symlinks found in the root directory only.
// It does not recurse into subdirectories.
var FindSymlinks = func(dirPath string) ([]string, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, ErrFile{dirPath, err, "os.Open"}
	}
	defer dir.Close()
	entries, err := dir.ReadDir(-1) // -1 to read all entries
	if err != nil {
		return nil, ErrDirEntries{dirPath, err}
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
var ReadSymlink = func(path string) (string, error) {
	target, err := os.Readlink(path)
	if err != nil {
		return "", ErrPathFailed{path, err, "remove"}
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
var VerifyLink = func(link, expectedTarget string) string {
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

var IsDirectory = func(path string) bool {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

var PathExists = func(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err) && !os.IsPermission(err)
}

var IsSymlink = func(path string) bool {
	fileInfo, err := os.Lstat(path) // Lstat to not follow symlinks
	if err != nil {
		return false
	}
	return fileInfo.Mode()&os.ModeSymlink != 0
}

// Touch creates a new empty file or updates the modification time of an existing file at the given path.
// Returns an error if the operation fails.
var Touch = func(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		pathFile, err := os.Create(path)
		if err != nil {
			return ErrFile{path, err, "os.Create"}
		}
		if err := pathFile.Close(); err != nil {
			return ErrFile{path, err, "path.Close"}
		}
		return nil
	} else if err != nil {
		return ErrFile{path, err, "os.Stat"}
	}

	currentTime := time.Now()
	if err := os.Chtimes(path, currentTime, currentTime); err != nil {
		return ErrFile{path, err, "os.Chtimes"}
	}

	return nil
}

var CheckRootPrivileges = func() bool {
	if runtime.GOOS == "windows" {
		return false // use msi installer for windows
	} else {
		if os.Geteuid() == 0 {
			return true
		}
		if User().Uid == "0" {
			return true
		}

		return false
	}
}

var User = func() *user.User {
	currentUser, err := UserCurrent()
	if err != nil {
		return &user.User{
			Uid:      "-1",
			Gid:      "-1",
			Username: "nobody",
			Name:     "nobody",
			HomeDir:  "/dev/null",
		}
	}
	return currentUser
}
