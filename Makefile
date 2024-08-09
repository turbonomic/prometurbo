OUTPUT_DIR=build
SOURCE_DIRS = cmd pkg
PACKAGES := go list ./... | grep -v /vendor | grep -v /out
SHELL='/bin/bash'
REMOTE=github.ibm.com
USER=turbonomic
PROJECT=prometurbo
PROMETURBO_VERSION=8.13.3-SNAPSHOT
bin=prometurbo
DEFAULT_VERSION=latest
REMOTE_URL=$(shell git config --get remote.origin.url)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
REVISION=$(shell git show -s --format=%cd --date=format:'%Y%m%d%H%M%S000')

YAMLLINT_VERSION := 1.32
PROMETURBO_OPERATOR_CRD_OPERATOR_HUB := deploy/prometurbo-operator/deploy/crds/charts.helm.k8s.io_prometurbos_crd.yaml

.DEFAULT_GOAL := build

GIT_COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -R)
BUILD_TIMESTAMP=$(shell date +'%Y%m%d%H%M%S000')
PROJECT_PATH=$(REMOTE)/$(USER)/$(PROJECT)
VERSION=$(or $(PROMETURBO_VERSION), $(DEFAULT_VERSION))
LDFLAGS='\
 -X "$(PROJECT_PATH)/version.GitCommit=$(GIT_COMMIT)" \
 -X "$(PROJECT_PATH)/version.BuildTime=$(BUILD_TIME)" \
 -X "$(PROJECT_PATH)/version.Version=$(VERSION)"'

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

LINUX_ARCH=amd64 arm64 ppc64le s390x

$(LINUX_ARCH): clean
	env GOOS=linux GOARCH=$@ go build -ldflags $(LDFLAGS) -o $(OUTPUT_DIR)/linux/$@/$(bin) ./cmd

product: $(LINUX_ARCH)

debug-product: clean
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -gcflags "-N -l" -o ${OUTPUT_DIR}/${bin}_debug ./cmd

build: clean
	go build -ldflags $(LDFLAGS) -o ${bin} ./cmd

buildInfo:
			$(shell test -f git.properties && rm -rf git.properties)
			@echo 'turbo-version.remote.origin.url=$(REMOTE_URL)' >> git.properties
			@echo 'turbo-version.commit.id=$(GIT_COMMIT)' >> git.properties
			@echo 'turbo-version.branch=$(BRANCH)' >> git.properties
			@echo 'turbo-version.branch.version=$(VERSION)' >> git.properties
			@echo 'turbo-version.commit.time=$(REVISION)' >> git.properties
			@echo 'turbo-version.build.time=$(BUILD_TIMESTAMP)' >> git.properties

debug: clean
	go build -ldflags $(LDFLAGS) -gcflags "-N -l" -o ${bin}.debug ./cmd

docker: product
	cd build; DOCKER_BUILDKIT=1 docker build -t turbonomic/prometurbo --build-arg $(GIT_COMMIT) --load .


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

.PHONY: python
PYTHON = $(LOCALBIN)/python3
python: $(PYTHON)  ## Install Python locally if necessary. Darwin OS is specific to mac users if running locally
$(PYTHON):
	@if ! command -v python3 >/dev/null 2>&1; then \
		mkdir -p $(LOCALBIN); \
		if [ `uname -s` = "Darwin" ]; then \
			brew install python@3; \
		else \
			sudo apt update && sudo apt install python3; \
		fi \
	fi
	mkdir -p $(dir $(PYTHON))
	ln -sf `command -v python3` $(PYTHON)

.PHONY: yamllint
yamllint: python
	$(PYTHON) -m pip install yamllint==$(YAMLLINT_VERSION)


yaml-lint-check: yamllint
	$(PYTHON) -m yamllint -d '{extends: default, rules: {line-length: {max: 180, level: warning}, indentation: {indent-sequences: whatever}}}' $(PROMETURBO_OPERATOR_CRD_OPERATOR_HUB)
	rm -rf ./bin

PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
REPO_NAME ?= icr.io/cpopen/turbonomic
.PHONY: multi-archs
multi-archs:
	env GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o ${bin} ./cmd
.PHONY: docker-buildx
docker-buildx:
	docker buildx create --name prometurbo-builder
	- docker buildx use prometurbo-builder
	- docker buildx build --platform=$(PLATFORMS) --label "git-commit=$(GIT_COMMIT)" --label "git-version=$(VERSION)" --provenance=false --push --tag $(REPO_NAME)/$(PROJECT):$(VERSION) -f build/Dockerfile.multi-archs --build-arg VERSION=$(VERSION) .
	docker buildx rm prometurbo-builder

check-upstream-dependencies:
	./scripts/travis/check_upstream_dependencies.sh

# Check, Build, and Test:
#   - Runs code checks including formatting, linting, and vetting.
#   - Builds the project to ensure compilation success.
#   - Runs tests to verify code functionality and correctness.
.PHONY: check-build-test
check-build-test:
	@echo "Checking, building, and testing..."
	make fmtcheck
	make vet
	make yaml-lint-check
	make product
	make test
	@echo "Check, build, and test completed."

# Pull Request Build Target
.PHONY: pull-request-build
pull-request-build: check-upstream-dependencies check-build-test
