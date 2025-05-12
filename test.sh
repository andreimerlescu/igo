#!/bin/bash
# shellcheck disable=SC2086
START_TIME=$(date +%s.%N)
SECONDS=0
VERSION="$(cat VERSION)"

# Define command line arguments
declare -A params=()
params[build]="true"
params[rm]=""
params[debug]="false"
params[verbose]="false"
params[clear]="true"

# Define documentation for each parameter
declare -A documentation=()
documentation[build]="Build the Docker image"
documentation[rm]="Remove the Docker image"
documentation[debug]="Enable debug mode"
documentation[verbose]="Enable verbose mode"
documentation[clear]="Clear console before starting"

# Include params helper
source params.sh

# Parse command line arguments
parse_arguments "$@"

CLEAR="${params[clear]}"
if [[ -n "$CLEAR" ]] && [[ "${CLEAR}" != "false" ]]; then
  clear
fi

# Parse debug mode
DEBUG="${params[debug]}"
if [[ -n "$DEBUG" ]] && [[ "${DEBUG}" != "false" ]]; then
  DEBUG="--debug"
fi
[[ "${DEBUG}" == "false"  ]] && DEBUG=""

# Parse verbose mode
VERBOSE="${params[verbose]}"
if [[ -n "$VERBOSE" ]] && [[ "${VERSION}" != "false" ]]; then
  VERBOSE="--verbose"
fi
[[ "${VERBOSE}" == "false"  ]] && VERBOSE=""

# Parse branch name
BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "Branch: $BRANCH"
BRANCH="$(echo "$BRANCH" | tr '/' '-')"

# Prepare the counter
COUNTER_NAME="igo-tests-${BRANCH}"
echo "Using counter name: $COUNTER_NAME"
TEST_ID="${VERSION}q$(counter -name "${COUNTER_NAME}" -add)"
echo "Test ID: $TEST_ID"

# Remove old images and containers
echo "Docker Image: igo:${TEST_ID}"
if [[ "${params[rm]}" == "true" ]]; then
  echo "Removing all igo:${VERSION}q* images..."
  IMAGES_TO_REMOVE=$(docker images --format "{{.Repository}}:{{.Tag}}" "igo:${VERSION}*")
  if [ -n "$IMAGES_TO_REMOVE" ]; then
    echo "$IMAGES_TO_REMOVE" | xargs docker rmi || echo "Failed to remove some images"
  else
    echo "No matching images found to remove"
  fi
  docker rmi "igo:${VERSION}" || echo "can not remove non-existent igo:${VERSION}"
fi

# Build the Docker image
if [[ "${params[build]}" == "true" ]]; then
  docker build -t "igo:${VERSION}" . || { echo "Docker build failed"; exit 1; }
fi
docker tag "igo:${VERSION}" "igo:${TEST_ID}" || { echo "Docker tag failed"; exit 1; }

# Ensure tester is executable
chmod +x tester.sh

# Run the tests
echo "Running tests in container '${DEBUG}'..."
if ! docker $DEBUG run --rm --env=TEST_ID=$TEST_ID --env=BRANCH=$BRANCH --env=VERSION=$VERSION --env=DEBUG=$DEBUG --env=VERBOSE=$VERBOSE --entrypoint "/home/tester/tester.sh" "igo:$TEST_ID"; then
  END_TIME=$(date +%s.%N)
  DURATION=$(echo "$END_TIME - $START_TIME" | bc)
  echo "Tests failed - took $DURATION seconds"
  exit 1
else
  END_TIME=$(date +%s.%N)
  DURATION=$(echo "$END_TIME - $START_TIME" | bc)
  echo "Tests completed successfully in $DURATION seconds!"
  exit 0
fi
