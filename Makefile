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
LDFLAGS    := 

CFLAGS  += $(shell pkg-config --cflags libcjson)
LDFLAGS += $(shell pkg-config --libs libcjson)

_SRCS := main.c network.c game_logic.c game_server.c game_types.c json_messages.c client.c util.c hash_table.c room.c
SERVER_SRC := $(addprefix $(SERVER_DIR)/,$(_SRCS))

_TEST_SRCS := test.c game_logic.c hash_table.c util.c
TEST_SRC := $(addprefix $(SERVER_DIR)/,$(_TEST_SRCS))

# ==========================================
# 3. BUILD TARGETS
# ==========================================
.PHONY: all build-all clean build-server build-cli build-ssh run-server run-cli run-ssh test-server debug-server

all: build-all

build-all: build-server build-cli build-ssh

# --- C Server Targets ---
build-server:
	$(CC) $(SERVER_SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/server

run-server: build-server
	$(BUILD_DIR)/server

test-server:
	$(CC) $(TEST_SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/server-test
	$(BUILD_DIR)/server-test

debug-server:
	$(CC) $(ASAN_FLAGS) $(CFLAGS) $(LDFLAGS) $(SERVER_SRC) -o $(BUILD_DIR)/server_debug
	lldb -- $(BUILD_DIR)/server_debug

# --- Go CLI Targets ---
build-cli:
	cd cli && go build -o ../$(BUILD_DIR)/guessh-cli ./cmd/guessh-cli/main.go

run-cli: build-cli
	./$(BUILD_DIR)/guessh-cli

# --- Go SSH Targets ---
build-ssh:
	cd cli && go build -o ../$(BUILD_DIR)/guessh-ssh ./cmd/ssh/main.go

run-ssh: build-ssh
	./$(BUILD_DIR)/guessh-ssh

# ==========================================
# 4. UTILITIES
# ==========================================
clean:
	rm -rf $(BUILD_DIR)/*
	@echo "Cleaned build directory."
