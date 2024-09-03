export GO111MODULE := on

all: build

# Repo settings
GIT_REPO = github.com/OpenNHP/opennhp

# Version and build settings
MAKEFLAGS += --no-print-directory
COMMIT_ID = $(shell git show -s --format=%H)
COMMIT_TIME = $(shell git show -s --format=%cd --date=format:'%Y-%m-%d %H:%M:%S')
BUILD_TIME = $(shell date "+%Y-%m-%d %H:%M:%S")
OS_NAME = $(shell uname -s | tr A-Z a-z)

# Server plugins
SERVER_PLUGINS = $(sort $(dir $(wildcard */)))

# Color definition
COLOUR_GREEN=\033[0;32m
COLOUR_RED=\033[0;31m
COLOUR_BLUE=\033[0;34m
END_COLOUR=\033[0m

# build server plugins
build:
	@echo "$(COLOUR_BLUE)[NHP-Server Plugins] Start building... $(END_COLOUR)"
	$(foreach plugin, $(SERVER_PLUGINS), $(MAKE) -C $(plugin);)
	@echo "$(COLOUR_GREEN)[NHP-Server Plugins] ${SERVER_PLUGINS} for platform ${OS_NAME} successfully built!$(END_COLOUR)"

.PHONY: all build