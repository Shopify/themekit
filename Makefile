.PHONY: build
install: bundle ## Build and install the theme binary
	@go install github.com/Shopify/themekit/cmd/theme;
lint: # Lint all packages
ifeq ("$(shell which golint 2>/dev/null)","")
	@go get -u golang.org/x/lint/golint
endif
	@golint -set_exit_status ./...
test: lint ## lint and test the code
	@go test -race -cover -covermode=atomic ./...
clean: ## Remove all temporary build artifacts
	@rm -rf build && echo "project cleaned";
all: clean bundle ## will build a binary for all platforms
	@export GOOS=windows GOARCH=386 EXT=.exe; $(MAKE) build;
	@export GOOS=windows GOARCH=amd64 EXT=.exe; $(MAKE) build;
	@export GOOS=darwin GOARCH=386; $(MAKE) build;
	@export GOOS=darwin GOARCH=amd64; $(MAKE) build;
	@export GOOS=linux GOARCH=386; $(MAKE) build;
	@export GOOS=linux GOARCH=amd64; $(MAKE) build;
	@export GOOS=freebsd GOARCH=386; $(MAKE) build;
	@export GOOS=freebsd GOARCH=amd64; $(MAKE) build;
build:
	@mkdir -p build/dist/${GOOS}-${GOARCH} && \
    echo "[${GOOS}-${GOARCH}] build started" && \
		go build \
			-ldflags="-s -w" \
			-o build/dist/${GOOS}-${GOARCH}/theme${EXT} \
			github.com/Shopify/themekit/cmd/theme && \
		echo "[${GOOS}-${GOARCH}] build complete";
bundle:
	@go run src/static/cmd/bundle.go
sha: ## Generate sha256 for a darwin build for usage with homebrew
	@shasum -a 256 ./build/dist/darwin-amd64/theme
md5s: ## Generate md5 sums for all builds
	@echo "darwinamd64sum: $(shell md5 -q ./build/dist/darwin-amd64/theme)"
	@echo "windows386sum: $(shell md5 -q ./build/dist/windows-386/theme.exe)"
	@echo "windowsamd64sum: $(shell md5 -q ./build/dist/windows-amd64/theme.exe)"
	@echo "linux386sum: $(shell md5 -q ./build/dist/linux-386/theme)"
	@echo "linuxamd64sum: $(shell md5 -q ./build/dist/linux-amd64/theme)"
	@echo "freebsd386sum: $(shell md5 -q ./build/dist/freebsd-386/theme)"
	@echo "freebsdamd64sum: $(shell md5 -q ./build/dist/freebsd-amd64/theme)"
serve_docs: ## Start the dev server for the jekyll static site serving the theme kit docs.
ifeq ("$(shell which jekyll 2>/dev/null)","")
	@gem install jekyll
endif
	@cd docs && jekyll serve
help:
	@grep -E '^[a-zA-Z_0-9-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
