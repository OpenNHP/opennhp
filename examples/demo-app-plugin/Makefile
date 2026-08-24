export GO111MODULE := on

all: build

PluginName = demo-app
TargetName = ${PluginName}.so
ServerDir = ../../release/nhp-server
LD_FLAGS = "-s -w"

# Color definition
COLOUR_GREEN=\033[0;32m
COLOUR_BLUE=\033[0;34m
END_COLOUR=\033[0M

build:
	@echo "$(COLOUR_BLUE)[Plugin-${PluginName}] Start building... $(END_COLOUR)"
	go build -buildmode=plugin -trimpath -ldflags ${LD_FLAGS} -v \
	    -o ${ServerDir}/plugins/${PluginName}/${TargetName} ./main.go
	mkdir -p ${ServerDir}/plugins/${PluginName}/etc
	cp ./etc/*.toml ${ServerDir}/plugins/${PluginName}/etc/
	@echo "$(COLOUR_GREEN)[Plugin-${PluginName}] Successfully built!$(END_COLOUR)"

.PHONY: all build
