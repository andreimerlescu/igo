#!/bin/bash

set -e  # BEST PRACTICES: Exit immediately if a command exits with a non-zero status
set -u  # SECURITY: Exit if an unset variable is used to prevent potential security risks
set -C  # SECURITY: Prevent existing files from being overwritten using the '>' operator
[ -n "${DEBUG:-}" ] && [ "${DEBUG:-}" != "false" ] && set -x  # DEVELOPER EXPERIENCE: Enable debug mode
[ -n "${VERBOSE:-}" ] && [ "${VERBOSE:-}" != "false" ] && set -v  # DEVELOPER EXPERIENCE: Enable verbose mode

if [ -n "${DEBUG:-}" ] && [ "${DEBUG:-}" != "false" ]; then
  echo "RUNNING SHIM GO: DEBUG=${DEBUG:-} VERBOSE=${VERBOSE:-}"
fi

declare GODIR
GODIR="${HOME:-"/home/$(whoami)"}/go"

function safe_exit() {
  echo "ERROR: $1" >&2
  exit 1
}

get_go_binary_path_for_version() {
    local binary="${GODIR}/versions/${1}/go/bin/gofmt.${1}"
    { [ -f "$binary" ] && echo "$binary"; } || echo ""
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
      gomod_version=$(grep -E "^go [0-9]+\.[0-9]+(\.[0-9]+|[a-zA-Z0-9]+)?" "$dir/go.mod" | awk '{print $2}')
      if [[ -n "$gomod_version" ]]; then
        if [[ "$gomod_version" =~ ^[0-9]+\.[0-9]+$ ]]; then
          echo "${gomod_version}.0"
        else
          echo "$gomod_version"
        fi
        return
      fi
    fi
    dir=$(dirname "$dir")
  done
  if [ ! -f "${GODIR}/version" ]; then
    safe_exit "No global Go version installed at ${GODIR}/version."
  fi
  cat "${GODIR}/version"
}

# Invoke the real go binary with any arguments passed to the shim
GOVERSION="$(find_version)"
GOBINARY="$(get_go_binary_path_for_version "${GOVERSION}")"
if [[ -z "${GOBINARY}" ]]; then
  igo -i "${GOVERSION}" || safe_exit "Failed to install Go version ${GOVERSION}"
fi

GOBIN="${GODIR}/versions/${GOVERSION}/go/bin"
GOROOT="${GODIR}/versions/${GOVERSION}/go"
GOPATH="${GODIR}/versions/${GOVERSION}"
GOMODCACHE="${GODIR}/versions/${GOVERSION}/go/pkg/mod"

export GOBIN
export GOROOT
export GOPATH
export GOMODCACHE

exec "${GOBINARY}" "$@"
