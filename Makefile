# ==========================================
# 1. ENVIRONMENT & CONFIGURATION
# ==========================================
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

BUILD_DIR := bin
$(shell mkdir -p $(BUILD_DIR))

# ==========================================
# 2. GUESSH-GAMED CONFIGURATION
# ==========================================
CC         := gcc
CFLAGS     := -std=gnu17 -Wall -Wextra -pedantic
LDFLAGS    := -g -O0

CFLAGS  += $(shell pkg-config --cflags libcjson)
LDFLAGS += $(shell pkg-config --libs libcjson)
ASAN_FLAGS := -g -O0 -fsanitize=address -fno-omit-frame-pointer

SRC_DIR := server/src
TEST_DIR := server/test

ALL_SRC := $(wildcard $(SRC_DIR)/*.c)
CORE_SRC := $(filter-out $(SRC_DIR)/main.c, $(ALL_SRC))
TEST_SRC := $(wildcard $(TEST_DIR)/*.c)
TEST_BUILD_SRC := $(CORE_SRC) $(TEST_SRC)

# ==========================================
# 3. BUILD TARGETS
# ==========================================
.PHONY: all build-all clean build-gamed build-tui build-cli build-sshd run-gamed run-tui run-cli run-sshd test-gamed debug-gamed

all: build-all

build-all: build-gamed build-tui build-cli build-sshd

# --- guessh-gamed ---
build-gamed:
	$(CC) $(ALL_SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/guessh-gamed

run-gamed: build-gamed
	$(BUILD_DIR)/guessh-gamed

test-gamed:
	$(CC) $(TEST_BUILD_SRC) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/guessh-gamed-test
	$(BUILD_DIR)/guessh-gamed-test

debug-gamed:
	$(CC) $(ALL_SRC) $(ASAN_FLAGS) $(CFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/guessh-gamed-dbg
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
