.PHONY: build
install: ## Build and install the theme binary
	@go install github.com/Shopify/themekit/cmd/theme;
vet:
	@go vet $(shell go list ./...)
lint: # Lint all packages
	@go list ./... | xargs -n 1 golint -set_exit_status
test: lint vet unit_test # lint, vet and test the code
	@go test -race -cover $(shell go list ./...)
all: clean # will build a binary for all platforms
	@export GOOS=windows GOARCH=386 EXT=.exe; $(MAKE) build;
	@export GOOS=windows GOARCH=amd64 EXT=.exe; $(MAKE) build;
	@export GOOS=darwin GOARCH=386; $(MAKE) build;
	@export GOOS=darwin GOARCH=amd64; $(MAKE) build;
	@export GOOS=linux GOARCH=386; $(MAKE) build;
	@export GOOS=linux GOARCH=amd64; $(MAKE) build;
release: # will run release on the built binaries uploading them to S3
	@go install github.com/Shopify/themekit/cmd/tkrelease && tkrelease $(shell git tag --points-at HEAD)
clean: # Remove all temporary build artifacts
	@rm -rf build && echo "project cleaned";
build:
	@mkdir -p build/dist/${GOOS}-${GOARCH} && \
    echo "[${GOOS}-${GOARCH}] build started" && \
		go build \
			-ldflags="-s -w" \
			-o build/dist/${GOOS}-${GOARCH}/theme${EXT} \
			github.com/Shopify/themekit/cmd/theme && \
		echo "[${GOOS}-${GOARCH}] build complete";
gen_sha: # Generate sha256 for a darwin build for usage with homebrew
	@shasum -a 256 ./build/dist/darwin-amd64/theme
serve_docs: ## Start the dev server for the jekyll static site serving the theme kit docs.
	@cd docs && jekyll serve
help: ## Prints this message
	@grep -E '^[a-zA-Z_0-9-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
