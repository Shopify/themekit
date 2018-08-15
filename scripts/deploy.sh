#!/bin/bash
set -e

main() {
  if [ -z "$BUILDKITE_TAG" ]; then
    echo "No tag to deploy"
    return
  fi
  make all
  go build -o ./build/release ./cmd/tkrelease
  echo "Deploying $BUILDKITE_TAG"
  ./build/release -k="$AWS_KEY" -s="$AWS_SECRET" $BUILDKITE_TAG
  make gen_sha
}

main "$@"
