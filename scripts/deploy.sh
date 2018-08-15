#!/bin/bash
set -e

main() {
  current_tag="$(shell git tag --points-at HEAD)"
  if [ -z "$current_tag" ]; then
    echo "No tag to deploy"
    return
  fi
  make all
  go build -o ./build/release ./cmd/tkrelease
  echo "Deploying $current_tag"
  ./build/release -k="$AWS_KEY" -s="$AWS_SECRET" $current_tag
  make gen_sha
}

main "$@"
