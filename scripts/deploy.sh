#!/bin/bash
set -e

main() {
  make all
  current_tag="$(git tag --points-at HEAD)"
  if [ -z "$current_tag" ]; then
    echo "No tag to deploy"
    return
  fi
  go build -o ./build/release ./cmd/tkrelease
  echo "Deploying $current_tag"
  ./build/release -k="$AWS_KEY" -s="$AWS_SECRET" $current_tag
  make gen_sha
}

main "$@"
