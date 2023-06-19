.PHONY: build
install: bundle ## Build and install the theme binary
	@CGO_ENABLED=0 go install github.com/Shopify/themekit/cmd/theme;
lint: # Lint all packages
ifeq ("$(shell which golint 2>/dev/null)","")
	@go get -u golang.org/x/lint/golint
endif
	@golint -set_exit_status ./...
test: lint ## lint and test the code
	go test -race -cover -covermode=atomic ./...
clean: ## Remove all temporary build artifacts
	@rm -rf build && echo "project cleaned";
all: clean bundle ## will build a binary for all platforms
	@export GOOS=windows GOARCH=386 EXT=.exe; $(MAKE) build;
	@export GOOS=windows GOARCH=amd64 EXT=.exe; $(MAKE) build;
	@export GOOS=darwin GOARCH=amd64; $(MAKE) build;
	@export GOOS=linux GOARCH=386; $(MAKE) build;
	@export GOOS=linux GOARCH=amd64; $(MAKE) build;
	@export GOOS=freebsd GOARCH=386; $(MAKE) build;
	@export GOOS=freebsd GOARCH=amd64; $(MAKE) build;
build:
	@mkdir -p build/release
	@mkdir -p build/dist/${GOOS}-${GOARCH} && \
    echo "[${GOOS}-${GOARCH}] build started" && \
		CGO_ENABLED=0 go build \
			-ldflags="-s -w" \
			-o build/dist/${GOOS}-${GOARCH}/theme${EXT} \
			github.com/Shopify/themekit/cmd/theme && \
		zip --quiet --junk-paths build/release/${GOOS}-${GOARCH}.zip \
			build/dist/${GOOS}-${GOARCH}/theme${EXT} && \
		echo "[${GOOS}-${GOARCH}] build complete";
bundle:
	@go run src/static/cmd/bundle.go
sha: ## Generate sha256 for a darwin build for usage with homebrew
	@shasum -a 256 ./build/dist/darwin-amd64/theme

md5s: ## Generate md5 sums for all builds
	@echo "| OS      | Architecture | md5 checksums              |"
	@echo "| :------ | :----------- | :------------------------- |"
	@echo "| macOS   | 64-bit       | $(shell md5 -q ./build/dist/darwin-amd64/theme)|"
	@echo "| Windows | 64-bit       | $(shell md5 -q ./build/dist/windows-amd64/theme.exe)|"
	@echo "| Windows | 32-bit       | $(shell md5 -q ./build/dist/windows-386/theme.exe)|"
	@echo "| Linux   | 64-bit       | $(shell md5 -q ./build/dist/linux-amd64/theme)|"
	@echo "| Linux   | 32-bit       | $(shell md5 -q ./build/dist/linux-386/theme)|"
	@echo "| FreeBSD | 64-bit       | $(shell md5 -q ./build/dist/freebsd-amd64/theme)|"
	@echo "| FreeBSD | 32-bit       | $(shell md5 -q ./build/dist/freebsd-386/theme)|"
help:
	@grep -E '^[a-zA-Z_0-9-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
