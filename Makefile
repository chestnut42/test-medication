export GOBIN := $(PWD)/bin
export PATH := $(GOBIN):$(PATH)

GOLANGLINT_VERSION := 2.1.16

.PHONY: default
default: test


./bin:
	mkdir -p ./bin

# Tools
./bin/goimports: | ./bin
	go install -modfile tools/go.mod golang.org/x/tools/cmd/goimports

./bin/golangci-lint: | ./bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v$(GOLANGLINT_VERSION)

./bin/gowrap: | ./bin
	go install -modfile tools/go.mod github.com/hexdigest/gowrap/cmd/gowrap

./bin/minimock: | ./bin
	go install -modfile tools/go.mod github.com/gojuno/minimock/v3/cmd/minimock


.PHONY: clean
clean:
	rm -rf ./bin

.PHONY: generate
generate: ./bin/gowrap ./bin/minimock ./bin/goimports
	go generate ./...
	goimports -w -local github.com/farawaygg .

.PHONY: lint
lint: ./bin/golangci-lint
	golangci-lint run -v ./...

.PHONY: test
test:
	go test -race $(GOFLAGS) -v ./... -count 1

.PHONY: build
build:
	GOGC=off go build -v -o ./bin/medication ./cmd/medication

.PHONY: tidy
tidy:
	go mod tidy
	cd tools && go mod tidy
