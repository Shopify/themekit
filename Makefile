PACKAGES = ./kit/... ./commands/... ./theme/...

.PHONY: all build clean zip help

test: ## Run all tests
	@go test -v -race $(PACKAGES);

vet: ## Lint a verify go code.
	@go vet $(PACKAGES);

dist: clean  ## Build binaries for all platforms, zip, and upload to S3
	@$(MAKE) windows && $(MAKE) mac && $(MAKE) linux && $(MAKE) zip && $(MAKE) upload_to_s3;

clean: ## Remove all temporary build artifacts
	@rm -rf build && echo "project cleaned";

all:
	@mkdir -p build/development && go build -ldflags="-s -w" -o build/development/theme github.com/Shopify/themekit/cmd/theme;

build:
	@mkdir -p build/dist/${GOOS}-${GOARCH} && go build -ldflags="-s -w" -o build/dist/${GOOS}-${GOARCH}/theme${EXT} github.com/Shopify/themekit/cmd/theme;

install: # Build and install the theme binary
	@go install github.com/Shopify/themekit/cmd/theme;

build64:
	@export GOARCH=amd64; $(MAKE) build;

build32:
	@export GOARCH=386; $(MAKE) build;

windows: ## Build binaries for Windows (32 and 64 bit)
	@echo "building win-64" && export GOOS=windows; export EXT=.exe; $(MAKE) build64 && echo "win-64 build complete";
	@echo "building win-64" && export GOOS=windows; export EXT=.exe; $(MAKE) build32 && echo "win-64 build complete";

mac: ## Build binaries for Mac OS X (64 bit)
	@echo "building darwin-64" && export GOOS=darwin; $(MAKE) build64 && echo "darwin-64 build complete";

linux: ## Build binaries for Linux (32 and 64 bit)
	@echo "building linux-64" && export GOOS=linux; $(MAKE) build64 && echo "linux-64 build complete";
	@echo "building linux-32" && export GOOS=linux; $(MAKE) build32 && echo "linux-32 build complete";

zip: ## Create zip file with distributable binaries
	@echo "compressing releases" && ./scripts/compress && echo "finished compressing";

upload_to_s3: ## Upload zip file with binaries to S3
	@echo "uploading to S3" && bundle exec ruby ./scripts/release && echo "upload complete";

help:
	@grep -E '^[a-zA-Z_0-9-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
