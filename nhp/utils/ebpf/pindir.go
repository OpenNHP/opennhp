package ebpf

import "path/filepath"

// Where each variant's maps are pinned.
//
// The AC and the server have no maps in common, but both pin into a
// directory on the bpffs, and the AC's rule writers -- the ones that install
// a knock's grant -- have to open the very maps the running filter reads.
// So the writers need the directory that filter's load pinned into, not a
// hard-coded path.
//
// A hard-coded path was wrong twice over. It did not follow a load that
// pinned elsewhere, and it made it impossible for a test to write a rule
// without writing it into the map of an access controller running on the
// same host -- which is a test that opens a port on a live host.
//
// Unsynchronized, like the links in engine_linux.go: a daemon loads once at
// startup and writes rules after that, and the two do not overlap.
var (
	acPinDir     = DefaultPinDir
	serverPinDir = DefaultPinDir
)

// activePinDir is the directory the variant's maps are pinned into.
func activePinDir(variant EngineVariant) string {
	if variant == VariantServer {
		return serverPinDir
	}
	return acPinDir
}

// recordPinDir remembers where a load pinned the variant's maps.
//
// The rule writers read it back through activePinDir, so a grant written
// after a load lands in the map that load created -- and a test that points
// the variant at a directory of its own never reaches the live filter's.
func recordPinDir(variant EngineVariant, dir string) {
	if variant == VariantServer {
		serverPinDir = dir
		return
	}
	acPinDir = dir
}

// acMapPath is the path of one of the AC's pinned maps.
//
// Every whitelist map the rule writers touch belongs to the AC's filter:
// the server's own filter pins only its program and its relay list, so it
// has no map to write a grant into.
func acMapPath(name string) string {
	return filepath.Join(activePinDir(VariantAC), name)
}
