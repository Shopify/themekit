.PHONY: build clean zip help

install: # Build and install the theme binary
	@go install github.com/Shopify/themekit/cmd/theme;

test: ## Run all tests
	@go test -race -cover $(shell glide novendor)

vet: ## Verify go code.
	@go vet $(shell glide novendor)

lint: ## Lint all packages
	@glide novendor | xargs -n 1 golint -set_exit_status

check: lint vet test # lint, vet and test the code

dist: lint vet test clean  ## Build binaries for all platforms, zip, and upload to S3
	@$(MAKE) windows && $(MAKE) mac && $(MAKE) linux && $(MAKE) zip && $(MAKE) upload_to_s3;

clean: ## Remove all temporary build artifacts
	@rm -rf build && echo "project cleaned";

build:
	@mkdir -p build/dist/${GOOS}-${GOARCH} && go build -ldflags="-s -w" -o build/dist/${GOOS}-${GOARCH}/theme${EXT} github.com/Shopify/themekit/cmd/theme;

build64:
	@export GOARCH=amd64; $(MAKE) build;

build32:
	@export GOARCH=386; $(MAKE) build;

windows: ## Build binaries for Windows (32 and 64 bit)
	@echo "building win-64" && export GOOS=windows; export EXT=.exe; $(MAKE) build64 && echo "win-64 build complete";
	@echo "building win-32" && export GOOS=windows; export EXT=.exe; $(MAKE) build32 && echo "win-32 build complete";

mac: ## Build binaries for Mac OS X (64 bit)
	@echo "building darwin-64" && export GOOS=darwin; $(MAKE) build64 && echo "darwin-64 build complete";

linux: ## Build binaries for Linux (32 and 64 bit)
	@echo "building linux-64" && export GOOS=linux; $(MAKE) build64 && echo "linux-64 build complete";
	@echo "building linux-32" && export GOOS=linux; $(MAKE) build32 && echo "linux-32 build complete";
	@echo "building deb" && ./scripts/build_deb && echo "deb build complete";

zip: ## Create zip file with distributable binaries
	@echo "compressing releases" && ./scripts/compress && echo "finished compressing";

upload_to_s3: ## Upload zip file with binaries to S3
	@echo "uploading to S3" && bundle exec ruby ./scripts/release && echo "upload complete";

serve_docs: ## Start the dev server for the jekyll static site serving the theme kit docs.
	@cd docs && jekyll serve

tools: ## Installs tools required for developing theme kit
	@curl https://glide.sh/get | sh
	@go get -u github.com/golang/lint/golint
	@gem install jekyll

help:
	@grep -E '^[a-zA-Z_0-9-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
