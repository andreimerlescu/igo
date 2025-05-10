#!/bin/bash
# shellcheck disable=SC2086
SECONDS=0
VERSION="$(cat VERSION)"

declare -A params=()
params[build]="true"
params[rm]=""
params[debug]="false"
params[verbose]="false"


declare -A documentation=()
documentation[build]="Build the Docker image"
documentation[rm]="Remove the Docker image"

source params.sh
parse_arguments "$@"

DEBUG="${params[debug]}"
if [[ -n "$DEBUG" ]]; then
  DEBUG="--debug"
fi

BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "Branch: $BRANCH"
BRANCH="$(echo "$BRANCH" | tr '/' '-')"
TEST_ID="${VERSION}q$(counter -name "igo-tests-${BRANCH}" -add)"
echo "Test ID: $TEST_ID"

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

if [[ "${params[build]}" == "true" ]]; then
  docker build -t "igo:${VERSION}" . || { echo "Docker build failed"; exit 1; }
fi
docker tag "igo:${VERSION}" "igo:${TEST_ID}" || { echo "Docker tag failed"; exit 1; }

chmod +x tester.sh

echo "Running tests in container..."
if ! docker $DEBUG run --rm --env=TEST_ID=$TEST_ID --env=BRANCH=$BRANCH --env=VERSION=$VERSION --env=DEBUG=$DEBUG --entrypoint "/home/tester/tester.sh" "igo:$TEST_ID"; then
  echo "Tests failed - took $SECONDS seconds"
  exit 1
else
  echo "Tests completed successfully in $SECONDS seconds!"
  exit 0
fi
