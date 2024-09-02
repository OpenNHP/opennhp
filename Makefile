export GO111MODULE := on

all: generate-version-and-build

# Repo settings
GIT_REPO = github.com/OpenNHP/opennhp

# Version and build settings
MAKEFLAGS += --no-print-directory
OS_NAME = $(shell uname -s | tr A-Z a-z)

# Version number auto increment 
TIMESTAMP=$(shell date +%y%m%d%H%M%S)
VERSION = $(shell cat version/VERSION).$(TIMESTAMP)
# Other version settings
COMMIT_ID = $(shell git show -s --format=%H)
COMMIT_TIME = $(shell git show -s --format=%cd --date=format:'%Y-%m-%d %H:%M:%S')
BUILD_TIME = $(shell date "+%Y-%m-%d %H:%M:%S")
# Built Package File Name
PACKAGE_FILE = opennhp-$(VERSION).tar.gz
# Go build flags
LD_FLAGS = "-s -w -X '${GIT_REPO}/version.Version=${VERSION}' -X '${GIT_REPO}/version.CommitId=${COMMIT_ID}' -X '${GIT_REPO}/version.CommitTime=${COMMIT_TIME}' -X '${GIT_REPO}/version.BuildTime=${BUILD_TIME}'"

# Color definition
COLOUR_GREEN=\033[0;32m
COLOUR_RED=\033[0;31m
COLOUR_BLUE=\033[0;34m
END_COLOUR=\033[0m

# Plugins
NHP_PLUGINS = server/plugins

generate-version-and-build:
	@echo "$(COLOUR_BLUE)[OpenNHP] Start building... $(END_COLOUR)"
	@echo "$(COLOUR_BLUE)Version: ${VERSION} $(END_COLOUR)"
	@echo "$(COLOUR_BLUE)Commit id: ${COMMIT_ID} $(END_COLOUR)"
	@echo "$(COLOUR_BLUE)Commit time: ${COMMIT_TIME} $(END_COLOUR)"
	@echo "$(COLOUR_BLUE)Build time: ${BUILD_TIME} $(END_COLOUR)"
	@$(MAKE) init
	@$(MAKE) agentd
	@$(MAKE) acd
	@$(MAKE) serverd
	@$(MAKE) agentsdk
	@$(MAKE) plugins
	@$(MAKE) archive
	@echo "$(COLOUR_GREEN)[OpenNHP] Build for platform ${OS_NAME} successfully done!$(END_COLOUR)"

init:
	git clean -df release
	go mod tidy

agentd:
	go build -trimpath -ldflags ${LD_FLAGS} -v -o ./release/nhp-agent/nhp-agentd ./agent/main/main.go
	cp ./agent/main/etc/*.toml ./release/nhp-agent/etc/

acd:
	go build -trimpath -ldflags ${LD_FLAGS} -v -o ./release/nhp-ac/nhp-acd ./ac/main/main.go
	cp ./ac/main/etc/*.toml ./release/nhp-ac/etc/

serverd:
	go build -trimpath -ldflags ${LD_FLAGS} -v -o ./release/nhp-server/nhp-serverd ./server/main/main.go
	mkdir -p ./release/nhp-server/etc
	cp ./server/main/etc/*.toml ./release/nhp-server/etc/

agentsdk:
ifeq ($(OS_NAME), linux)
	go build -a -trimpath -buildmode=c-shared -ldflags ${LD_FLAGS} -v -o ./release/nhp-agent/nhp-agent.so ./agent/main/main.go ./agent/main/export.go
	gcc ./agent/sdkdemo/nhp-agent-demo.c -I ./release/nhp-agent -l:nhp-agent.so -L./release/nhp-agent -Wl,-rpath=. -o ./release/nhp-agent/nhp-agent-demo
endif

plugins:
	@if test -d $(NHP_PLUGINS); then $(MAKE) -C $(NHP_PLUGINS); fi

archive:
	@echo "$(COLOUR_BLUE)[opennhp] Start archiving... $(END_COLOUR)"
	@cd release && mkdir -p archive && tar -czvf ./archive/$(PACKAGE_FILE) nhp-agent nhp-ac nhp-server
	@echo "$(COLOUR_GREEN)[opennhp] Package ${PACKAGE_FILE} archived!$(END_COLOUR)"

.PHONY: all generate-version-and-build init agentsdk plugins archive
