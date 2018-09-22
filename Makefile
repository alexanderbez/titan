PACKAGES=$(shell go list ./... | grep -v '/vendor/')
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_FLAGS = -tags netgo -ldflags "-X github.com/alexanderbez/titan/version.GitCommit=${COMMIT_HASH}"

all: tools deps test install

build:
	@echo "--> Building..."
	go build $(BUILD_FLAGS) .

install:
	@echo "--> Installing..."
	go install $(BUILD_FLAGS) .

clean:
	@echo "--> Cleaning directory..."
	@rm -rf ./vendor

deps: clean
	@echo "--> Fetching vendor dependencies..."
	@dep ensure -v

tools: get-tools
	@echo "--> Fetching tools..."

############################
### Tools / Dependencies ###
############################

DEP = github.com/golang/dep/cmd/dep
GOLINT = github.com/tendermint/lint/golint
GOMETALINTER = gopkg.in/alecthomas/gometalinter.v2
UNCONVERT = github.com/mdempsky/unconvert
INEFFASSIGN = github.com/gordonklaus/ineffassign
MISSPELL = github.com/client9/misspell/cmd/misspell
ERRCHECK = github.com/kisielk/errcheck
UNPARAM = mvdan.cc/unparam

DEP_CHECK := $(shell command -v dep 2> /dev/null)
GOLINT_CHECK := $(shell command -v golint 2> /dev/null)
GOMETALINTER_CHECK := $(shell command -v gometalinter.v2 2> /dev/null)
UNCONVERT_CHECK := $(shell command -v unconvert 2> /dev/null)
INEFFASSIGN_CHECK := $(shell command -v ineffassign 2> /dev/null)
MISSPELL_CHECK := $(shell command -v misspell 2> /dev/null)
ERRCHECK_CHECK := $(shell command -v errcheck 2> /dev/null)
UNPARAM_CHECK := $(shell command -v unparam 2> /dev/null)

get-tools:
ifdef DEP_CHECK
	@echo "Dep is already installed."
else
	@echo "--> Installing dep"
	go get -v $(DEP)
endif
ifdef GOLINT_CHECK
	@echo "Golint is already installed"
else
	@echo "--> Installing golint"
	go get -v $(GOLINT)
endif
ifdef GOMETALINTER_CHECK
	@echo "Gometalinter.v2 is already installed."
else
	@echo "--> Installing gometalinter.v2"
	go get -v $(GOMETALINTER)
endif
ifdef UNCONVERT_CHECK
	@echo "Unconvert is already installed."
else
	@echo "--> Installing unconvert"
	go get -v $(UNCONVERT)
endif
ifdef INEFFASSIGN_CHECK
	@echo "Ineffassign is already installed."
else
	@echo "--> Installing ineffassign"
	go get -v $(INEFFASSIGN)
endif
ifdef MISSPELL_CHECK
	@echo "misspell is already installed."
else
	@echo "--> Installing misspell"
	go get -v $(MISSPELL)
endif
ifdef ERRCHECK_CHECK
	@echo "errcheck is already installed."
else
	@echo "--> Installing errcheck"
	go get -v $(ERRCHECK)
endif
ifdef UNPARAM_CHECK
	@echo "unparam is already installed."
else
	@echo "--> Installing unparam"
	go get -v $(UNPARAM)
endif

#######################
### Testing / Misc. ###
#######################

test: test-unit test-lint

test-unit:
	@go test -v --vet=off $(PACKAGES)

test-race:
	@go test -v -race --vet=off $(PACKAGES)

test-lint:
	@echo "--> Running gometalinter..."
	@gometalinter.v2 --config=gometalinter.json --exclude=vendor ./...

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/ethermint"
	godoc -http=:6060

format:
	@echo "--> Formatting Golang files..."
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -w -s
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs misspell -w

.PHONY: build install clean deps tools get-tools test test-unit test-race \
test-lint godocs format
