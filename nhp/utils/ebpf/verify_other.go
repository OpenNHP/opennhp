//go:build !linux

package ebpf

// VerifyLoad is Linux-only: it loads an eBPF object into the kernel, and
// there is nothing to load off Linux.
//
// It is declared here so that the command that drives it builds everywhere —
// the tool is part of the tree, and a build that only works on the platform
// it inspects is a build that is never checked on the other one.
func VerifyLoad(params EngineLoadParams) error {
	return ErrEBPFSupportedOnlyOnLinux
}
