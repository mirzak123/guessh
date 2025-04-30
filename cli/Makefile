BUILD_DIR=../bin
BUILD_NAME=cli

all: build run

build:
	go build -o $(BUILD_DIR)/$(BUILD_NAME) cmd/cli/main.go

run:
	$(BUILD_DIR)/$(BUILD_NAME)
