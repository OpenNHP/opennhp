package agent_test

import (
	"github.com/OpenNHP/opennhp/endpoints/agent"
	"github.com/OpenNHP/opennhp/nhp/common"
)

// ExampleUdpAgent_UseInMemoryConfig starts an agent whose configuration is
// supplied in code. No etc/*.toml is read and no file watcher is armed, so a
// config.toml that appears later cannot override the identity, key, or cipher
// scheme set here.
func ExampleUdpAgent_UseInMemoryConfig() {
	a := &agent.UdpAgent{}
	a.UseInMemoryConfig(&agent.Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    "<base64 agent private key>",
		KnockUser:           agent.KnockUser{UserId: "user@example.com", OrganizationId: "org-1"},
	})
	if err := a.Start("/path/to/agent", 0); err != nil {
		return
	}
	defer a.Stop()
	// The agent now knocks with the in-code identity.
}

// ExampleUdpAgent_SetConfigSource shows the SDK pattern: start in memory mode
// with no key, then install the real key and identity programmatically (for
// example with ReinitWithKey and SetKnockUser) instead of reading
// etc/config.toml. The file watcher is never armed in either case.
func ExampleUdpAgent_SetConfigSource() {
	a := &agent.UdpAgent{}
	a.SetConfigSource(agent.ConfigSourceMemory)
	if err := a.Start("/path/to/agent", 0); err != nil {
		return
	}
	defer a.Stop()
	// Install the key and identity here, then knock.
}
