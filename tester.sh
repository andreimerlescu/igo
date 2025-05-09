#!/bin/bash
set -e
echo "Starting igo test script..."

# Initial setup
export PATH=/bin:$PATH

# Display igo version
echo "=== IGO VERSION ==="
igo -version || exit 1
echo

# List installed versions (should be empty)
echo "=== INITIAL LIST (Should be empty) ==="
igo -cmd list || exit 1
echo
exit 0
# Install Go 1.24.2
echo "=== INSTALLING GO 1.24.2 ==="
igo -cmd install -gover 1.24.2
echo

# Verify Go version is 1.24.2
echo "=== VERIFYING GO 1.24.2 ==="
source ~/.bashrc || source ~/.zshrc || echo "Failed to source shell config"
go version | grep "go1.24.2" && echo "Go 1.24.2 verified!" || echo "FAIL: Go 1.24.2 not active"
echo

# Install Go 1.24.3
echo "=== INSTALLING GO 1.24.3 ==="
igo -cmd install -gover 1.24.3
echo

# List installed versions (should see both with 1.24.2 activated)
echo "=== LISTING GO VERSIONS (Should show 1.24.2 and 1.24.3 with 1.24.2 active) ==="
igo -cmd list
echo

# Switch to Go 1.24.3
echo "=== SWITCHING TO GO 1.24.3 ==="
igo -cmd use -gover 1.24.3
echo

# Verify Go version is 1.24.3
echo "=== VERIFYING GO 1.24.3 ==="
source ~/.bashrc || source ~/.zshrc || echo "Failed to source shell config"
go version | grep "go1.24.3" && echo "Go 1.24.3 verified!" || echo "FAIL: Go 1.24.3 not active"
echo

# Install summarize
echo "=== INSTALLING SUMMARIZE ==="
go install github.com/andreimerlescu/summarize@latest
echo

# Run summarize on GODIR
echo "=== RUNNING SUMMARIZE ON GODIR ==="
summarize ~/go
echo

# Check the summary files
echo "=== CHECKING SUMMARY FILES ==="
find ~/go/summaries -name "*.md" -exec cat {} \;
echo

# List directory structure
echo "=== LISTING DIRECTORY STRUCTURE ==="
find ~/go -type d | sort
echo

# Remove Go 1.24.2
echo "=== REMOVING GO 1.24.2 ==="
igo -cmd uninstall -gover 1.24.2
echo

# List installed versions
echo "=== LISTING GO VERSIONS (After removing 1.24.2) ==="
igo -cmd list
echo

# Remove Go 1.24.3
echo "=== REMOVING GO 1.24.3 ==="
igo -cmd uninstall -gover 1.24.3
echo

# List installed versions (should be empty)
echo "=== LISTING GO VERSIONS (Should be empty) ==="
igo -cmd list
echo

# Verify summarize command fails
echo "=== VERIFYING SUMMARIZE COMMAND FAILS ==="
if summarize ~/go 2>/dev/null; then
  echo "FAIL: Summarize still working after Go removal"
else
  echo "SUCCESS: Summarize command failed as expected"
fi
echo

echo "All tests completed!"
