# ==========================================
# 1. ENVIRONMENT & CONFIGURATION
# ==========================================
# Load .env variables if the file exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Shared Variables
BUILD_DIR := bin
$(shell mkdir -p $(BUILD_DIR))

# ==========================================
# 2. C SERVER CONFIGURATION
# ==========================================
SERVER_DIR := server
CC         := gcc
CFLAGS     := -std=gnu17 -Wall -Wextra -pedantic
LDFLAGS    := -g -O0

CFLAGS  += $(shell pkg-config --cflags libcjson)
LDFLAGS += $(shell pkg-config --libs libcjson)
ASAN_FLAGS := -g -O0 -fsanitize=address -fno-omit-frame-pointer

_SRCS := main.c network.c game_logic.c game_server.c game_types.c json_messages.c client.c util.c hash_table.c room.c timer.c
SERVER_SRC := $(addprefix $(SERVER_DIR)/,$(_SRCS))

_TEST_SRCS := test.c hash_table.c util.c timer.c
TEST_SRC := $(addprefix $(SERVER_DIR)/,$(_TEST_SRCS))

# ==========================================
# 3. BUILD TARGETS
# ==========================================
.PHONY: all build-all clean build-gamed build-tui build-cli build-sshd run-gamed run-tui run-cli run-sshd test-gamed debug-gamed

all: build-all

build-all: build-gamed build-tui build-cli build-sshd

# --- guessh-gamed ---
build-gamed:
	$(CC) $(SERVER_SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/server

run-gamed: build-gamed
	$(BUILD_DIR)/server

test-gamed:
	$(CC) $(TEST_SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/guessh-gamed-test
	$(BUILD_DIR)/guessh-gamed-test

debug-gamed:
	$(CC) $(SERVER_SRC) $(ASAN_FLAGS) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/guessh-gamed-dbg
	lldb -- $(BUILD_DIR)/guessh-gamed-dbg

# --- guessh-sshd ---
build-sshd:
	cd cli && go build -o ../$(BUILD_DIR)/guessh-sshd ./cmd/guessh-ssh/main.go

run-sshd: build-sshd
	./$(BUILD_DIR)/guessh-ssh

# --- guessh-tui ---
build-tui:
	cd cli && go build -o ../$(BUILD_DIR)/guessh-tui ./cmd/guessh-tui/main.go

run-tui: build-tui
	./$(BUILD_DIR)/guessh-tui

# --- guessh-cli ---
build-cli:
	cd cli && go build -o ../$(BUILD_DIR)/guessh-cli ./cmd/guessh-cli/main.go

run-cli: build-cli
	./$(BUILD_DIR)/guessh-cli

# ==========================================
# 4. UTILITIES
# ==========================================
clean:
	rm -rf $(BUILD_DIR)/*
	@echo "Cleaned build directory."
