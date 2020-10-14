#!/usr/bin/env bash

git checkout cmd/static/generated-assets.go
git checkout src/static/_testdata/static/generated-assets.go

if [ "$(git diff)" ]; then
  echo "Oops, something changed!"
  exit 1
fi
