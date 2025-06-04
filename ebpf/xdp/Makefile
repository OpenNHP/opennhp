# 检查 clang 是否安装
CLANG := $(shell command -v clang 2>/dev/null)
ifeq ($(CLANG),)
    $(error "clang is not installed. Please install clang to compile eBPF programs.")
endif

# 其余 Makefile 内容不变...
EBPF_SRC = nhp_ebpf_xdp.c
EBPF_OBJ = nhp_ebpf_xdp.o
CLANG_OPTS = -O2 -target bpf -g -Wall -I.

all: $(EBPF_OBJ)

$(EBPF_OBJ): $(EBPF_SRC)
	$(CLANG) $(CLANG_OPTS) -c $(EBPF_SRC) -o $(EBPF_OBJ)

clean:
	rm -f $(EBPF_OBJ)

.PHONY: all clean
