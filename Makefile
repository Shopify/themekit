SUBPROJECTS = theme

all:
	for subproject in $(SUBPROJECTS); \
	do \
		mkdir -p build/development; \
	  godep go build -o build/development/$${subproject} cmd/$${subproject}/*.go; \
  done

build:
	for subproject in $(SUBPROJECTS); \
	do \
	  mkdir -p build/dist/${GOOS}-${GOARCH}; \
		godep go build -o build/dist/${GOOS}-${GOARCH}/$${subproject} cmd/$${subproject}/*.go; \
	done

clean:
	rm -rf build/

.PHONY: all build clean zip

build64:
	export GOARCH=amd64; $(MAKE) build

build32:
	export GOARCH=386; $(MAKE) build

windows:
	export GOOS=windows; $(MAKE) build64
	export GOOS=windows; $(MAKE) build32

mac:
	export GOOS=darwin; $(MAKE) build64

linux:
	export GOOS=linux; $(MAKE) build64
	export GOOS=linux; $(MAKE) build32

zip:
	./compress

dist: windows mac linux zip

