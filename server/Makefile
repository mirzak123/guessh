BUILD_DIR := ../bin
BUILD_NAME := server
BUILD_TEST_NAME := server-test
SRC := main.c network.c game_logic.c game_server.c

all: build run

build:
	gcc $(SRC) -o $(BUILD_DIR)/$(BUILD_NAME)

debug:
	gcc -g -O0 -Wall -Wextra -pedantic $(SRC) -o $(BUILD_DIR)/$(BUILD_NAME)_debug

run:
	$(BUILD_DIR)/$(BUILD_NAME)

run-debug:
	$(BUILD_DIR)/$(BUILD_NAME)_debug

test:
	gcc test.c game_logic.c -o $(BUILD_DIR)/$(BUILD_TEST_NAME) && $(BUILD_DIR)/$(BUILD_TEST_NAME)
