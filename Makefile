
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
COMMIT=$(shell git rev-parse --short HEAD)

# don't override user values
ifeq (,$(REVISION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(REVISION))
    REVISION := $(BRANCH)-$(COMMIT)
  endif
endif

export GOPROXY=https://goproxy.io,direct

LD_FLAGS:=-ldflags "-X github.com/redesblock/cmd.commit=$(REVISION)"

all:
	@go mod tidy
	@go fmt ./...
	@go build -tags=jsoniter $(LD_FLAGS) -o ./bin/ ./...

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
	docker build -t redesblock/hop .