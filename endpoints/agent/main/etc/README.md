# nhp-agent `etc/` configuration

This directory holds the default configuration shipped with the
`nhp-agentd` binary. `make agentd` (see the root `Makefile`, target
`agentd`) builds `endpoints/agent/main/main.go` into
`release/nhp-agent/nhp-agentd` and copies every `*.toml` here into
`release/nhp-agent/etc/`. The `certs/` directory is copied too. Files
that are not `*.toml` (this README, for example) stay in source as
documentation and are NOT shipped in the release bundle.

## `server.toml` — the server-peer table

`server.toml` is the agent's table of nhp-server peers. Each
`[[Servers]]` block declares one logical server identity (one public
key) and 1..N physical instances that share it. The schema is shared
with nhp-ac and nhp-db (see `nhp/common/clusterconfig`), so a
`[[Servers]]` block can be copied between peer tables unchanged. See the
field reference in the header comment of `server.toml` itself for every
field.

Two forms are accepted:

- **Cluster form (recommended):** top-level `Name` / `PubKeyBase64` /
  `ExpireTime` / `LoadBalance` plus one or more `[[Servers.Instances]]`
  with `Host` / `Ip` / `Port` / `Weight`.
- **Legacy flat form:** `Hostname` / `Ip` / `Port` at the top level of
  `[[Servers]]`, no `[[Instances]]`. Auto-promoted to a single-instance
  cluster on load with a deprecation warning. Do not mix the two forms
  in the same entry.

## The bundled peer targets the local docker-compose environment

The `[[Servers]]` block shipped here points at `Ip = "127.0.0.1"`,
`Port = 62206`. That is the **local docker-compose** environment:
`docker/docker-compose.yaml` publishes the `nhp-server` service's UDP
port as `62206:62206/udp` to the host, so an agent binary running on the
host can knock the containerized server at `127.0.0.1:62206`.

The `PubKeyBase64` is the SM2 (gmsm) public key matching
`docker/nhp-server/etc/config.toml` (`DefaultCipherScheme = 1`). The
commented Curve25519 line in `server.toml` is the same key in curve
form — switch to it if you set the server's `DefaultCipherScheme = 0`.

### Quick start

```bash
# 1. Build the agent (writes release/nhp-agent/nhp-agentd + etc/):
make agentd

# 2. Start the matching docker-compose stack:
cd docker && docker-compose up -d nhp-server nhp-ac nhp-relay

# 3. Register a key (reads etc/server.toml for the peer table):
./release/nhp-agent/nhp-agentd register
```

## Retargeting a different server

Edit the `[[Servers]]` block: set `PubKeyBase64` to the target server's
public key, and the instance `Host`/`Ip` + `Port` to its NHP UDP
endpoint. Keep the key type consistent with the server's
`DefaultCipherScheme`:

- `DefaultCipherScheme = 0` → Curve25519 key (32-byte base64)
- `DefaultCipherScheme = 1` → SM2 key (64-byte base64)

## Public demo

The live demo runs two independent nhp-server clusters
(`server.opennhp.org` and `server2.opennhp.org`). The agent-side
`[[Servers]]` peer blocks for each demo cluster are shown as copyable
TOML on the registration page at <https://reg.opennhp.org/> — open the
"Demo nhp-server configuration" panel and pick a cluster.

Caveat: native `nhp-agentd` is **UDP-direct only** (no relay transport),
and the demo's `62206/udp` is **not exposed to the public internet** —
the public path is `443/TLS` through nginx + `nhp-relay`, which the
browser (js-agent) demo uses. The public-demo TOML is therefore
reference material unless your network can reach the server's UDP port
directly (e.g. from within the demo VPC). For local development, use the
bundled local-docker peer above.

### Key rotation — always re-fetch the demo peer

The demo nhp-server keys are **re-generated** whenever the
`deploy-demo-v2` workflow runs with `regenerate_keys=yes`, and may
otherwise change between deploys. A stale `PubKeyBase64` will fail peer
validation at the server (the agent's knock is rejected before the
Noise handshake completes).

So when pointing a native agent at the demo:

- **Fetch the `[[Servers]]` block fresh from <https://reg.opennhp.org/>**
  (the "Demo nhp-server configuration" panel) every time — the page reads
  the currently-deployed `config.json` with `cache: 'no-store'`, so it
  always reflects the live keys.
- **Do not cache an old copy** of the demo `server.toml`; an outdated
  pubkey will silently break knocking after the next rotation.
- The **local-docker peer** bundled here is unaffected — its key is tied
  to `docker/nhp-server/etc/config.toml`, not the demo deploy pipeline.
