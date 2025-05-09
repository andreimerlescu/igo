#!/bin/bash
# shellcheck disable=SC2086

VERSION="$(cat VERSION)"
DEBUG=${1:-}

docker rm -f igo 2> /dev/null || echo "No container to remove"
docker rmi "igo:${VERSION}" 2> /dev/null || echo "No image to remove"
docker build -t "igo:${VERSION}" . || { echo "Docker build failed"; exit 1; }

chmod +x tester.sh

echo "Running tests in container..."
docker $DEBUG run --rm --entrypoint "/home/tester/tester.sh" "igo:$VERSION"
echo "Tests completed successfully!"
