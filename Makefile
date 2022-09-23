
COMMIT_HASH ?= "$(shell git describe --long --dirty --always --match "" || true)"
CLEAN_COMMIT ?= "$(shell git describe --long --always --match "" || true)"
COMMIT_TIME ?= "$(shell git show -s --format=%ct $(CLEAN_COMMIT) || true)"

export GOPROXY=https://goproxy.io,direct

LD_FLAGS ?= \
-X github.com/redesblock/mop/cmd/version.commitHash="$(COMMIT_HASH)" \
-X github.com/redesblock/mop/cmd/version.commitTime="$(COMMIT_TIME)"

all:
	@go mod tidy
	@go fmt ./...
	@go build -tags=jsoniter -trimpath -ldflags "$(LD_FLAGS)" -o ./bin/ ./...

generate:
	@go generate ./...

linux:
	@go mod tidy
	@go fmt ./...
	@GOOS=linux GOARCH=amd64 go build -tags=jsoniter $(LD_FLAGS) -o ./bin/ ./...
###############################################################################
###                                 Docker                                 ###
###############################################################################
docker:
	docker build -t redesblock/mop .