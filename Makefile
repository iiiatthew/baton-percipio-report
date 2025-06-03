GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
BUILD_DIR = dist/${GOOS}_${GOARCH}
PROJECT_NAME = baton-percipio-report

ifeq ($(GOOS),windows)
OUTPUT_PATH = ${BUILD_DIR}/${PROJECT_NAME}.exe
else
OUTPUT_PATH = ${BUILD_DIR}/${PROJECT_NAME}
endif

.PHONY: build
build:
	go build -o ${OUTPUT_PATH} ./cmd/${PROJECT_NAME}

# Build for Linux
.PHONY: build-linux:
	GOOS=linux GOARCH=amd64 go build -mod=mod -o ${OUTPUT_PATH} ./cmd/${PROJECT_NAME}
	GOOS=linux GOARCH=arm64 go build -mod=mod -o ${OUTPUT_PATH} ./cmd/${PROJECT_NAME}

# Build for macOS
.PHONY: build-macos:
	GOOS=darwin GOARCH=amd64 go build -mod=mod -o ${OUTPUT_PATH} ./cmd/${PROJECT_NAME}
	GOOS=darwin GOARCH=arm64 go build -mod=mod -o ${OUTPUT_PATH} ./cmd/${PROJECT_NAME}

# Build for Windows
.PHONY: build-windows:
	GOOS=windows GOARCH=amd64 go build -mod=mod -o ${OUTPUT_PATH} ./cmd/${PROJECT_NAME}

# Build for all platforms
.PHONY: build-all: build-linux build-macos build-windows

.PHONY: update-deps
update-deps:
	go get -d -u ./...
	go mod tidy -v
	go mod vendor

.PHONY: add-dep
add-dep:
	go mod tidy -v
	go mod vendor

.PHONY: lint
lint:
	golangci-lint run
