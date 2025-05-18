package main

import (
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

func TestAbout(t *testing.T) {
	result := about()
	if !strings.Contains(result, PRODUCT) {
		t.Errorf("about() should contain the product name, got: %s", result)
	}
	if !strings.Contains(result, AUTHOR) {
		t.Errorf("about() should contain the author info, got: %s", result)
	}
}

func TestCaptureIntWithNilError(t *testing.T) {
	captureInt(42, nil)
}

func TestCaptureIntWithError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("captureInt should have panicked with an error")
		}
	}()
	captureInt(42, errors.New("test error"))
}

var osExit = func(code int) {
	os.Exit(code)
}

func TestDiscard(t *testing.T) {
	discard(nil)

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	discard(errors.New("test error"))

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = originalStdout

	if !strings.Contains(string(out), "test error") {
		t.Errorf("discard() should print error message")
	}
}

func TestCaptureOpenFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-capture-open-file-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	originalCapture := capture
	defer func() { capture = originalCapture }()
	captureWasCalled := false
	capture = func(errs ...error) {
		if errs != nil && len(errs) > 0 && errs[0] != nil {
			captureWasCalled = true
			panic(errs[0])
		}
	}

	file := captureOpenFile(tempPath, os.O_RDONLY, 0600)
	if file == nil {
		t.Errorf("captureOpenFile should return a file handle")
	} else {
		file.Close()
	}
	if captureWasCalled {
		t.Errorf("capture should not be called with valid file")
	}

	captureWasCalled = false
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("captureOpenFile should panic with non-existent file")
		}
	}()
	captureOpenFile("/nonexistent/file", os.O_RDONLY, 0600)
}

func TestRemoveSymlinkOrBackupPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-remove-symlink-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	nonExistentPath := filepath.Join(tempDir, "nonexistent")
	err = removeSymlinkOrBackupPath(nonExistentPath)
	if err != nil {
		t.Errorf("removeSymlinkOrBackupPath should not return error for non-existent path: %v", err)
	}

	linkPath := filepath.Join(tempDir, "symlink")
	targetPath := filepath.Join(tempDir, "target")
	err = os.WriteFile(targetPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	err = os.Symlink(targetPath, linkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	err = removeSymlinkOrBackupPath(linkPath)
	if err != nil {
		t.Errorf("Failed to remove symlink: %v", err)
	}
	if PathExists(linkPath) {
		t.Errorf("Symlink should be removed")
	}

	filePath := filepath.Join(tempDir, "regular-file")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create regular file: %v", err)
	}

	err = removeSymlinkOrBackupPath(filePath)
	if err != nil {
		t.Errorf("Failed to backup regular file: %v", err)
	}
	if PathExists(filePath) {
		t.Errorf("Regular file should be renamed")
	}
	if !PathExists(filePath + ".bak") {
		t.Errorf("Backup file should exist")
	}
}

func TestMakeDirsWritable(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-make-dirs-writable-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0555)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	err = makeDirsWritable(tempDir)
	if err != nil {
		t.Errorf("makeDirsWritable failed: %v", err)
	}

	info, err := os.Stat(subDir)
	if err != nil {
		t.Fatalf("Failed to stat subdirectory: %v", err)
	}
	if info.Mode()&0200 == 0 {
		t.Errorf("Subdirectory should be writable, but it's not")
	}
}

func TestIsDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-is-directory-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile, err := os.CreateTemp(tempDir, "test-file-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFilePath := tempFile.Name()
	tempFile.Close()

	if !IsDirectory(tempDir) {
		t.Errorf("IsDirectory should return true for directory: %s", tempDir)
	}

	if IsDirectory(tempFilePath) {
		t.Errorf("IsDirectory should return false for file: %s", tempFilePath)
	}

	if IsDirectory(filepath.Join(tempDir, "nonexistent")) {
		t.Errorf("IsDirectory should return false for non-existent path")
	}
}

func TestPathExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-path-exists-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile, err := os.CreateTemp(tempDir, "test-file-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFilePath := tempFile.Name()
	tempFile.Close()

	if !PathExists(tempDir) {
		t.Errorf("PathExists should return true for existing directory: %s", tempDir)
	}

	if !PathExists(tempFilePath) {
		t.Errorf("PathExists should return true for existing file: %s", tempFilePath)
	}

	if PathExists(filepath.Join(tempDir, "nonexistent")) {
		t.Errorf("PathExists should return false for non-existent path")
	}
}

func TestIsSymlink(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-is-symlink-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "regular-file")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	linkPath := filepath.Join(tempDir, "symlink")
	err = os.Symlink(filePath, linkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	if !isSymlink(linkPath) {
		t.Errorf("isSymlink should return true for symlink: %s", linkPath)
	}

	if isSymlink(filePath) {
		t.Errorf("isSymlink should return false for regular file: %s", filePath)
	}

	if isSymlink(filepath.Join(tempDir, "nonexistent")) {
		t.Errorf("isSymlink should return false for non-existent path")
	}
}

func TestFindSymlinks(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-find-symlinks-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	file1Path := filepath.Join(tempDir, "file1")
	file2Path := filepath.Join(tempDir, "file2")
	err = os.WriteFile(file1Path, []byte("test1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	err = os.WriteFile(file2Path, []byte("test2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	link1Path := filepath.Join(tempDir, "link1")
	link2Path := filepath.Join(tempDir, "link2")
	err = os.Symlink(file1Path, link1Path)
	if err != nil {
		t.Fatalf("Failed to create link1: %v", err)
	}
	err = os.Symlink(file2Path, link2Path)
	if err != nil {
		t.Fatalf("Failed to create link2: %v", err)
	}

	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	subLinkPath := filepath.Join(subDir, "sublink")
	err = os.Symlink(file1Path, subLinkPath)
	if err != nil {
		t.Fatalf("Failed to create sublink: %v", err)
	}

	symlinks, err := FindSymlinks(tempDir)
	if err != nil {
		t.Errorf("FindSymlinks failed: %v", err)
	}

	if len(symlinks) != 2 {
		t.Errorf("FindSymlinks should find 2 symlinks, found %d", len(symlinks))
	}

	foundLink1 := false
	foundLink2 := false
	for _, link := range symlinks {
		if link == link1Path {
			foundLink1 = true
		} else if link == link2Path {
			foundLink2 = true
		}
	}

	if !foundLink1 {
		t.Errorf("FindSymlinks should find link1 at %s", link1Path)
	}
	if !foundLink2 {
		t.Errorf("FindSymlinks should find link2 at %s", link2Path)
	}
}

func TestReadSymlink(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-read-symlink-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "target-file")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	relLinkPath := filepath.Join(tempDir, "rel-link")
	err = os.Symlink("target-file", relLinkPath)
	if err != nil {
		t.Fatalf("Failed to create relative symlink: %v", err)
	}

	absLinkPath := filepath.Join(tempDir, "abs-link")
	err = os.Symlink(filePath, absLinkPath)
	if err != nil {
		t.Fatalf("Failed to create absolute symlink: %v", err)
	}

	target, err := ReadSymlink(relLinkPath)
	if err != nil {
		t.Errorf("ReadSymlink failed with relative link: %v", err)
	}
	if target != filePath {
		t.Errorf("ReadSymlink with relative link should return absolute path to target: expected=%s, got=%s", filePath, target)
	}

	target, err = ReadSymlink(absLinkPath)
	if err != nil {
		t.Errorf("ReadSymlink failed with absolute link: %v", err)
	}
	if target != filePath {
		t.Errorf("ReadSymlink with absolute link should return absolute path to target: expected=%s, got=%s", filePath, target)
	}

	_, err = ReadSymlink(filePath)
	if err == nil {
		t.Errorf("ReadSymlink should fail with non-symlink path")
	}
}

func TestVerifyLink(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-verify-link-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	file1Path := filepath.Join(tempDir, "file1")
	file2Path := filepath.Join(tempDir, "file2")
	err = os.WriteFile(file1Path, []byte("test1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	err = os.WriteFile(file2Path, []byte("test2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	linkPath := filepath.Join(tempDir, "link")
	err = os.Symlink(file1Path, linkPath)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	result := VerifyLink(linkPath, file1Path)
	if result != "✅" {
		t.Errorf("VerifyLink should return success emoji for correct target, got %s", result)
	}

	result = VerifyLink(linkPath, file2Path)
	if result != "❌" {
		t.Errorf("VerifyLink should return failure emoji for incorrect target, got %s", result)
	}

	result = VerifyLink(filepath.Join(tempDir, "nonexistent"), file1Path)
	if result != "❌" {
		t.Errorf("VerifyLink should return failure emoji for non-existent symlink, got %s", result)
	}
}

func TestUser(t *testing.T) {
	originalUserCurrent := userCurrent
	defer func() { userCurrent = originalUserCurrent }()

	currentUser, _ := user.Current()
	u := User()
	if u.Uid != currentUser.Uid || u.Username != currentUser.Username {
		t.Errorf("User() should return current user info")
	}

	userCurrent = func() (*user.User, error) {
		return nil, errors.New("user.Current error")
	}
	u = User()
	if u.Uid != "-1" || u.Username != "nobody" {
		t.Errorf("User() should return fallback user with Uid=-1, Username=nobody when user.Current fails")
	}
}
