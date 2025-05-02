BUILD_DIR=../bin
BUILD_NAME=server
BUILD_TEST_NAME=server-test

all: build run

build:
	gcc main.c -o $(BUILD_DIR)/$(BUILD_NAME)

run:
	$(BUILD_DIR)/$(BUILD_NAME)

test:
	gcc test.c game.c -o $(BUILD_DIR)/$(BUILD_TEST_NAME) && $(BUILD_DIR)/$(BUILD_TEST_NAME)
