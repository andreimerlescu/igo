#!/bin/bash

set -e  # BEST PRACTICES: Exit immediately if a command exits with a non-zero status
set -u  # SECURITY: Exit if an unset variable is used to prevent potential security risks
set -C  # SECURITY: Prevent existing files from being overwritten using the '>' operator
[ -n "${DEBUG:-}" ] && [ "${DEBUG}" == "true" ] && set -x  # DEVELOPER EXPERIENCE: Enable debug mode
[ -n "${VERBOSE:-}" ] && [ "${VERBOSE}" == "true" ] && set -v  # DEVELOPER EXPERIENCE: Enable verbose mode

if [ -n "${DEBUG:-}" ] && [ "${DEBUG}" == "true" ]; then
  echo "RUNNING SHIM GO: DEBUG=${DEBUG} VERBOSE=${VERBOSE}"
fi

declare GODIR
GODIR="${HOME:-"/home/$(whoami)"}/go"

function safe_exit() {
  echo "ERROR: $1" >&2
  exit 1
}

get_go_binary_path_for_version() {
    local version="$1"
    if [ ! -f "${GODIR}/versions/${version}/go/bin/go.${version}" ]; then
      local GOVERSION=""
      [ -f "$PWD/.go_version" ] && GOVERSION="$(cat "$PWD/.go_version")"
    else
      echo "${GODIR}/versions/${version}/go/bin/go.${version}"
    fi
}

find_version() {
  local dir="$PWD"
  while [[ "$dir" != "/" ]]; do
    if [[ -f "$dir/.go_version" ]]; then
      cat "$dir/.go_version"
      return
    fi
    if [[ -f "$dir/go.mod" ]]; then
      local gomod_version
      gomod_version=$(grep -E "^go [0-9]+\.[0-9]+(\.[0-9]+)?" "$dir/go.mod" | awk '{print $2}')
      if [[ -n "$gomod_version" ]]; then
        if [[ "$gomod_version" =~ ^[0-9]+\.[0-9]+$ ]]; then
          echo "${gomod_version}.0"
        else
          echo "$gomod_version"
        fi
        return
      fi
    fi
    dir="$(dirname "$dir")"
  done

  if [ ! -f "${GODIR}/version" ]; then
    safe_exit "No global Go version installed at ${GODIR}/version."
  fi
  cat "${GODIR}/version"
}

version="$(find_version)"
if [ "${version}" == "" ]; then
  safe_exit "Invalid version detected."
fi

# Invoke the real go binary with any arguments passed to the shim
GOBINARY="$(get_go_binary_path_for_version "${version}")"
[ "${GOBINARY}"  == "" ] && safe_exit "a .go_version is set to '${version}' but it isn't installed yet"

GOBIN="${GODIR}/versions/${version}/go/bin"
GOROOT="${GODIR}/versions/${version}/go"
GOPATH="${GODIR}/versions/${version}"

export GOBIN
export GOROOT
export GOPATH

exec "${GOBINARY}" "$@"
