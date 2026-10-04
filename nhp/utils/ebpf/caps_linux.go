//go:build linux

package ebpf

// Giving back the capabilities the load needed.
//
// nhp-serverd is granted CAP_BPF, CAP_NET_ADMIN and CAP_PERFMON as *ambient*
// capabilities (terraform/demo/userdata/server.sh, and the drop-in the
// deploy-server job installs on existing hosts). It needs all three for about
// a second at startup: CAP_BPF to load the program and create its maps,
// CAP_NET_ADMIN to attach an XDP program to the interface, CAP_PERFMON to open
// the perf ring the event log reads.
//
// Keeping them for the life of the process is a much bigger grant than it
// looks. CAP_BPF together with CAP_PERFMON loads kprobe and tracing programs,
// which may call bpf_probe_read_kernel() — arbitrary kernel memory, including
// other processes' credentials and this daemon's own Noise private key.
// CAP_NET_ADMIN detaches this very filter, rewrites routes and rewrites
// netfilter rules. And this is a process that parses untrusted UDP from the
// whole internet and dlopens third-party auth plugins into its own address
// space, so any memory-safety bug or hostile plugin inherits whatever the
// process still holds.
//
// Nothing after the attach needs them:
//   - ReplaceRelayIPs drives the whitelist through a map fd the process already
//     holds (but see bpfSyscallNeedsCapability below);
//   - the perf reader is opened before the drop, and reading it is read(2) on
//     an fd;
//   - detaching — the watchdog, Stop() — is close(2) on the link fd, and
//     removing a bpffs pin is unlink(2) under a directory the service user
//     already owns.
//
// So the loader drops them as its last act, whether the attach succeeded or
// failed. Only the server variant does this: the AC's loader has the same
// lifetime question but a different answer (it rewrites its maps on every
// knock) and this change deliberately leaves its behavior alone.
//
// What it gives back is those three and nothing else. On the demo unit that is
// the whole set the process has — AmbientCapabilities= grants exactly these
// three to an otherwise unprivileged user — so the drop still leaves a daemon
// holding nothing at all, which is what the deploy job asserts. But nhp-serverd
// also ships as a container that runs as root (docker/Dockerfile.server), and
// there the process holds the full root set for reasons that have nothing to do
// with this loader: CAP_DAC_OVERRIDE for volume-mounted config owned by another
// uid, CAP_NET_BIND_SERVICE for an http.toml reload that rebinds 80/443, and
// whatever a dlopen'd plugin was built to expect. Resetting to the empty set
// there — on every path, including a host with no etc/xdp.toml that never asked
// for any of this — is an irreversible, silent and entirely unrelated privilege
// change, so the drop is written as a subtraction of the loader's three from
// whatever the process already holds, never as an assignment of a fixed set.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"kernel.org/pub/linux/libs/security/libcap/cap"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// unprivilegedBpfSysctl is the kernel's switch for whether bpf(2) may be called
// at all without CAP_BPF. 0 allows unprivileged calls; 1 and 2 ("disabled" and
// "disabled without recourse") make *every* command privileged, map updates on
// an fd this process already owns included.
const unprivilegedBpfSysctl = "/proc/sys/kernel/unprivileged_bpf_disabled"

// bpfSyscallNeedsCapability reports whether this kernel will still let the
// process write its own maps after the capabilities are gone.
//
// The hope is "yes": the fd is the capability, and a daemon that loaded its
// program while privileged should be able to keep its maps in step afterwards.
// That is true only while kernel.unprivileged_bpf_disabled is 0. Distros that
// build with CONFIG_BPF_UNPRIV_DEFAULT_OFF (Amazon Linux 2023, Ubuntu, Debian)
// ship it as 1 or 2, and there bpf(2) is gated wholesale — a dropped CAP_BPF
// would turn every later `ReplaceRelayIPs` into EPERM.
//
// That matters more than the capability does. The xdp.toml reload path is how a
// replaced relay's new address reaches the live whitelist, and it is the only
// remote way to correct a whitelist on a host whose only way in is that
// whitelist. Breaking it to drop one capability trades a hardening win for a
// lockout, so on such a kernel CAP_BPF is kept and said so in the log.
//
// An unreadable sysctl is read as "needs the capability": the conservative
// answer keeps the daemon working rather than quietly breaking reloads.
func bpfSyscallNeedsCapability() bool {
	return bpfSyscallNeedsCapabilityFrom(unprivilegedBpfSysctl)
}

func bpfSyscallNeedsCapabilityFrom(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return true
	}
	return v != 0
}

// loaderCaps are the three capabilities the unit grants for the load, and the
// only three this file ever takes away. Anything else the process happens to
// hold was granted for some other reason and is not the loader's to give back —
// see the "subtraction, never an assignment" note at the top of this file.
var loaderCaps = []cap.Value{cap.BPF, cap.NET_ADMIN, cap.PERFMON}

// reduceLoaderCaps builds the set the process should hold once the loader is
// done: held, minus CAP_NET_ADMIN and CAP_PERFMON, minus CAP_BPF unless the
// reload path still needs it. It reports whether CAP_BPF survived.
//
// Split out from the drop itself so a test can check the arithmetic on a set it
// constructs — the part that decides whether an unrelated capability such as
// CAP_DAC_OVERRIDE comes out the other side — without a privileged process to
// drop from.
func reduceLoaderCaps(held *cap.Set, keepBpf bool) (want *cap.Set, keptBpf bool, err error) {
	// A strict reduction of what the process already holds, never a wish list:
	// capset(2) refuses to raise a bit that is not in the current permitted
	// set, so a set built from scratch that named CAP_BPF on a host where the
	// unit's AmbientCapabilities= never took would fail the whole call with
	// EPERM and leave the capabilities — all of them — in place.
	want, err = held.Dup()
	if err != nil {
		return nil, false, fmt.Errorf("cannot copy the current capability set: %w", err)
	}

	drop := []cap.Value{cap.NET_ADMIN, cap.PERFMON}
	if keepBpf {
		has, getErr := held.GetFlag(cap.Permitted, cap.BPF)
		keptBpf = getErr == nil && has
	}
	if !keptBpf {
		drop = append(drop, cap.BPF)
	}

	for _, vec := range []cap.Flag{cap.Effective, cap.Permitted, cap.Inheritable} {
		if setErr := want.SetFlag(vec, false, drop...); setErr != nil {
			return nil, false, fmt.Errorf("cannot build the reduced capability set: %w", setErr)
		}
	}
	return want, keptBpf, nil
}

// dropLoaderPrivileges takes CAP_BPF, CAP_NET_ADMIN and CAP_PERFMON out of the
// ambient, permitted, effective and inheritable sets — all three where the
// kernel allows it, CAP_BPF kept where bpf(2) itself is privileged.
//
// It subtracts those three; it does not assign a fixed set. On the demo unit
// the two are the same thing (the grant is exactly these three, so what is left
// is empty), and everywhere else the difference is the point: a root container
// keeps CAP_DAC_OVERRIDE and CAP_NET_BIND_SERVICE, which it was given for
// reasons this loader knows nothing about and cannot get back once dropped.
//
// allowKeepBpf is the caller's answer to "is there still a map to write?".
// Only a load that produced a live filter has one: on a failed load there is no
// handle, no whitelist and no reload path, so the CAP_BPF exemption below has
// nothing left to protect and the drop is unconditional. That is the case the
// daemon runs in for the longest — a load that fails is fail-open for the life
// of the process — so it is the one that least deserves a capability.
//
// Both steps are process-wide, not just this goroutine's thread: libcap's cap
// package routes prctl(2) and capset(2) through psx, which applies them to
// every thread of the process. A per-thread drop would be worse than none —
// /proc/<pid>/status would show CapEff: 0 for the thread group leader while the
// runtime's other threads carried the full set, i.e. a check that passes and a
// guarantee that does not hold. (The pure-Go syscall.AllThreadsSyscall cannot
// be used here: nhp-serverd links cgo, by way of the plugin package, and that
// function returns ENOTSUP in any binary that does.)
//
// The bounding set is left alone: dropping from it needs CAP_SETPCAP, which
// this daemon is deliberately not granted. It is already narrowed to these
// three by CapabilityBoundingSet= in the unit, and with the loader's bits gone
// from the permitted set and NoNewPrivileges=true a bounding bit grants nothing
// on its own.
func dropLoaderPrivileges(allowKeepBpf bool) error {
	keepBpf := allowKeepBpf && bpfSyscallNeedsCapability()

	// Ambient first. It is what survives an exec, so it is the one set that
	// could hand these capabilities to something that is not this program at
	// all. Clearing permitted and inheritable below forces the loader's three
	// out of it anyway — the kernel holds no ambient bit that is not in both —
	// but doing it explicitly means a failure here is reported as itself. Only
	// the three are lowered, and all three even when CAP_BPF is kept below: an
	// exec'd child has no map of ours to write, so nothing needs to survive one.
	if err := cap.SetAmbient(false, loaderCaps...); err != nil {
		return fmt.Errorf("cannot clear the loader's ambient capabilities: %w", err)
	}

	held := cap.GetProc()
	want, keptBpf, err := reduceLoaderCaps(held, keepBpf)
	if err != nil {
		return err
	}
	if err := want.SetProc(); err != nil {
		return fmt.Errorf("cannot reduce the capability set from %q to %q: %w", held, want, err)
	}

	if keptBpf {
		log.Info("XDP loader privileges dropped: CAP_NET_ADMIN and CAP_PERFMON released, CAP_BPF kept because %s is non-zero (bpf(2) is privileged on this kernel, and the xdp.toml reload path writes the relay whitelist map). The daemon now holds %q",
			unprivilegedBpfSysctl, want)
		return nil
	}
	log.Info("XDP loader privileges dropped: CAP_BPF, CAP_NET_ADMIN and CAP_PERFMON released. The daemon now holds %q", want)
	return nil
}

// dropLoaderPrivilegesOrWarn is the form the loader calls: a failure to drop is
// not a reason to unwind a filter that is working, so it is logged loudly and
// the daemon carries on with the capabilities it was given.
func dropLoaderPrivilegesOrWarn(allowKeepBpf bool) {
	if err := dropLoaderPrivileges(allowKeepBpf); err != nil {
		log.Error("could not drop the XDP loader's capabilities; nhp-serverd keeps CAP_BPF, CAP_NET_ADMIN and CAP_PERFMON for the rest of its life: %v", err)
	}
}

var (
	dropOnce sync.Once
	// dropLoaderPrivilegesFn is the drop itself, behind a variable so a test
	// can observe the once-guard without disarming the rest of the test
	// binary. Never reassigned outside tests.
	dropLoaderPrivilegesFn = dropLoaderPrivilegesOrWarn
)

// DropLoaderPrivileges gives the loader's capabilities back, at most once per
// process, and is the call every path out of the decision to filter has to
// reach — not just the paths that got as far as loading something.
//
// The loader drops on its own way out, but it is only one of the outcomes. A
// host with no etc/xdp.toml, an unparsable one, Enabled = false, a RelayIPs the
// trie cannot hold, an out-of-range NhpMinFrameBytes or a listener the filter
// would black-hole never reaches the loader at all — and the unit grants the
// three capabilities as *ambient* regardless, so each of those states used to
// mean a daemon holding CAP_BPF, CAP_NET_ADMIN and CAP_PERFMON for its whole
// life having never once used them. That is the same grant the loader is
// careful to give back after a failed load, held for the same reason (there is
// no filter) by a process in the same position (untrusted UDP, dlopen'd
// plugins) — so the caller drops on those paths too, with allowKeepBpf false
// because a host that is not filtering has no whitelist map to reload.
//
// Once, because the two callers overlap on the attach path: the loader's own
// deferred drop runs first and settles the CAP_BPF question with the answer
// that has a live map behind it, and the caller's later unconditional call must
// not then take CAP_BPF away from a filter whose reload path needs it. First
// call wins, and on every non-attach path the first call is the caller's.
func DropLoaderPrivileges(allowKeepBpf bool) {
	dropOnce.Do(func() { dropLoaderPrivilegesFn(allowKeepBpf) })
}
