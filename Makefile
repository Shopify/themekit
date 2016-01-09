SUBPROJECTS = theme

all:
	for subproject in $(SUBPROJECTS); \
	do \
		mkdir -p build/development; \
	  godep go build -o build/development/$${subproject} github.com/Shopify/themekit/cmd/$${subproject}; \
  done

install: 
	godep go install github.com/Shopify/themekit/cmd/theme 

build:
	for subproject in $(SUBPROJECTS); \
	do \
	  mkdir -p build/dist/${GOOS}-${GOARCH}; \
		godep go build -o build/dist/${GOOS}-${GOARCH}/$${subproject}${EXT} github.com/Shopify/themekit/cmd/$${subproject}; \
	done

test:
	go test \
	github.com/Shopify/themekit \
	github.com/Shopify/themekit/atom \
	github.com/Shopify/themekit/bucket \
	github.com/Shopify/themekit/commands \
	github.com/Shopify/themekit/theme

clean:
	rm -rf build/

.PHONY: all build clean zip

build64:
	export GOARCH=amd64; $(MAKE) build

build32:
	export GOARCH=386; $(MAKE) build

windows:
	export GOOS=windows; export EXT=.exe; $(MAKE) build64
	export GOOS=windows; export EXT=.exe; $(MAKE) build32

mac:
	export GOOS=darwin; $(MAKE) build64

linux:
	export GOOS=linux; $(MAKE) build64
	export GOOS=linux; $(MAKE) build32

zip:
	./compress

upload_to_s3:
	./release

dist: clean windows mac linux zip upload_to_s3
