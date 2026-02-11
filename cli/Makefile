BUILD_DIR = bin
BUILD_NAME = guessh-cli

all: build run

build:
	go build -o $(BUILD_DIR)/$(BUILD_NAME) cmd/guessh-cli/main.go

run:
	$(BUILD_DIR)/$(BUILD_NAME)
