export GO111MODULE := on

all: generate-version-and-build

# Repo settings
GOMODULE = github.com/OpenNHP/opennhp/nhp

# Version and build settings
MAKEFLAGS += --no-print-directory
OS_NAME = $(shell uname -s | tr A-Z a-z)

# Version number auto increment 
TIMESTAMP=$(shell date +%y%m%d%H%M%S)
VERSION = $(shell cat nhp/version/VERSION).$(TIMESTAMP)
# Other version settings
COMMIT_ID = $(shell git show -s --format=%H)
COMMIT_TIME = $(shell git show -s --format=%cd --date=format:'%Y-%m-%d %H:%M:%S')
BUILD_TIME = $(shell date "+%Y-%m-%d %H:%M:%S")
# Built Package File Name
PACKAGE_FILE = opennhp-$(VERSION).tar.gz
# Go build flags
LD_FLAGS = "-s -w -X '${GOMODULE}/version.Version=${VERSION}' -X '${GOMODULE}/version.CommitId=${COMMIT_ID}' -X '${GOMODULE}/version.CommitTime=${COMMIT_TIME}' -X '${GOMODULE}/version.BuildTime=${BUILD_TIME}'"

# Color definition
COLOUR_GREEN=\033[0;32m
COLOUR_RED=\033[0;31m
COLOUR_BLUE=\033[0;34m
END_COLOUR=\033[0m

# Plugins
NHP_SERVER_PLUGINS = ./endpoints/server/plugins

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
	@$(MAKE) de
	@$(MAKE) agentsdk
	@$(MAKE) devicesdk
	@$(MAKE) plugins
	@$(MAKE) archive
	@echo "$(COLOUR_GREEN)[OpenNHP] Build for platform ${OS_NAME} successfully done!$(END_COLOUR)"

init:
	@echo "$(COLOUR_BLUE)[OpenNHP] Initializing... $(END_COLOUR)"
	git clean -df release
	cd nhp && go mod tidy
	cd endpoints && go mod tidy

agentd:
	@echo "$(COLOUR_BLUE)[OpenNHP] Building nhp-agent... $(END_COLOUR)"
	cd endpoints && \
	go build -trimpath -ldflags ${LD_FLAGS} -v -o ../release/nhp-agent/nhp-agentd ./agent/main/main.go && \
	cp ./agent/main/etc/*.toml ../release/nhp-agent/etc/

acd:
	@echo "$(COLOUR_BLUE)[OpenNHP] Building nhp-ac... $(END_COLOUR)"
	cd endpoints && \
	go build -trimpath -ldflags ${LD_FLAGS} -v -o ../release/nhp-ac/nhp-acd ./ac/main/main.go && \
	cp ./ac/main/etc/*.toml ../release/nhp-ac/etc/

serverd:
	@echo "$(COLOUR_BLUE)[OpenNHP] Building nhp-server... $(END_COLOUR)"
	cd endpoints && \
	go build -trimpath -ldflags ${LD_FLAGS} -v -o ../release/nhp-server/nhp-serverd ./server/main/main.go && \
	mkdir -p ../release/nhp-server/etc; \
	cp ./server/main/etc/*.toml ../release/nhp-server/etc/

de:
	@echo "$(COLOUR_BLUE)[OpenNHP] Building nhp-de... $(END_COLOUR)"
	cd endpoints && \
	go build -trimpath -ldflags ${LD_FLAGS} -v -o ../release/nhp-de/nhp-de ./de/main/main.go && \
	mkdir -p ../release/nhp-de/etc; \
	cp ./de/main/etc/*.toml ../release/nhp-de/etc/

agentsdk:
	@echo "$(COLOUR_BLUE)[OpenNHP] Building agent SDK... $(END_COLOUR)"
ifeq ($(OS_NAME), linux)
	cd endpoints && \
	go build -a -trimpath -buildmode=c-shared -ldflags ${LD_FLAGS} -v -o ../release/nhp-agent/nhp-agent.so ./agent/main/main.go ./agent/main/export.go && \
	gcc ./agent/sdkdemo/nhp-agent-demo.c -I ../release/nhp-agent -l:nhp-agent.so -L../release/nhp-agent -Wl,-rpath=. -o ../release/nhp-agent/nhp-agent-demo
endif

devicesdk:
	@echo "$(COLOUR_BLUE)[OpenNHP] Building nhp SDK... $(END_COLOUR)"
ifeq ($(OS_NAME), linux)
	cd nhp && \
	go build -a -trimpath -buildmode=c-shared -ldflags ${LD_FLAGS} -v -o ../release/nhp-device/nhpdevice.so ./core/main/main.go ./core/main/nhpdevice.go
#	gcc ./core/sdkdemo/nhp-device-demo.c -I ./release/nhp-device -I ./core/main -l:nhpdevice.so -L./release/nhp-device -Wl,-rpath=. -o ./release/nhp-device/nhp-device-demo
endif

plugins:
	@echo "$(COLOUR_BLUE)[OpenNHP] Building plugins... $(END_COLOUR)"
	@if test -d $(NHP_SERVER_PLUGINS); then $(MAKE) -C $(NHP_SERVER_PLUGINS); fi

test:
	@echo "[OpenNHP] Runing Tests for the Output Binaries ..."
	@echo "$(COLOUR_GREEN)[OpenNHP] All Tests Are Done!$(END_COLOUR)"

archive:
	@echo "$(COLOUR_BLUE)[OpenNHP] Start archiving... $(END_COLOUR)"
	@cd release && mkdir -p archive && tar -czvf ./archive/$(PACKAGE_FILE) nhp-agent nhp-ac nhp-de nhp-server
	@echo "$(COLOUR_GREEN)[OpenNHP] Package ${PACKAGE_FILE} archived!$(END_COLOUR)"

.PHONY: all generate-version-and-build init agentd acd serverd de agentsdk devicesdk plugins test archive
