//go:build linux

package ebpf

// The kernel's verdict on an eBPF object.
//
// Compiling an object proves nothing about it: the kernel's verifier decides
// whether a program is loadable at all, and it can refuse a program that
// compiled without a warning (an instruction or complexity limit, a helper the
// program type does not allow). On the nhp-ac a refused object is an outage —
// nhp-acd cannot start, and the fail-closed backstop keeps tcp/443 shut — and
// on the nhp-server it is a silent loss of protection, because nhp-serverd
// falls open.
//
// So the objects are loaded before anything depends on them: in CI on the
// runner's kernel, and on the demo hosts on the kernel they actually run on,
// before the running daemon is stopped.
//
// This is that load, without the attach. It runs the steps the daemon's load
// runs — LoadCollectionSpec, the .rodata constants, then the load itself —
// and stops there. Nothing is attached, nothing outside the pin directory it
// owns is touched, and the pins it made are removed again.
//
// It loads through cilium/ebpf rather than through bpftool, and that is not a
// preference.
//
// bpftool's libbpf refuses these objects outright:
//
//	libbpf: sec 'xdp': program 'xdp_white_prog' is static and not supported
//
// The programs are declared static in the C sources, and libbpf does not
// support that; cilium/ebpf does, which is the only reason the daemons can
// load them at all. A gate built on bpftool therefore never asked the kernel
// anything: it reported its own plumbing.
//
// bpftool has a second problem, which is why it is not merely replaced but
// replaced by this. libbpf pins a map by name relative to its pin root, and
// bpftool sets no pin root, so a LIBBPF_PIN_BY_NAME map lands in
// /sys/fs/bpf/<name> — the very directory the running daemon pins into. The
// load then opens the live conn_track, the live spp and the live
// nhp_relay_ips, and binds the throwaway program to them. A load writes
// nothing into them, so that is harmless today; but it is not what such a
// gate promises, and the first object that changes a pinned map's shape would
// fail the gate with a mismatch against a map it was never meant to touch.
//
// Here the pin directory is one of the verify load's own, beside the daemon's
// entries rather than among them:
//
//	/sys/fs/bpf/nhp-verify-<pid>/conn_track   <- this load's, removed again
//	/sys/fs/bpf/conn_track                     <- the running daemon's, untouched
//
// The loader looks for a map it may reuse by name inside the directory it is
// given, so a verify load finds nothing to reuse and creates its own.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/cilium/ebpf"
)

// verifyLogSize is the size of the verifier log buffer the load asks the
// kernel for.
//
// The daemon's own load does not ask for a log: it has nothing to do with one
// beyond the error message. A check does, because the log is the answer —
// which branch the verifier could not follow, which helper it may not call —
// and an error without it leaves nothing to act on.
const verifyLogSize = 64 * 1024

// VerifyLoad asks the kernel whether it accepts the objects a daemon would
// load, without attaching them and without touching a pin a running daemon
// owns.
//
// The parameters are the loader's own (EngineLoadParams), so the constants
// written into the objects are the ones the daemon would write: the server's
// knock port and datagram floor, the AC's IPv6 knob and the two source
// narrowings. A zero value leaves the object's own default in place, which is
// what the daemon does too.
//
// The load needs CAP_BPF: a load without it is refused with EPERM, which says
// nothing about the program.
func VerifyLoad(params EngineLoadParams) error {
	switch params.Variant {
	case VariantServer:
		return verifyServerObjects(params)
	default:
		return verifyAcObjects(params)
	}
}

// verifyPinDir creates the pin directory a verify load owns.
//
// It is a directory of its own under the base, not the base itself: the
// daemon's pins are the base's entries (a map named conn_track, a program
// named xdp_white_prog), and the loader looks for a map it may reuse by name
// inside the directory it is given. A verify load that pinned into the base
// would find the running daemon's maps and bind to them — which is how a
// check that is supposed to change nothing ends up sharing the live filter's
// state.
//
// The name carries the pid, so two verify loads cannot share a directory, and
// the directory is cleared before it is created: a leftover from a run that
// was killed would otherwise be found, and reused.
func verifyPinDir(base string) (string, error) {
	if base == "" {
		base = DefaultPinDir
	}

	dir := filepath.Join(base, fmt.Sprintf("nhp-verify-%d", os.Getpid()))
	if err := os.RemoveAll(dir); err != nil {
		return "", fmt.Errorf("failed to clear the verify pin directory %s: %w", dir, err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create the verify pin directory %s: %w", dir, err)
	}
	return dir, nil
}

// verifyAcObjects loads the AC's two objects the way nhp-acd loads them.
//
// The XDP object and the TC egress object share their maps, so both go into
// the same pin directory: the second load finds the maps the first one
// created and reuses them, exactly as the daemon's two loads do.
//
// The assignment is the daemon's own (bpfObjects, tcBpfObjects), so the load
// fails by name on a map or a program the daemon needs and this one does not
// have — the same failure the daemon would report, from the same cause.
func verifyAcObjects(params EngineLoadParams) error {
	if params.ProgObjPath == "" || params.TcProgObjPath == "" {
		return errors.New("the AC's verify load needs both the XDP and the TC egress object")
	}

	spec, err := ebpf.LoadCollectionSpec(params.ProgObjPath)
	if err != nil {
		return fmt.Errorf("failed to read the AC's XDP object %s: %w", params.ProgObjPath, err)
	}
	tcSpec, err := ebpf.LoadCollectionSpec(params.TcProgObjPath)
	if err != nil {
		return fmt.Errorf("failed to read the AC's TC egress object %s: %w", params.TcProgObjPath, err)
	}

	// The same constants the daemon writes, and before the load: the verifier
	// folds them in as it goes, so afterwards there is no moment left at
	// which they could still be set.
	if constErr := setAcConstants(spec, params); constErr != nil {
		return constErr
	}

	dir, err := verifyPinDir(params.PinDir)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	opts := verifyOptions(dir)

	var objs bpfObjects
	if loadErr := spec.LoadAndAssign(&objs, &opts); loadErr != nil {
		return fmt.Errorf("the kernel refused the AC's XDP object %s: %w", params.ProgObjPath, loadErr)
	}
	defer closeAssigned(&objs)

	var tcObjs tcBpfObjects
	if loadErr := tcSpec.LoadAndAssign(&tcObjs, &opts); loadErr != nil {
		return fmt.Errorf("the kernel refused the AC's TC egress object %s: %w", params.TcProgObjPath, loadErr)
	}
	defer closeAssigned(&tcObjs)

	return nil
}

// verifyServerObjects loads the server's XDP object the way nhp-serverd
// loads it.
func verifyServerObjects(params EngineLoadParams) error {
	if params.ProgObjPath == "" {
		return errors.New("the server's verify load needs the XDP object")
	}

	spec, err := ebpf.LoadCollectionSpec(params.ProgObjPath)
	if err != nil {
		return fmt.Errorf("failed to read the server's XDP object %s: %w", params.ProgObjPath, err)
	}

	// The knock port and the datagram floor go in before the load, the same
	// way the daemon writes them: the verifier prunes the branches that
	// depend on them, so a load without them is not the program that will be
	// attached.
	if constErr := setServerConstants(spec, params); constErr != nil {
		return constErr
	}

	dir, err := verifyPinDir(params.PinDir)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	opts := verifyOptions(dir)

	var objs serverBpfObjects
	if loadErr := spec.LoadAndAssign(&objs, &opts); loadErr != nil {
		return fmt.Errorf("the kernel refused the server's XDP object %s: %w", params.ProgObjPath, loadErr)
	}
	defer closeAssigned(&objs)

	return nil
}

// verifyOptions are the loader's options with the pin directory of this load
// and the verifier log turned on.
//
// The log is what a refusal has to come with: the error says the kernel said
// no, and the log says why.
func verifyOptions(dir string) ebpf.CollectionOptions {
	return ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: dir,
		},
		Programs: ebpf.ProgramOptions{
			LogLevel:     ebpf.LogLevelInstruction,
			LogSizeStart: verifyLogSize,
		},
	}
}

// closeAssigned closes the programs and maps a load assigned.
//
// LoadAndAssign leaves the assigned objects open on purpose — the daemon
// attaches and pins them afterwards — so a verify load has to close them
// itself. They are closed before the pin directory is removed, so the pins go
// with the maps rather than lingering on a map that is still open.
func closeAssigned(assigned any) {
	value := reflect.ValueOf(assigned)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !field.IsValid() || field.Kind() != reflect.Pointer || field.IsNil() {
			continue
		}
		switch obj := field.Interface().(type) {
		case *ebpf.Program:
			obj.Close()
		case *ebpf.Map:
			obj.Close()
		}
	}
}
