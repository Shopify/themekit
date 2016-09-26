SUBPROJECTS = theme

all:
	for subproject in $(SUBPROJECTS); \
	do \
		mkdir -p build/development; \
	  godep go build -o build/development/$${subproject} github.com/Shopify/themekit/cmd/$${subproject}; \
  done

install: # Build and install the theme binary
	godep go install github.com/Shopify/themekit/cmd/theme

build:
	for subproject in $(SUBPROJECTS); \
	do \
	  mkdir -p build/dist/${GOOS}-${GOARCH}; \
		godep go build -o build/dist/${GOOS}-${GOARCH}/$${subproject}${EXT} github.com/Shopify/themekit/cmd/$${subproject}; \
	done


debug: # Example: 'make debug ARGS="version" or 'make debug ARGS="remove templates/404.liquid"'
	cd cmd/theme &&	godebug run -instrument=github.com/Shopify/themekit,github.com/Shopify/themekit/commands main.go $(ARGS)

test: ## Run all tests
	go test -v -race \
	github.com/Shopify/themekit/kit \
	github.com/Shopify/themekit/atom \
	github.com/Shopify/themekit/commands \
	github.com/Shopify/themekit/theme

clean: ## Remove all temporary build artifacts
	rm -rf build/

.PHONY: all build clean zip help

build64:
	export GOARCH=amd64; $(MAKE) build

build32:
	export GOARCH=386; $(MAKE) build

windows: ## Build binaries for Windows (32 and 64 bit)
	export GOOS=windows; export EXT=.exe; $(MAKE) build64
	export GOOS=windows; export EXT=.exe; $(MAKE) build32

mac: ## Build binaries for Mac OS X (64 bit)
	export GOOS=darwin; $(MAKE) build64

linux: ## Build binaries for Linux (32 and 64 bit)
	export GOOS=linux; $(MAKE) build64
	export GOOS=linux; $(MAKE) build32

zip: ## Create zip file with distributable binaries
	./scripts/compress

upload_to_s3: ## Upload zip file with binaries to S3
	./scripts/release

dist: clean windows mac linux zip upload_to_s3 ## Build binaries for all platforms, zip, and upload to S3

help:
	@grep -E '^[a-zA-Z_0-9-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
