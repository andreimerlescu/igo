#!/bin/bash

# Get the version from VERSION file
WORKSPACE_DIR="${GITHUB_WORKSPACE:-$(pwd)}"
VERSION_FILE="$(cat "${WORKSPACE_DIR}/VERSION" | tr -d '\n')"
echo "Version from file: $VERSION_FILE"

# Get the latest tag (if any)
LATEST_TAG=$(git tag --sort=-v:refname | head -n 1)

if [ -z "$LATEST_TAG" ]; then
  echo "No tags found. VERSION file value is valid."
  exit 0
fi

echo "Latest tag: $LATEST_TAG"

# Remove 'v' prefix if present for comparison
VERSION_NUM=${VERSION_FILE#v}
LATEST_TAG_NUM=${LATEST_TAG#v}

# Use sort for version comparison (make sure -V is available in your environment)
NEWER_VERSION=$(printf "%s\n%s" "$VERSION_NUM" "$LATEST_TAG_NUM" | sort -V | tail -n 1)

if [ "$NEWER_VERSION" = "$VERSION_NUM" ] && [ "$VERSION_NUM" != "$LATEST_TAG_NUM" ]; then
  echo "VERSION file ($VERSION_FILE) is greater than the latest tag ($LATEST_TAG)"
else
  echo "ERROR: VERSION file ($VERSION_FILE) is NOT greater than latest tag ($LATEST_TAG)"
  exit 1
fi