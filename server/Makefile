BREW_PREFIX := /opt/homebrew
CFLAGS := -std=c17 -Wall -Wextra -pedantic -I$(BREW_PREFIX)/include
LDFLAGS := -L$(BREW_PREFIX)/lib -lcjson

BUILD_DIR := ../bin
BUILD_NAME := server
BUILD_TEST_NAME := server-test
SRC := main.c network.c game_logic.c game_server.c game_types.c json_messages.c client.c util.c
TEST_SRC := test.c game_logic.c hash_table.c

CC := gcc

all: build run

build:
	$(CC) $(SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BUILD_NAME)

debug:
	$(CC) -g -O0 $(CFLAGS) $(LDFLAGS) $(SRC) -o $(BUILD_DIR)/$(BUILD_NAME)_debug

run:
	$(BUILD_DIR)/$(BUILD_NAME)

debug-run: debug
	lldb -- $(BUILD_DIR)/$(BUILD_NAME)_debug

test-build:
	$(CC) -g -O0 $(TEST_SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BUILD_TEST_NAME)

test: test-build
	$(BUILD_DIR)/$(BUILD_TEST_NAME)

test-debug:
	lldb -- $(BUILD_DIR)/$(BUILD_TEST_NAME)

clean:
	rm -rf $(BUILD_DIR)/*
