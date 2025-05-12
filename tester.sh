#!/bin/bash
set -e
echo "Starting igo test script..."

declare -i TESTS
declare COUNTER_DIR

START_TIME=$(date +%s.%N)

function test_took() {
  echo "Test $TESTS took $SECONDS seconds"
  echo
}

# Initial setup
export PATH=/bin:$PATH

# Display igo version
echo "=== IGO VERSION ==="
SECONDS=0
igo -version || exit 1
TESTS=$(( TESTS + 1 ))
test_took

echo "=== IGO ENVIRONMENT ==="
igo -cmd env "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$((TESTS + 1))
test_took
echo

# List installed versions (should be empty)
echo "=== INITIAL LIST (Should be empty) ==="
SECONDS=0
igo -cmd list || exit 1
TESTS=$((TESTS + 1))
test_took

# Install Go 1.24.2
echo "=== INSTALLING GO 1.24.2 ==="
SECONDS=0
igo -cmd install -gover 1.24.2 "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$((TESTS + 1))
test_took

echo "=== IGO ENVIRONMENT ==="
SECONDS=0
igo -cmd env "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$((TESTS + 1))
test_took
echo

if command -v tree 2>&1; then
  echo "=== LISTING FILES ==="
  tree -L 3 || exit 1
  TESTS=$((TESTS + 1))
  echo
fi

echo "=== ENVIRONMENT VARIABLES ==="
env | sort || exit 1
TESTS=$((TESTS + 1))
echo

USERNAME=$(whoami | tr -d '\n')

echo "=== RELOADING SHELL CONFIG ==="
{ [ -f ~/.profile ] && source ~/.profile; echo "Loaded ~/.profile into shell..."; } || { echo "Failed to source $USERNAME shell config"; exit 1; }
{ [ -f ~/.zshrc.local ] && source ~/.zshrc.local; echo "Loaded ~/.zshrc.local"; } || { echo "Failed to source $USERNAME shell config"; exit 1; }
TESTS=$((TESTS + 1))
echo

echo "=== BASH PROFILE ==="
cat ~/.profile || exit 1
TESTS=$((TESTS + 1))
echo

echo "=== PATH ==="
echo "$PATH"
TESTS=$((TESTS + 1))
echo

# counter was installed alongside go version 1.24.2
SECONDS=0
mkdir -p .counters
COUNTER_DIR="$(realpath ".counters")"
export COUNTER_DIR
export COUNTER_USE_FORCE=1
TESTS=$(counter -name "tests-completed" -set $TESTS)
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
test_took
echo

# List installed versions (should see both with 1.24.2 activated)
echo "=== LISTING GO VERSIONS ==="
SECONDS=0
igo -cmd list "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$(test_completed)
test_took
echo

# Install Go 1.24.3
echo "=== INSTALLING GO 1.24.3 ==="
SECONDS=0
igo -cmd install -gover 1.24.3 "$DEBUG" "$VERBOSE" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== IGO ENVIRONMENT ==="
SECONDS=0
igo -cmd env "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== RELOADING ENVIRONMENT ==="
SECONDS=0
{ [ -f ~/.profile ] && source ~/.profile; echo "Loaded ~/.profile into shell..."; } || { echo "Failed to source $USERNAME shell config"; exit 1; }
{ [ -f ~/.zshrc.local ] && source ~/.zshrc.local; echo "Loaded ~/.zshrc.local in to shell..."; } || { echo "Failed to source $USERNAME shell config"; exit 1; }
TESTS=$(test_completed)
test_took
echo

echo "=== LISTING ~/go FILES ==="
SECONDS=0
ls -la ~/go || exit 1
ls -la "$(realpath ~/go/bin)" || exit 1
ls -la "$(realpath ~/go/shims)" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== ENVIRONMENT VARIABLES ==="
SECONDS=0
env | sort || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== PATH ==="
SECONDS=0
echo "$PATH"
TESTS=$(test_completed)
test_took
echo

echo "=== VERIFYING INSTALLATION ==="
SECONDS=0
{ go version | grep "go1.24.3" && echo "Go 1.24.3 verified!"; } || { echo "FAIL: Go 1.24.3 not active"; exit 1; }
TESTS=$(test_completed)
test_took
echo

# List installed versions
echo "=== LISTING GO VERSIONS ==="
SECONDS=0
igo -cmd list "$DEBUG" "$VERBOSE" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== IGO ENVIRONMENT ==="
SECONDS=0
igo -cmd env "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== SWITCHING TO GO 1.24.2 ==="
SECONDS=0
igo -cmd use -gover 1.24.2 "$DEBUG" "$VERBOSE" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== IGO ENVIRONMENT ==="
SECONDS=0
igo -cmd env "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$(test_completed)
test_took
echo

# List installed versions
echo "=== LISTING GO VERSIONS ==="
SECONDS=0
igo -cmd list "$DEBUG" "$VERBOSE" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== VERIFYING INSTALLATION ==="
SECONDS=0
v=$(go version)
{ go version | grep "go1.24.2" && echo "Go $v verified!"; } || { echo "FAIL: Go 1.24.2 not active; got $v"; exit 1; }
unset v
TESTS=$(test_completed)
test_took
echo

echo "=== SWITCHING TO GO 1.24.3 ==="
SECONDS=0
igo -cmd use -gover 1.24.3 "$DEBUG" "$VERBOSE" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== IGO ENVIRONMENT ==="
SECONDS=0
igo -cmd env "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$(test_completed)
test_took
echo

# List installed versions
echo "=== LISTING GO VERSIONS ==="
SECONDS=0
igo -cmd list "$DEBUG" "$VERBOSE" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== VERIFYING INSTALLATION ==="
SECONDS=0
v=$(go version)
{ go version | grep "go1.24.3" && echo "Go $v verified!"; } || { echo "FAIL: Go 1.24.3 not active; got $v"; exit 1; }
unset v
TESTS=$(test_completed)
test_took
echo

# Remove Go 1.24.2
echo "=== REMOVING GO 1.24.2 ==="
SECONDS=0
igo -cmd uninstall -gover 1.24.2 "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$(test_completed)
test_took
echo

echo "=== IGO ENVIRONMENT ==="
SECONDS=0
igo -cmd env "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$(test_completed)
test_took
echo

# List installed versions
echo "=== LISTING GO VERSIONS (After removing 1.24.2) ==="
SECONDS=0
igo -cmd list "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$(test_completed)
test_took
echo

# Remove Go 1.24.3
echo "=== REMOVING GO 1.24.3 ==="
SECONDS=0
igo -cmd uninstall -gover 1.24.3 "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$((TESTS + 1))
test_took
echo

# List installed versions (should be empty)
echo "=== LISTING GO VERSIONS (Should be empty) ==="
SECONDS=0
igo -cmd list "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$((TESTS + 1))
test_took
echo

echo "=== IGO ENVIRONMENT ==="
SECONDS=0
igo -cmd env "${DEBUG}" "${VERBOSE}" || exit 1
TESTS=$((TESTS + 1))
test_took
echo

END_TIME=$(date +%s.%N)
DURATION=$(echo "$END_TIME - $START_TIME" | bc)
echo "Completed $TESTS tests in $DURATION seconds!"
