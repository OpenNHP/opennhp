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
	@$(MAKE) kgc
	@$(MAKE) agentsdk
	@$(MAKE) devicesdk
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

kgc:
	@echo "$(COLOUR_BLUE)[KGC] Building KGC module... $(END_COLOUR)"
	mkdir -p ./release/kgc/etc
	@cd kgc/main && go build -trimpath -ldflags ${LD_FLAGS} -v -o ../../release/kgc/kgc ./main.go
	cp ./kgc/main/etc/*.toml ./release/kgc/etc/ 2>/dev/null || true
	@echo "$(COLOUR_GREEN)[KGC] Build completed!$(END_COLOUR)"


agentsdk:
ifeq ($(OS_NAME), linux)
	go build -a -trimpath -buildmode=c-shared -ldflags ${LD_FLAGS} -v -o ./release/nhp-agent/nhp-agent.so ./agent/main/main.go ./agent/main/export.go
	gcc ./agent/sdkdemo/nhp-agent-demo.c -I ./release/nhp-agent -l:nhp-agent.so -L./release/nhp-agent -Wl,-rpath=. -o ./release/nhp-agent/nhp-agent-demo
endif

devicesdk:
ifeq ($(OS_NAME), linux)
	go build -a -trimpath -buildmode=c-shared -ldflags ${LD_FLAGS} -v -o ./release/nhp-device/nhpdevice.so ./core/main/main.go ./core/main/nhpdevice.go
#	gcc ./core/sdkdemo/nhp-device-demo.c -I ./release/nhp-device -I ./core/main -l:nhpdevice.so -L./release/nhp-device -Wl,-rpath=. -o ./release/nhp-device/nhp-device-demo
endif

plugins:
	@if test -d $(NHP_PLUGINS); then $(MAKE) -C $(NHP_PLUGINS); fi

test:
	@echo "[OpenNHP] Runing Tests for the Output Binaries ..."
	@echo "$(COLOUR_GREEN)[OpenNHP] All Tests Are Done!$(END_COLOUR)"

archive:
	@echo "$(COLOUR_BLUE)[OpenNHP] Start archiving... $(END_COLOUR)"
	@cd release && mkdir -p archive && tar -czvf ./archive/$(PACKAGE_FILE) nhp-agent nhp-ac nhp-server
	@echo "$(COLOUR_GREEN)[OpenNHP] Package ${PACKAGE_FILE} archived!$(END_COLOUR)"

.PHONY: all generate-version-and-build init agentsdk devicesdk plugins test archive 
