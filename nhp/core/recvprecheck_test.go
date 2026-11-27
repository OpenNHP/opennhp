package core

import "testing"

// RecvPrecheck reads the header in place, so it must reject packets too
// short to hold one rather than panic or read past them. This covers a bare
// slice (as the C SDK passes) and a pool buffer whose content is shorter
// than the buffer (as the server's relay path passes for the inner packet).
func TestRecvPrecheckRejectsShortPackets(t *testing.T) {
	d := NewDevice(NHP_SERVER, NewECDH(ECC_CURVE25519).PrivateKey(), nil)
	if d == nil {
		t.Fatal("NewDevice failed")
	}
	minLen := (&Packet{Content: make([]byte, PacketBufferSize)}).MinimalLength()

	for _, n := range []int{0, 1, 7, 12, minLen - 1} {
		bare := &Packet{Content: make([]byte, n)}
		pooled := d.AllocatePoolPacket()
		pooled.Content = pooled.Buf[:n]
		for name, pkt := range map[string]*Packet{"bare": bare, "pooled": pooled} {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s %d bytes: panicked: %v", name, n, r)
					}
				}()
				if _, _, err := d.RecvPrecheck(pkt); err == nil {
					t.Errorf("%s %d bytes: accepted", name, n)
				}
			}()
		}
		d.ReleasePoolPacket(pooled)
	}
}
