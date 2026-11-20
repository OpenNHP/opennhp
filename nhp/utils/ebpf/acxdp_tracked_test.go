package ebpf

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cilium/ebpf"
)

// The AC's XDP object is tracked in the source tree, next to the daemon's
// configuration: endpoints/ac/main/etc/nhp_ac_xdp.o.
//
// It is there because the loader's object paths are relative to the working
// directory (endpoints/ac/ebpf/ebpfegine.go asks for "etc/nhp_ac_xdp.o"), so
// a run started from endpoints/ac/main loads the object from there and not
// from the release directory `make ebpf-objs` fills.
//
// A tracked build artifact is a promise, and these tests keep the half of it
// that does not need a kernel:
//
//   - the artifact is an eBPF object, and it carries the program the loader
//     attaches. The kernel is asked about this exact file too, on every PR:
//     the ebpf-verifier job in .github/workflows/ubuntu-build.yml loads it
//     with nhp-ebpf-verify. An object nobody loads is an object nobody
//     knows -- a refusal of it would otherwise surface only on the machine
//     that runs from the source tree, where the fail-closed backstop then
//     keeps tcp/443 shut.
//   - and it is the program the current source builds. A stale copy -- one
//     left behind when a map's shape or a constant changed -- would be
//     loaded by such a run and would enforce the old policy.
//
// The comparison is on the object's interface and not on its instructions:
// the map shapes and the constants come from the C sources, so they are the
// same whatever compiler built the file, while the instructions are the
// compiler's business.

// acTrackedXdpProg is the program the AC's loader attaches, as the loader
// names it (see bpfObjects in engine_linux.go).
const acTrackedXdpProg = "xdp_white_prog"

// acTrackedObjPath is the AC's XDP object as it is tracked in the source
// tree.
func acTrackedObjPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "endpoints", "ac", "main", "etc", "nhp_ac_xdp.o")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			t.Skipf("%s is not tracked in this tree; there is nothing to keep honest", path)
		}
		t.Fatalf("cannot stat %s: %v", path, err)
	}
	return path
}

// TestTrackedAcObjectIsTheAcFilter checks that the tracked object is the
// AC's ingress program and not some other ELF file that happens to be
// committed: the loader attaches xdp_white_prog, and an object without it
// fails the load at startup.
func TestTrackedAcObjectIsTheAcFilter(t *testing.T) {
	path := acTrackedObjPath(t)

	spec, err := ebpf.LoadCollectionSpec(path)
	if err != nil {
		t.Fatalf("read the tracked AC object %s: %v", path, err)
	}

	prog, ok := spec.Programs[acTrackedXdpProg]
	if !ok {
		t.Fatalf("%s carries no %s program; the loader attaches that one", path, acTrackedXdpProg)
	}
	if prog.Type != ebpf.XDP {
		t.Errorf("%s: %s is a %s program, but the loader attaches it to XDP", path, acTrackedXdpProg, prog.Type)
	}
}

// TestTrackedAcObjectMatchesTheBuiltOne compares the tracked object with the
// one `make ebpf-objs` builds from the current source.
//
// A tracked object that no longer matches its source is a stale object: it
// is loaded by a run from the source tree, and it enforces the map shapes
// and the constants of the day it was built. The comparison is on the
// interface -- the maps, the programs and the .rodata constants -- because
// that is what the loader and the program agree on; the instructions
// themselves are the compiler's business and are left alone.
func TestTrackedAcObjectMatchesTheBuiltOne(t *testing.T) {
	trackedPath := acTrackedObjPath(t)
	builtPath := filepath.Join("..", "..", "..", "release", "nhp-ac", "etc", "nhp_ac_xdp.o")
	if _, err := os.Stat(builtPath); err != nil {
		if os.IsNotExist(err) {
			t.Skipf("%s is not built; run `make ebpf-objs` to compare the tracked object with a fresh one", builtPath)
		}
		t.Fatalf("cannot stat %s: %v", builtPath, err)
	}

	tracked := loadObjectSpec(t, trackedPath)
	built := loadObjectSpec(t, builtPath)

	comparePrograms(t, trackedPath, builtPath, tracked, built)
	compareMaps(t, trackedPath, builtPath, tracked, built)
	compareConstants(t, trackedPath, builtPath, tracked, built)
}

// loadObjectSpec parses an eBPF object the way the loader does.
func loadObjectSpec(t *testing.T, path string) *ebpf.CollectionSpec {
	t.Helper()
	spec, err := ebpf.LoadCollectionSpec(path)
	if err != nil {
		t.Fatalf("read the eBPF object %s: %v", path, err)
	}
	return spec
}

// comparePrograms checks that the two objects carry the same programs.
func comparePrograms(t *testing.T, trackedPath, builtPath string, tracked, built *ebpf.CollectionSpec) {
	t.Helper()

	for name, trackedProg := range tracked.Programs {
		builtProg, ok := built.Programs[name]
		if !ok {
			t.Errorf("%s carries the program %q, which %s does not: the tracked object is stale", trackedPath, name, builtPath)
			continue
		}
		if trackedProg.Type != builtProg.Type {
			t.Errorf("%s: the program %q is a %s, %s builds it as a %s", trackedPath, name, trackedProg.Type, builtPath, builtProg.Type)
		}
		if trackedProg.SectionName != builtProg.SectionName {
			t.Errorf("%s: the program %q is in the section %q, %s builds it in %q", trackedPath, name, trackedProg.SectionName, builtPath, builtProg.SectionName)
		}
	}
	for name := range built.Programs {
		if _, ok := tracked.Programs[name]; !ok {
			t.Errorf("%s carries no program %q, which %s builds: the tracked object is stale", trackedPath, name, builtPath)
		}
	}
}

// compareMaps checks that the two objects declare the same maps, with the
// same shape.
//
// The maps of the data sections (.rodata, .data, .bss) are left out: they
// are the maps cilium/ebpf synthesises for the sections the constants live
// in, and their size depends on how the compiler laid those sections out.
// The constants themselves are compared by compareConstants.
func compareMaps(t *testing.T, trackedPath, builtPath string, tracked, built *ebpf.CollectionSpec) {
	t.Helper()

	for name, trackedMap := range tracked.Maps {
		if isDataSectionMap(name) {
			continue
		}
		builtMap, ok := built.Maps[name]
		if !ok {
			t.Errorf("%s carries the map %q, which %s does not: the tracked object is stale", trackedPath, name, builtPath)
			continue
		}
		if trackedMap.Type != builtMap.Type {
			t.Errorf("%s: the map %q is a %s, %s builds it as a %s", trackedPath, name, trackedMap.Type, builtPath, builtMap.Type)
		}
		if trackedMap.KeySize != builtMap.KeySize {
			t.Errorf("%s: the map %q has a %d byte key, %s builds it with %d bytes", trackedPath, name, trackedMap.KeySize, builtPath, builtMap.KeySize)
		}
		if trackedMap.ValueSize != builtMap.ValueSize {
			t.Errorf("%s: the map %q has a %d byte value, %s builds it with %d bytes", trackedPath, name, trackedMap.ValueSize, builtPath, builtMap.ValueSize)
		}
		if trackedMap.MaxEntries != builtMap.MaxEntries {
			t.Errorf("%s: the map %q holds %d entries, %s builds it with %d", trackedPath, name, trackedMap.MaxEntries, builtPath, builtMap.MaxEntries)
		}
		if trackedMap.Flags != builtMap.Flags {
			t.Errorf("%s: the map %q has the flags %d, %s builds it with %d", trackedPath, name, trackedMap.Flags, builtPath, builtMap.Flags)
		}
		if trackedMap.Pinning != builtMap.Pinning {
			t.Errorf("%s: the map %q is pinned %v, %s builds it pinned %v", trackedPath, name, trackedMap.Pinning, builtPath, builtMap.Pinning)
		}
	}
	for name := range built.Maps {
		if isDataSectionMap(name) {
			continue
		}
		if _, ok := tracked.Maps[name]; !ok {
			t.Errorf("%s carries no map %q, which %s builds: the tracked object is stale", trackedPath, name, builtPath)
		}
	}
}

// isDataSectionMap reports whether a map is the one cilium/ebpf synthesises
// for a data section (.rodata, .data, .bss) rather than one of the
// program's own maps.
func isDataSectionMap(name string) bool {
	return name == ".bss" || strings.HasPrefix(name, ".data") || strings.HasPrefix(name, ".rodata")
}

// compareConstants checks that the two objects carry the same .rodata
// constants, with the same values.
//
// These are the constants the loader writes before the load (the IPv6 knob
// and the two source ranges), so a tracked object whose constants differ
// from the current source's would enforce the old policy.
func compareConstants(t *testing.T, trackedPath, builtPath string, tracked, built *ebpf.CollectionSpec) {
	t.Helper()

	for name, trackedVar := range tracked.Variables {
		builtVar, ok := built.Variables[name]
		if !ok {
			t.Errorf("%s carries the constant %q, which %s does not: the tracked object is stale", trackedPath, name, builtPath)
			continue
		}
		if !bytes.Equal(trackedVar.Value, builtVar.Value) {
			t.Errorf("%s: the constant %q is %v, %s builds it as %v", trackedPath, name, trackedVar.Value, builtPath, builtVar.Value)
		}
	}
	for name := range built.Variables {
		if _, ok := tracked.Variables[name]; !ok {
			t.Errorf("%s carries no constant %q, which %s builds: the tracked object is stale", trackedPath, name, builtPath)
		}
	}
}
