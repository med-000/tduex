BIN_NAME := tduex
CMD_PATH := ./cmd/tduex
PREFIX ?= /usr/local
USER_BIN ?= $(HOME)/.local/bin

.PHONY: build install install-user uninstall

build:
	go build -o $(BIN_NAME) $(CMD_PATH)

install:
	go build -o $(PREFIX)/bin/$(BIN_NAME) $(CMD_PATH)

install-user:
	mkdir -p $(USER_BIN)
	go build -o $(USER_BIN)/$(BIN_NAME) $(CMD_PATH)

uninstall:
	rm -f $(PREFIX)/bin/$(BIN_NAME)
