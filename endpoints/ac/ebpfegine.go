//go:build linux

package ac

import (
	// "log"

	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/version"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

type bpfObjects struct {
	XdpProg       *ebpf.Program `ebpf:"xdp_white_prog"`
	Whitelist     *ebpf.Map     `ebpf:"whitelist"`
	Icmpwhitelist *ebpf.Map     `ebpf:"icmpwhitelist"`
	Sdwhitelist   *ebpf.Map     `ebpf:"sdwhitelist"`
	Srcportlist   *ebpf.Map     `ebpf:"src_port_list"`
	Conntrack     *ebpf.Map     `ebpf:"conn_track"`
}

var xdpLink link.Link

func (a *UdpAC) Ebpf_engine_load() error {
	log.Info("=== NHP-AC Ebpf Engine %s started   ===", version.Version)
	log.Info("=========================================================")
	err := a.loadBaseConfig()
	if err != nil {
		log.Error("Failed to loadBaseConfig for ac : %v", err)
		return err
	}

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Error("Failed to remove memlock limit: %v", err)
	}

	var ebpfenginename string
	if len(a.config.EbpfEngineName) > 0 {
		ebpfenginename = a.config.EbpfEngineName
	}
	bpfDir := "etc"
	specPath := filepath.Join(bpfDir, ebpfenginename)

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		log.Error("eBPF object file not found: %s", specPath)
		return err
	}

	spec, err := ebpf.LoadCollectionSpec(specPath)
	if err != nil {
		log.Error("failed to load eBPF object: %v", err)
		return err
	}

	var objs bpfObjects
	if err := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf/", // automatically mounted to
		},
	}); err != nil {
		log.Error("Failed to load and assign eBPF objects: %v", err)
	}
	// defer objs.XdpProg.Close()
	// defer objs.Whitelist.Close()
	// defer objs.Conntrack.Close()

	if err := objs.XdpProg.Pin("/sys/fs/bpf/xdp_white_prog"); err != nil {
		log.Error("Failed to pin XDP program: %v", err)
		return err
	}

	ifaceName, err := getDefaultRouteInterface()
	if err != nil {
		log.Error("Failed to get default route interface: %v", err)
		return err
	}
	log.Info("Default route interface: %s\n", ifaceName)
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Info("Failed to find interface %s: %v\n", ifaceName, err)
		os.Exit(1)
	}

	xdpLink, err = link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpProg,
		Interface: iface.Index,
		Flags:     link.XDPGenericMode, // XDPGenericMode and XDPDriverMode
	})
	if err != nil {
		log.Error("Failed to attach XDP program to enp2s0: %v", err)
	}
	// defer link.Close()
	log.Info("Successfully attached XDP program to enp2s0")
	return nil
}

func getDefaultRouteInterface() (string, error) {
	cmd := exec.Command("ip", "route")
	output, err := cmd.Output()
	if err != nil {
		log.Error("Error running ip route:", err)
		return "", err
	}

	re := regexp.MustCompile(`default via (\S+) dev (\S+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 3 {
		log.Error("Failed to parse default route")
		return "", fmt.Errorf("failed to parse default route")
	}
	interfaceName := matches[2]
	return interfaceName, nil
}
