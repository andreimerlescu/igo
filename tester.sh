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

# Install Go 1.24.2
echo "=== INSTALLING GO 1.24.2 ==="
igo -cmd install -gover 1.24.2 --debug || exit 1
echo

echo "=== LISTING FILES ==="
tree -L 3 || exit 1
echo

echo "=== ENVIRONMENT VARIABLES ==="
env | sort || exit 1
echo

USERNAME=$(whoami | tr -d '\n')

source ~/.profile || { echo "Failed to source $USERNAME shell config"; exit 1; }
source ~/.bashrc || { echo "Failed to source $USERNAME shell config"; exit 1; }


echo "=== BASH PROFILE ==="
cat ~/.profile || exit 1
echo

echo "=== PATH ==="
echo "$PATH"
echo

# counter was installed alongside go version 1.24.2
declare -i TESTS
declare COUNTER_DIR
mkdir -p .counters
COUNTER_DIR="$(realpath ".counters")"
export COUNTER_DIR
export COUNTER_USE_FORCE=1
TESTS=$(counter -name "tests-completed" -set 7)
function test_completed() {
  TESTS=$(counter -name "tests-completed" -add)
  export TESTS
  echo $TESTS
}

# we'll use the counter during the test
TESTS=$(test_completed) # counter was set up
ID=$(genwordpass)
TEST_ID="$TESTS-$ID"
echo "TEST_ID: $TEST_ID"
TESTS=$(test_completed) # genwordpass was successfully consumed

# List installed versions (should see both with 1.24.2 activated)
echo "=== LISTING GO VERSIONS ==="
igo -cmd list || exit 1
TESTS=$(test_completed)
echo

exit 0

# Install Go 1.24.3
echo "=== INSTALLING GO 1.24.3 ==="
igo -cmd install -gover 1.24.3 $DEBUG $VERBOSE || exit 1
echo

exit 0

# Switch to Go 1.24.3
echo "=== SWITCHING TO GO 1.24.3 ==="
igo -cmd use -gover 1.24.3
echo

exit 0

# Verify Go version is 1.24.3
echo "=== VERIFYING GO 1.24.3 ==="
source ~/.bashrc || source ~/.zshrc || echo "Failed to source shell config"
go version | grep "go1.24.3" && echo "Go 1.24.3 verified!" || echo "FAIL: Go 1.24.3 not active"
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
