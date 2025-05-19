#!/bin/bash
MAX_EXEC_TIME=300
VERSION="$(cat VERSION)"
export VERSION
handle_timeout() {
  echo "Script timed out after ${MAX_EXEC_TIME} seconds"
  exit 1
}

cleanup() {
  echo "Caught Ctrl-C, cleaning up..."
  pkill -P $$ || true
  echo "Terminated running tasks."
  exit 1
}

declare -A params=()
declare -A documentation=()

source testing/cli.sh

export params
export documentation

trap cleanup SIGINT

# 12m57s timeout for test-in-timeout.sh where tester.sh is invoked inside docker
# 777 => 12m57s => 12 57 ms => 6 9 ms => 6 9 (13) (19) => 369 9/11
# 777 => 369 9/11 [[[ REALITY IS A PROGRAM ]]]
# PROGRAM OR BE PROGRAMMED
timeout --foreground --kill-after=777s "${MAX_EXEC_TIME}s" bash <<'EOF' || handle_timeout
source testing/test-in-timeout.sh
EOF
