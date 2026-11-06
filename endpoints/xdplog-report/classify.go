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

// nonAttackCategories are the DROP categories that are not attacks.
//
// CatKnockExpired is the nhp-ac's own class: a peer that knocked, and whose
// knock window closed before it spoke again. That is a legitimate peer
// arriving late -- a retransmit, a session that outlived the window, an
// agent whose clock ran ahead -- and not a scan. Publishing it as one would
// put an authorized user's address into the site's attacker totals, masked to
// a /24 but counted, and counted as a country too.
//
// The packets are dropped either way: this is about the accounting, not about
// what the filter admits.
var nonAttackCategories = map[Category]struct{}{
	CatKnockExpired: {},
}

// isAttackReason reports whether a DROP reason counts towards the attack
// totals -- the packet total, the source and country counts, the per-source
// and per-port rankings, and the day's facts.
//
// An unknown reason counts as an attack. The classifier is a closed list, so a
// reason this build has not heard of is more likely a new attack class than a
// new non-attack one, and under-reporting attacks is the worse mistake.
func isAttackReason(reason string) bool {
	cat, known := classify(reason)
	if !known {
		return true
	}
	_, nonAttack := nonAttackCategories[cat]
	return !nonAttack
}
