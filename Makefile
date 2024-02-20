MAIN_PACKAGE_PATH := ./cmd/crawler
MAIN_BUILD_PATH := ./build
BINARY_NAME := crawler

## build: build the application
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	go build -o=${MAIN_BUILD_PATH}/${BINARY_NAME} ${MAIN_PACKAGE_PATH}.go
