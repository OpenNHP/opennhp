# NHP Agent: configuration sources

`agent.UdpAgent` can get its configuration from one of two sources. This guide
is for SDK and embedded users who want to configure the agent in code instead
of with `etc/*.toml` files. For the C, Android and iOS SDK bindings, see
[docs/agent_sdk.md](../../docs/agent_sdk.md).

## 1. Choosing a mode

| | `ConfigSourceFile` (default) | `ConfigSourceMemory` |
| --- | --- | --- |
| Where config comes from | `etc/config.toml`, `dhp.toml`, `server.toml`, `resource.toml` under the working directory | Your code: a `Config` value plus `AddServer` / `AddResource` |
| Hot reload | Yes, every file is watched and reloaded live | No watcher is ever armed |
| Can a file on disk change the agent? | Yes | No. A `config.toml` that appears later cannot override the key, identity, cipher scheme, peers or resources |
| Typical use | The standalone `nhp-agentd` daemon | Mobile SDKs, agents embedded in another app, tests |

The zero value is `ConfigSourceFile`, so existing callers don't change. Choose
memory mode when the host app owns the agent's state, for example when a mobile
app registers a key and keeps it in its own storage, or when a test needs an
agent that reads nothing from disk.

## 2. Minimal end-to-end example

There is no `server.toml` or `resource.toml` in memory mode. Servers are added
with `AddServer` and resources with `AddResource`, **after** `Start` (both need
the device that `Start` creates).

```go
import (
	"github.com/OpenNHP/opennhp/endpoints/agent"
	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

func run() error {
	a := &agent.UdpAgent{}

	// 1. Supply the base config in code. Call this before Start.
	a.UseInMemoryConfig(&agent.Config{
		LogLevel:            4,
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    agentPrivateKeyBase64, // must not be empty
		KnockUser: agent.KnockUser{
			UserId:         "user@example.com",
			OrganizationId: "org-1",
		},
	})

	// 2. Start. workingDir is used for logs (see section 5), not for config.
	if err := a.Start("/path/to/workdir", 4); err != nil {
		return err
	}
	defer a.Stop()

	// 3. Add the nhp-server peer (replaces server.toml).
	a.AddServer(&core.UdpPeer{
		Type:         core.NHP_SERVER,
		PubKeyBase64: serverPublicKeyBase64,
		Hostname:     "nhp.example.com", // or Ip: "203.0.113.10"
		Port:         62206,
		ExpireTime:   0, // Unix seconds; 0 never expires
	})

	// 4. Add the protected resource (replaces resource.toml). Reference the
	//    server by its public key: a server added with AddServer has no
	//    operator-chosen cluster name.
	if err := a.AddResource(&agent.KnockResource{
		AuthServiceId: "example",
		ResourceId:    "demo",
		ServerPubKey:  serverPublicKeyBase64,
	}); err != nil {
		return err // e.g. common.ErrKnockResourceUnknownClusterPubKey
	}

	// 5. Knock once...
	target := a.FindKnockTarget("example", "demo")
	ack, err := a.Knock(target)
	if err != nil {
		return err
	}
	_ = ack // ack.ResourceHost, ack.OpenTime, ...

	// ...or keep every added resource open with the periodic knock loop.
	// a.StartKnockLoop()
	return nil
}
```

A runnable version of steps 1 and 2 is in
[example_config_source_test.go](example_config_source_test.go).

## 3. `UseInMemoryConfig` vs `SetConfigSource`

Both switch the agent to memory mode and must be called before `Start`.

- **`UseInMemoryConfig(cfg)`**: use it when you already have the key and
  identity when the agent starts. The `Config` is copied, so changing `cfg`
  later has no effect on the agent. `cfg.PrivateKeyBase64` must be set or
  `Start` fails.
- **`SetConfigSource(ConfigSourceMemory)`** (same as `UseInMemoryConfig(nil)`):
  use it when the key doesn't exist yet, as in SDK registration. The first
  `Start` runs on a throwaway key, and you install the real key and identity
  afterwards:

  ```go
  a.SetConfigSource(agent.ConfigSourceMemory)
  if err := a.Start(workDir, 4); err != nil { ... }

  prk := core.NewECDH(core.ECC_CURVE25519).PrivateKey() // or a key you stored
  if err := a.ReinitWithKey(prk, common.CIPHER_SCHEME_CURVE); err != nil { ... }
  a.SetKnockUser("user@example.com", "org-1", nil)
  // now AddServer, AddResource, and knock as in section 2
  ```

  Don't knock until `ReinitWithKey` has run. A restart before it fails with
  `call ReinitWithKey, then restart` instead of coming back on a new random key.

## 4. Changing config at runtime

With no file watcher, changes go through method calls. All of these work on a
running agent:

| Change | Call |
| --- | --- |
| Add or replace a server | `AddServer(peer)` (same public key replaces the existing entry) |
| Remove a server | `RemoveServer(serverPubKeyBase64)` |
| Add or update a resource | `AddResource(res)` |
| Remove a resource | `RemoveResource(aspId, resId)` |
| Identity (user, org, user data) | `SetKnockUser(userId, orgId, userData)` |
| Private key and cipher scheme | `ReinitWithKey(privKeyBytes, cipherScheme)` |
| Device ID / host check results sent with knocks | `SetDeviceId`, `SetCheckResults` |

The knock loop picks up added and removed resources on its own.

**Restarts keep this state.** `Stop` followed by `Start` comes back with the
servers, resources, key and identity the agent last had, not the ones from the
original `Config`. Two things a restart does not bring back:

- **The periodic knock loop.** `Start` never starts it, so call
  `StartKnockLoop()` again after every restart if you use it.
- **Your log level, with `RestartAgent`.** `RestartAgent` (and the console's
  `/agent/restart`) always calls `Start` with log level 4 and starts only the
  DHP knock loop. With `UseInMemoryConfig`, `Config.LogLevel` still wins; with
  `SetConfigSource` the agent comes back at level 4. Prefer `Stop` + `Start`
  in SDK code.

To change something that has no setter (`LogLevel`, or `DHPConfig`), call
`Stop`, then `UseInMemoryConfig` with the new values, then `Start`. That
`Config` replaces the key and identity too, so include the current ones.
Servers and resources added earlier are kept. Never call `UseInMemoryConfig`
or `SetConfigSource` on a running agent: the mode is not guarded by a lock, and
switching it under a live agent is a data race.

## 5. What's not supported or behaves differently

- **File-backed operations are refused.** `RotateAgentKey` and `RotateTeeKey`
  return an `operation not supported in memory config mode` error, and the DHP
  console's `/config/server` endpoints return HTTP 501. To rotate the key,
  generate a new one and call `ReinitWithKey`.
- **Single-instance servers only.** `AddServer` creates a cluster with one
  instance. Multi-instance, load-balanced clusters (`[[Servers.Instances]]`)
  exist only in `server.toml`. Use `ServerPubKey`, not `Cluster`, in a
  `KnockResource`.
- **`dhp.toml` is not read.** Pass DHP settings in `Config.DHPConfig`.
- **Sealed (`v1$...`) keys work** in `Config.PrivateKeyBase64`. They are
  unsealed at `Start` with the passphrase from `NHP_KEY_PASSPHRASE_FILE` (a
  0600 file, recommended) or `NHP_KEY_PASSPHRASE`, the same as in file mode.
  `ReinitWithKey` takes raw key bytes, so a key installed with it is held
  unsealed.
- **What `workingDir` is still used for.** Logs go to `<workingDir>/logs`. The
  DHP web console, if you start it, keeps its own files under
  `<workingDir>/etc`: its TLS certificate in `etc/certs`, `console.toml`,
  `console.token`, and trusted applications in `etc/ta`. None of the four
  agent config files (`config.toml`, `dhp.toml`, `server.toml`,
  `resource.toml`) are read.
- **One agent per process.** Memory mode is not a way to run several agents
  side by side. The working directory and the logger are process-wide: each
  `Start` repoints both, so every agent's log lines go to the newest agent's
  file, and a failed `Start` leaves them pointed at a closed log. The config
  watchers are the exception: a memory-mode agent never touches them, so it
  doesn't turn off a file-mode agent's hot reload.

## 6. Key handling

In memory mode the agent's private key lives in your process memory and in
whatever code supplies it. The agent doesn't persist it anywhere, so storing it
securely is your job: use the platform keystore (iOS Keychain, Android
Keystore) or a secrets manager, never source code, logs or plain preference
files. `PrivateKeyBase64()` returns the key as configured (plain base64 unless
you supplied a sealed blob), so treat its output the same way. If you pass a
sealed `v1$` key, don't store its passphrase in the same place as the key.
