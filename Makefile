COMMIT_HASH ?= "$(shell git describe --long --dirty --always --match "" || true)"
CLEAN_COMMIT ?= "$(shell git describe --long --always --match "" || true)"
COMMIT_TIME ?= "$(shell git show -s --format=%ct $(CLEAN_COMMIT) || true)"
API_VERSION ?= "$(shell grep '^  version:' api/API.yaml | awk '{print $$2}')"
DEBUG_API_VERSION ?= "$(shell grep '^  version:' api/DebugAPI.yaml | awk '{print $$2}')"

LDFLAGS ?= -s -w \
-X github.com/redesblock/mop.commitHash="$(COMMIT_HASH)" \
-X github.com/redesblock/mop.commitTime="$(COMMIT_TIME)" \
-X github.com/redesblock/mop/pkg/api.Version="$(API_VERSION)" \
-X github.com/redesblock/mop/pkg/debugapi.Version="$(DEBUG_API_VERSION)"

.PHONY: all
all: binary

.PHONY: binary
binary: CGO_ENABLED=0
binary:
	go fmt ./...
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/mop ./cmd/mop

.PHONY: release
release: CGO_ENABLED=0
release:
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o release/mop-windows-amd64.exe ./cmd/mop
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o release/mop-linux-amd64 ./cmd/mop
	GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o release/mop-darwin-amd64 ./cmd/mop

.PHONY: protobuftools
protobuftools:
	which protoc || ( echo "install protoc for your system from https://github.com/protocolbuffers/protobuf/releases" && exit 1)
	which protoc-gen-gogofaster || ( cd /tmp && GO111MODULE=on go get -u github.com/gogo/protobuf/protoc-gen-gogofaster@v1.3.1 )

.PHONY: protobuf
protobuf: GOFLAGS=-mod=mod
protobuf: protobuftools
	go generate -run protoc ./...

.PHONY: clean
clean:
	go clean
	rm -rf dist/

FOLDER=$(shell pwd)
.PHONY: format
format:
	gofumpt -l -w $(FOLDER)
	gci -w -local $(go list -m) `find $(FOLDER) -type f \! -name "*.pb.go" -name "*.go" \! -path \*/\.git/\* -exec echo {} \;`

FORCE:
