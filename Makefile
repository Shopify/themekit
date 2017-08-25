.PHONY: build
install: ## Build and install the theme binary
	@go install github.com/Shopify/themekit/cmd/theme;
test: ## Run all tests
	@go test -race -cover $(shell glide novendor)
vet: ## Verify go code.
	@go vet $(shell glide novendor)
lint: ## Lint all packages
	@glide novendor | xargs -n 1 golint -set_exit_status
check: lint vet test ## lint, vet and test the code
all: clean windows mac linux ## will build a binary for all platforms
release: ## will run release on the built binaries uploading them to S3
	@go install github.com/Shopify/themekit/cmd/tkrelease && tkrelease $(shell git tag --points-at HEAD)
dist: check all ## Build binaries for all platforms and upload to S3
	@$(MAKE) release && $(MAKE) gen_sha
clean: ## Remove all temporary build artifacts
	@rm -rf build && echo "project cleaned";
build:
	@mkdir -p build/dist/${GOOS}-${GOARCH} && \
    echo "[${GOOS}-${GOARCH}] build started" && \
		go build \
			-ldflags="-s -w" \
			-o build/dist/${GOOS}-${GOARCH}/theme${EXT} \
			github.com/Shopify/themekit/cmd/theme && \
		echo "[${GOOS}-${GOARCH}] build complete";
windows: win32 win64 # Build binaries for Windows (32 and 64 bit)
win32:
	@export GOOS=windows GOARCH=386 EXT=.exe; $(MAKE) build;
win64:
	@export GOOS=windows GOARCH=amd64 EXT=.exe; $(MAKE) build;
mac: mac32 mac64 # Build binaries for Mac OS X (64 bit)
mac32:
	@export GOOS=darwin GOARCH=386; $(MAKE) build;
mac64:
	@export GOOS=darwin GOARCH=amd64; $(MAKE) build;
linux: lin32 lin64 # Build binaries for Linux (32 and 64 bit)
lin32:
	@export GOOS=linux GOARCH=386; $(MAKE) build;
lin64:
	@export GOOS=linux GOARCH=amd64; $(MAKE) build;
gen_sha: ## Generate sha256 for a darwin build for usage with homebrew
	@shasum -a 256 ./build/dist/darwin-amd64/theme
serve_docs: ## Start the dev server for the jekyll static site serving the theme kit docs.
	@cd docs && jekyll serve
init_tools: ## Will install tools needed to work on this repo
	@curl https://glide.sh/get | sh
	@go get -u github.com/golang/lint/golint
	@gem install jekyll
help: ## Prints this message
	@grep -E '^[a-zA-Z_0-9-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
