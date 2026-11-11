// Command nhp-ebpf-verify asks the kernel whether it accepts an eBPF object.
//
// The kernel's verifier is the only thing that can refuse a program: an
// object that compiles can still be rejected (an instruction or complexity
// limit, a helper the program type does not allow), and the rejection is
// what takes nhp-acd down — the fail-closed backstop then keeps tcp/443 shut
// — or leaves nhp-serverd without an ingress filter, which it survives by
// falling open.
//
// So the object is loaded before anything depends on it: in CI on the
// runner's kernel, and on the demo hosts on the kernel they actually run on,
// before the running daemon is stopped.
//
// This loads the object the way the daemon loads it — the same loader, the
// same .rodata constants, the same maps pinned by name — and stops before the
// attach. Nothing is attached, nothing outside the pin directory it owns is
// touched, and the pins it made are removed again. See VerifyLoad in
// nhp/utils/ebpf for why this goes through the loader and not through
// bpftool.
//
// It needs CAP_BPF, which in practice means root: a load without the
// capability is refused with EPERM, and that says nothing about the program.
//
// Usage:
//
//	nhp-ebpf-verify --variant ac --obj nhp_ebpf_xdp.o --tc-obj tc_egress.o
//	nhp-ebpf-verify --variant server --obj nhp_server_xdp.o --port 62206
//
// Exit status:
//
//	0  the kernel accepted the object
//	1  the kernel refused it (the verifier log is on stderr)
//	2  the arguments were wrong
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/OpenNHP/opennhp/nhp/utils/ebpf"
	"github.com/OpenNHP/opennhp/nhp/version"
)

const (
	exitAccepted = 0
	exitRefused  = 1
	exitUsage    = 2
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		variant    = flag.String("variant", "", `which filter's object to check: "ac" or "server"`)
		obj        = flag.String("obj", "", "the XDP object to check")
		tcObj      = flag.String("tc-obj", "", "the TC egress object to check (the AC only)")
		pinDir     = flag.String("pin-dir", "", "the bpffs directory the verify load pins into (default "+ebpf.DefaultPinDir+")")
		port       = flag.Int("port", 0, "the UDP port the daemon listens on (the server only)")
		minFrame   = flag.Int("min-frame-bytes", 0, "the shortest datagram the daemon accepts on it (the server only)")
		noIPv6     = flag.Bool("disable-ipv6-filter", false, "the AC's IPv6 tree is off (the AC only)")
		dhcpSource = flag.String("dhcp-source", "", "the address the DHCP answer comes from")
		ntpSource  = flag.String("ntp-source", "", "the address the NTP answer comes from")
	)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s --variant ac|server --obj <file.o> [--tc-obj <file.o>]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	params, err := paramsFromFlags(*variant, *obj, *tcObj, *pinDir, *port, *minFrame, *noIPv6, *dhcpSource, *ntpSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		flag.Usage()
		return exitUsage
	}

	if loadErr := ebpf.VerifyLoad(params); loadErr != nil {
		// The verifier log is part of the error: cilium/ebpf retries the load
		// with the log enabled when the first attempt is refused, and returns
		// it. Printing the error therefore prints the log.
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], loadErr)
		return exitRefused
	}

	fmt.Printf("the kernel accepted %s (%s, %s)\n", *obj, version.Version, version.CommitId)
	return exitAccepted
}

// paramsFromFlags turns the command line into the loader's own parameters.
//
// The values are the ones the daemon would use, because they are written into
// the object's .rodata before it is loaded and the verifier folds them in: a
// check that left them at their defaults would be checking a slightly
// different program than the one that gets attached. An unset value means
// "the object's own default", which is also what the daemon does with an
// unset setting.
func paramsFromFlags(variant, obj, tcObj, pinDir string, port, minFrame int, noIPv6 bool, dhcpSource, ntpSource string) (ebpf.EngineLoadParams, error) {
	params := ebpf.EngineLoadParams{
		ProgObjPath:       obj,
		TcProgObjPath:     tcObj,
		PinDir:            pinDir,
		DisableIPv6Filter: noIPv6,
	}

	switch variant {
	case "ac":
		params.Variant = ebpf.VariantAC
		if tcObj == "" {
			return params, errors.New("the AC loads two objects: --tc-obj is required as well")
		}
	case "server":
		params.Variant = ebpf.VariantServer
		if tcObj != "" {
			return params, errors.New("the server loads one object: --tc-obj does not apply to it")
		}
	case "":
		return params, errors.New("--variant is required: \"ac\" or \"server\"")
	default:
		return params, fmt.Errorf("unknown --variant %q: expected \"ac\" or \"server\"", variant)
	}

	if obj == "" {
		return params, errors.New("--obj is required")
	}

	if port < 0 || port > 65535 {
		return params, fmt.Errorf("--port %d is not a UDP port", port)
	}
	params.NhpPort = uint16(port)

	if minFrame < 0 || minFrame > 65535 {
		return params, fmt.Errorf("--min-frame-bytes %d does not fit in a UDP length", minFrame)
	}
	params.NhpMinFrameBytes = uint16(minFrame)

	dhcp, err := ebpf.ParseSourceNet(dhcpSource)
	if err != nil {
		return params, fmt.Errorf("--dhcp-source: %w", err)
	}
	params.DhcpSource = dhcp

	ntp, err := ebpf.ParseSourceNet(ntpSource)
	if err != nil {
		return params, fmt.Errorf("--ntp-source: %w", err)
	}
	params.NtpSource = ntp

	return params, nil
}
