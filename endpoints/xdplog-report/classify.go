package main

// Category is the human-readable grouping of a DROP reason.
type Category string

const (
	CatTCPScan   Category = "TCP scan / probe"
	CatSSH       Category = "SSH probe (non-whitelisted source)"
	CatNHPTCP    Category = "TCP to NHP knock port"
	CatUDPScan   Category = "UDP scan / non-NHP UDP"
	CatShortUDP  Category = "Short UDP (below NHP minimum frame)"
	CatOtherProt Category = "Other protocol (ICMP etc.)"
	CatIPv6      Category = "IPv6"
	CatOther     Category = "Other"
	CatUnknown   Category = "Unknown reason"
)

// dropCategories mirrors the DROP side of serverActionName in
// nhp/utils/ebpf/engine_linux.go; keep the two in sync. A cross-check test
// reads that file.
var dropCategories = map[string]Category{
	"TCP_SSH_OTHER": CatSSH,
	"TCP_NHP_PORT":  CatNHPTCP,
	"TCP_OTHER":     CatTCPScan,
	"UDP_OTHER":     CatUDPScan,
	"UDP_SHORT":     CatShortUDP,
	"NON_TCP_UDP":   CatOtherProt,
	"V6_OTHER":      CatIPv6,
	"OTHER":         CatOther,
}

// classify returns the category and whether the reason is known.
func classify(reason string) (Category, bool) {
	if c, ok := dropCategories[reason]; ok {
		return c, true
	}
	// UNKNOWN-n and any reason added to the XDP side later land here.
	return CatUnknown, false
}
