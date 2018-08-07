#!/bin/bash
set -e
# This block for generating coverage.txt was lifted from the codecov docs:
# https://github.com/codecov/example-go#caveat-multiple-files
echo "" > coverage.txt
for d in $(go list ./...); do
    go test -timeout 15s -race -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
bash <(curl -s https://codecov.io/bash)
