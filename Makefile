BUILD_ARGS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
EASY_RELEASE_BINARY=easy-release
PR_LINT_BINARY=pr-lint
DIST_FOLDER=dist

.PHONY: all
all: clean setup test build

.PHONY: clean
clean:
	rm -rf $(DIST_FOLDER)

.PHONY: setup
setup: 
	go mod tidy

.PHONY: test
test:
	go test ./... --cover

.PHONY: build
build:
	$(BUILD_ARGS) go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(DIST_FOLDER)/$(EASY_RELEASE_BINARY)/$(EASY_RELEASE_BINARY) cmd/easy-release/main.go
	$(BUILD_ARGS) go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(DIST_FOLDER)/$(PR_LINT_BINARY)/$(PR_LINT_BINARY) cmd/pr-lint/main.go

.PHONY: run
run:
	go run cmd/main.go
