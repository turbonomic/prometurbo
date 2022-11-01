OUTPUT_DIR=build
SOURCE_DIRS = cmd pkg
PACKAGES := go list ./... | grep -v /vendor | grep -v /out
SHELL='/bin/bash'
REMOTE=github.com
USER=turbonomic
PROJECT=prometurbo
bin=prometurbo
DEFAULT_VERSION=latest 

.DEFAULT_GOAL := build


GIT_COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -R)
PROJECT_PATH=$(REMOTE)/$(USER)/$(PROJECT)
VERSION=$(or $(PROMETURBO_VERSION), $(DEFAULT_VERSION))
LDFLAGS='\
 -X "$(PROJECT_PATH)/version.GitCommit=$(GIT_COMMIT)" \
 -X "$(PROJECT_PATH)/version.BuildTime=$(BUILD_TIME)" \
 -X "$(PROJECT_PATH)/version.Version=$(VERSION)"'

LINUX_ARCH=amd64 arm64 ppc64le s390x

$(LINUX_ARCH): clean
	env GOOS=linux GOARCH=$@ go build -ldflags $(LDFLAGS) -o $(OUTPUT_DIR)/linux/$@/$(bin) ./cmd

product: $(LINUX_ARCH)

debug-product: clean
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -gcflags "-N -l" -o ${OUTPUT_DIR}/${bin}_debug ./cmd

build: clean
	go build -ldflags $(LDFLAGS) -o ${bin} ./cmd

debug: clean
	go build -ldflags $(LDFLAGS) -gcflags "-N -l" -o ${bin}.debug ./cmd

docker: product
	cd build; DOCKER_BUILDKIT=1 docker build -t turbonomic/prometurbo --build-arg $(GIT_COMMIT) .

test: clean
	@go test -v -race ./pkg/...

.PHONY: fmtcheck
fmtcheck:
	@gofmt -s -l $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

.PHONY: vet
vet:
	@go vet $(shell $(PACKAGES))

clean:
	@rm -rf ${OUTPUT_DIR}/linux ${bin}
