OUTPUT_DIR=./_output
SOURCE_DIRS = cmd pkg
PACKAGES := go list ./... | grep -v /vendor | grep -v /out

.DEFAULT_GOAL := build

bin=prometurbo
product: fmtcheck vet
	env GOOS=linux GOARCH=amd64 go build -o ${OUTPUT_DIR}/${bin}.linux ./cmd

build: clean
	go build -o ${OUTPUT_DIR}/${bin} ./cmd

docker: clean
	docker build -t turbonomic/prometurbo:6.4dev --build-arg GIT_COMMIT=$(shell git rev-parse --short HEAD) .

test: clean
	@go test -v -race ./pkg/...

.PHONY: clean
clean:
	@: if [ -f ${OUTPUT_DIR} ] then rm -rf ${OUTPUT_DIR} fi

.PHONY: fmtcheck
fmtcheck:
	@gofmt -s -l $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

.PHONY: vet
vet:
	@go vet $(shell $(PACKAGES))

clean:
	@: if [ -f ${OUTPUT_DIR} ] then rm -rf ${OUTPUT_DIR} fi
