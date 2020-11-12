SHELL=/bin/bash
BIN=9ofm
GOOS=linux
GOARCH=amd64
BUILD_PATH=bin/$(GOOS)/$(GOARCH)/$(BIN)


all: clean vend build

## For development

run: build
	$(BUILD_PATH)

debug: build
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec $(BUILD_PATH)

build:
	go build -ldflags "-X main.sha1ver=$$(git rev-parse HEAD) -X main.buildTime=$$(date +'%Y-%m-%dT%T')" -gcflags="all=-N -l" -o $(BUILD_PATH)

test:
	./scripts/test-coverage.sh

test-race:
	go test -cover -v -race ./...

test-static-analysis:
	grep -R 'const allowTestDataCapture = false' runtime/ui/viewmodel
	go vet ./...
	@! gofmt -s -l . 2>&1 | grep -vE '^\.git/' | grep -vE '^\.cache/'
	golangci-lint run

#  add missing and remove unused modules
vend: 
	go mod vendor
	go mod tidy
	
clean:
	rm -rf dist
	go clean
