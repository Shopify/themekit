#!/usr/bin/env bash

gofmt -s -w .

# Ignore these when checking formatting
git checkout cmd/static/generated-assets.go \
  src/static/_testdata/static/generated-assets.go \

if [ "$(git diff)" ]; then
  echo "Bad formatting detected, please run gofmt"
  exit 1
fi
