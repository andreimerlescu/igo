#!/bin/bash

START_TIME=$(date +%s.%N)
SECONDS=0
declare -i BUILD_TIME=0
{
  if ! command -v docker >/dev/null; then
    echo "docker is not installed"
    exit 1
  fi

  if ! command -v counter >/dev/null; then
    echo "counter is not installed"
    go install github.com/andreimerlescu/counter@latest
    echo "counter installed"
    echo "counter version: $(counter -v)"
  fi

  if ! command -v govulncheck >/dev/null; then
    echo "govulncheck is not installed"
    go install golang.org/x/vuln/cmd/govulncheck@latest
    echo "govulncheck installed"
    echo "govulncheck version: $(govulncheck -v)"
  fi
}
echo "=== START GO INSTALLER TEST ==="

echo "--- CLI ARGUMENTS ---"

declare -A params=()
params[build]="true"
params[rm]=""
params[debug]="false"
params[verbose]="false"
params[clear]="true"

declare -A documentation=()
documentation[build]="Build the Docker image"
documentation[rm]="Remove the Docker image"
documentation[debug]="Enable debug mode"
documentation[verbose]="Enable verbose mode"
documentation[clear]="Clear console before starting"

function parse_arguments() {
	while [[ $# -gt 0 ]]; do
		case "$1" in
			--help|-h|help)
				print_usage
				exit 0
				;;
			--*)
				key="${1/--/}" # Remove '--' prefix
				key="${key//-/_}" # Replace '-' with '_' to match params[key]
				if [[ -n "${2}" && "${2:0:1}" != "-" ]]; then
					params[$key]="$2"
					shift 2
				else
					echo "Error: Missing value for $1" >&2
					exit 1
				fi
				;;
			*)
				echo "Unknown option: $1" >&2
				print_usage
				exit 1
				;;
		esac
	done
}

function print_usage(){
	echo "Usage: ${0} [OPTIONS]"
	mapfile -t sorted_keys < <(for param in "${!params[@]}"; do echo "$param"; done | sort)
	local -i padSize=3;
	for param in "${sorted_keys[@]}"; do local -i len="${#param}"; (( len > padSize )) && padSize=len; done
	((padSize+=3)) # add right buffer
	for param in "${sorted_keys[@]}"; do
		local d; local p; p="${params[$param]}"; { [[ -n "${p}" ]] && [[ "${#p}" != 0 ]] && d=" (default = '${p}')"; } || d=""
		echo "       --$(pad "$padSize" "${param}") ${documentation[$param]}${d}"
	done
}

function pad() {
  printf "%-${1}s\n" "${2}";
}

parse_arguments "$@"

export params
export documentation

CLEAR="${params[clear]}"
if [[ -n "$CLEAR" ]] && [[ "${CLEAR}" != "false" ]]; then
  clear
fi

# Parse debug mode
DEBUG="${params[debug]}"
if [[ -n "$DEBUG" ]] && [[ "${DEBUG}" != "false" ]]; then
  DEBUG="--debug"
fi
[[ "${DEBUG}" != "true"  ]] && DEBUG=""

# Parse verbose mode
VERBOSE="${params[verbose]}"
if [[ -n "$VERBOSE" ]] && [[ "${VERSION}" != "false" ]]; then
  VERBOSE="--verbose"
fi
[[ "${VERBOSE}" != "true"  ]] && VERBOSE=""

# Parse branch name
BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "Branch: $BRANCH"
BRANCH="$(echo "$BRANCH" | tr '/' '-')"

# Prepare the counter
COUNTER_NAME="igo-tests-${BRANCH}"
echo "Using counter name: $COUNTER_NAME"
TEST_ID="${VERSION}q$(counter -name "${COUNTER_NAME}" -add)"
echo "Test ID: $TEST_ID"

govulncheck ./... || echo "govulncheck failed"

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
  SECONDS=0
  q=""
  if [[ "${params[verbose]}" == "true" ]]; then
    q="-q"
  fi
  nc=""
  if [[ "${params[rm]}" != "true" ]]; then
    nc="--no-cache"
  fi

  docker build $nc $q -f ./Dockerfile -t "igo:${VERSION}" .  || { echo "Docker build failed"; exit 1; }
  BUILD_TIME=$SECONDS
fi
docker tag "igo:${VERSION}" "igo:${TEST_ID}" || { echo "Docker tag failed"; exit 1; }

# Ensure tester is executable
chmod +x tester.sh

# Run the tests
declare -i code=1 # default to error
echo "Running tests in container '${DEBUG}'..."
if ! docker $DEBUG run --rm --env=TEST_ID=$TEST_ID --env=BRANCH=$BRANCH --env=VERSION=$VERSION --env=DEBUG=$DEBUG --env=VERBOSE=$VERBOSE --entrypoint "/home/tester/tester.sh" "igo:$TEST_ID"; then
  END_TIME=$(date +%s.%N)
  DURATION=$(echo "$END_TIME - $START_TIME" | bc)
  if [ "$BUILD_TIME" -gt 0 ]; then
    DURATION=$(echo "$DURATION - $BUILD_TIME" | bc)
  fi
  echo "Tests failed - built in $BUILD_TIME seconds and took $DURATION seconds to fail"
else
  END_TIME=$(date +%s.%N)
  DURATION=$(echo "$END_TIME - $START_TIME" | bc)
  if [ "$BUILD_TIME" -gt 0 ]; then
    DURATION=$(echo "$DURATION - $BUILD_TIME" | bc)
  fi
  echo "Built in $BUILD_TIME seconds. Tests completed successfully in $DURATION seconds!"
  code=0
fi

if [[ "${params[rm]}" == "true" ]]; then
  echo "Pruning docker builder..."
  docker builder prune -f -a || echo "failed to prune docker builder"

  echo "Removing all igo:${VERSION}q* images..."
  IMAGES_TO_REMOVE=$(docker images --format "{{.Repository}}:{{.Tag}}" "igo:${VERSION}*")
  if [ -n "$IMAGES_TO_REMOVE" ]; then
    echo "$IMAGES_TO_REMOVE" | xargs docker rmi || echo "Failed to remove some images"
  else
    echo "No matching images found to remove"
  fi
  docker rmi "igo:${VERSION}" || echo "can not remove non-existent igo:${VERSION}"
fi

echo "=== END TEST.SH ==="

exit $code
