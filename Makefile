BINARY = s3apt
VET_REPORT = local/vet.report
TEST_REPORT = local/tests.xml
GOARCH = amd64

VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null | sed -e 's/^v//' || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# Symlink into GOPATH
GITHUB_USERNAME=SonOfBytes
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${BINARY}
CURRENT_DIR=$(shell pwd)

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH} -s -w"

# Build the project
all: clean deps test vet linux darwin releases

linux: deps
	cd ${BUILD_DIR}; \
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o release/s3-linux-${GOARCH} cmd/s3/s3.go ; \
	cd - >/dev/null

darwin: deps
	cd ${BUILD_DIR}; \
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o release/s3-darwin-${GOARCH} cmd/s3/s3.go ; \
	cd - >/dev/null

releases: darwin linux
	cd ${BUILD_DIR}; \
	ls -l release; \
	cd - >/dev/null

debian:
	cd ${BUILD_DIR}; \
	rm -f debian-package/usr/lib/apt/methods/s3; \
	mkdir -p debian-package/usr/lib/apt/methods; \
	cp release/s3-linux-${GOARCH} debian-package/usr/lib/apt/methods/s3; \
	chmod 555 debian-package/usr/lib/apt/methods/s3; \
	sed -i "s/VERSION/${VERSION}/g" debian-package/DEBIAN/control; \
	dpkg-deb --build debian-package release/apt-method-s3.deb; \
	sed -i "s/${VERSION}/VERSION/g" debian-package/DEBIAN/control; \
	cd - >/dev/null; \
	cd ${BUILD_DIR}/release; \
	dpkg-scanpackages . /dev/null | gzip -9c > Packages.gz; \
	cd - >/dev/null

deps:
	cd ${BUILD_DIR}; \
    go get -v ./... ; \
    cd - >/dev/null

test:
	cd ${BUILD_DIR}; \
	go test -v ./... ; \
	cd - >/dev/null

vet:
	-cd ${BUILD_DIR}; \
	go vet ./... > ${VET_REPORT} 2>&1 ; \
	cd - >/dev/null

fmt:
	cd ${BUILD_DIR}; \
	go fmt $$(go list ./... | grep -v /vendor/) ; \
	cd - >/dev/null

clean:
	cd ${BUILD_DIR}; \
	rm -f ${TEST_REPORT}; \
	rm -f ${VET_REPORT}; \
	rm -f release/*; \
	rm -f debian-package/usr/lib/apt/methods/*; \
	mkdir -p release; \
	mkdir -p local; \
	cd - >/dev/null

.PHONY: linux darwin test vet fmt clean