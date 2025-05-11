package main

import (
	"fmt"
	"os"
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
