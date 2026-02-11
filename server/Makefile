BREW_PREFIX := /opt/homebrew

CFLAGS := -std=c17 -Wall -Wextra -pedantic -I$(BREW_PREFIX)/include
LDFLAGS := -L$(BREW_PREFIX)/lib -lcjson
ASAN_FLAGS := -g -O0 -fsanitize=address -fno-omit-frame-pointer

BUILD_DIR := ../bin
BUILD_NAME := server
BUILD_TEST_NAME := server-test
SRC := main.c network.c game_logic.c game_server.c game_types.c json_messages.c client.c util.c hash_table.c room.c
TEST_SRC := test.c game_logic.c hash_table.c

CC := gcc

all: build run

build:
	$(CC) $(SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BUILD_NAME)

run:
	$(BUILD_DIR)/$(BUILD_NAME)

test-build:
	$(CC) -g -O0 $(TEST_SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BUILD_TEST_NAME)

test:
	# Note: test.c likely needs cJSON flags too if it uses the library
	$(CC) test.c game_logic.c util.c $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BUILD_TEST_NAME) && $(BUILD_DIR)/$(BUILD_TEST_NAME)

clean:
	rm -rf $(BUILD_DIR)/*

# Debug targets
debug-build:
	$(CC) -g -O0 $(CFLAGS) $(LDFLAGS) $(SRC) -o $(BUILD_DIR)/$(BUILD_NAME)_debug

debug: debug-build
	lldb -- $(BUILD_DIR)/$(BUILD_NAME)_debug

test-debug: test-build
	lldb -- $(BUILD_DIR)/$(BUILD_TEST_NAME)
