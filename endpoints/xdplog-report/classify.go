package main

// Category is the human-readable grouping of a DROP reason.
type Category string

const (
	CatTCPScan   Category = "TCP scan / probe"
	CatSYNScan   Category = "TCP SYN scan"
	CatSSH       Category = "SSH probe (non-whitelisted source)"
	CatNHPTCP    Category = "TCP to NHP knock port"
	CatUDPScan   Category = "UDP scan / non-NHP UDP"
	CatShortUDP  Category = "Short UDP (below NHP minimum frame)"
	CatOtherProt Category = "Other protocol (ICMP etc.)"
	CatIPv6      Category = "IPv6"
	CatOther     Category = "Other"
	CatUnknown   Category = "Unknown reason"
	// CatKnockExpired is the nhp-ac's own class: a peer whose knock window
	// closed, as opposed to a peer that never knocked at all. It is not a
	// scan, so it does not belong under any of the categories above.
	CatKnockExpired Category = "Knock window expired"
)

// dropCategories mirrors the DROP side of serverActionName in
// nhp/utils/ebpf/engine_linux.go. A cross-check test reads that file and
// requires every reason it defines to be present here. This map may have
// entries the checked-in engine does not, for deployments whose XDP build
// leads the branch.
var dropCategories = map[string]Category{
	"TCP_SSH_OTHER":    CatSSH,
	"TCP_NHP_PORT":     CatNHPTCP,
	"TCP_OTHER":        CatTCPScan,
	"TCP_SYN_SCAN":     CatSYNScan,
	"TCP_STEALTH_SCAN": CatTCPScan,
	"UDP_OTHER":        CatUDPScan,
	"UDP_SHORT":        CatShortUDP,
	"NON_TCP_UDP":      CatOtherProt,
	"V6_OTHER":         CatIPv6,
	"OTHER":            CatOther,
	"KNOCK_EXPIRED":    CatKnockExpired,
}

// classify returns the category and whether the reason is known.
func classify(reason string) (Category, bool) {
	if c, ok := dropCategories[reason]; ok {
		return c, true
	}
	// UNKNOWN-n and any reason added to the XDP side later land here.
	return CatUnknown, false
}
