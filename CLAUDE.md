# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OpenNHP is a Go-based Zero Trust security toolkit implementing two core protocols:
- **NHP (Network-infrastructure Hiding Protocol)**: Conceals server ports, IPs, and domains from unauthorized access
- **DHP (Data-content Hiding Protocol)**: Ensures data security via encryption and confidential computing

The system follows NIST Zero Trust Architecture with three core components that communicate via encrypted UDP packets using the Noise Protocol Framework.

## Git Commit Requirements

All commits must be signed with a verified GPG or SSH key. Unsigned commits will fail CI checks.

```bash
# Sign commits (if not configured globally)
git commit -S -m "your message"

# Amend to sign an existing commit
git commit --amend --no-edit -S
```

## Build Commands

```bash
# Full build (all components + SDKs + plugins + archive)
make

# Build individual components
make agentd      # Build nhp-agent daemon
make serverd     # Build nhp-server daemon
make acd         # Build nhp-ac (access controller) daemon
make db          # Build nhp-db daemon
make kgc         # Build nhp-kgc (key generation center)

# Build with eBPF support (requires clang)
make ebpf

# Build plugins
make plugins

# Initialize/tidy modules
make init
```

## Running Tests

```bash
# Run tests in the nhp module
cd nhp && go test ./...

# Run tests in the endpoints module
cd endpoints && go test ./...

# Run specific test file
cd nhp && go test -v ./test/packet_test.go

# Run benchmark tests
cd nhp && go test -bench=. ./core/benchmark/
```

## Code Formatting

**IMPORTANT**: All Go code must be properly formatted before committing. CI will fail if formatting is incorrect.

### Before Committing

Always run these commands on modified Go files:

```bash
# Format code with gofmt
gofmt -w <file.go>

# Fix import grouping with goimports
goimports -w <file.go>

# Or format all files in a directory
gofmt -w ./path/to/package/
goimports -w ./path/to/package/
```

### Import Grouping Style

Imports must be organized into three groups separated by blank lines:

1. Standard library imports
2. External third-party imports
3. Internal project imports

```go
import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/log"
)
```

### Verify Formatting

Check if files need formatting (no output means properly formatted):

```bash
gofmt -l <file.go>
goimports -l <file.go>
```

### Install goimports

If `goimports` is not installed:

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

## Docker Development

```bash
# Build and run the full stack
cd docker && docker-compose up --build

# Individual service testing
docker-compose up nhp-server
docker-compose up nhp-ac
docker-compose up nhp-agent
```

## Architecture

### Module Structure

The codebase uses two separate Go modules with a local replace directive:

- **`nhp/`**: Core protocol library
  - `core/`: Packet handling, cryptography, device management, Noise Protocol implementation
  - `common/`: Shared types and message definitions (AgentKnockMsg, ServerKnockAckMsg, etc.)
  - `utils/`: Utility functions
  - `plugins/`: Plugin handler interfaces (PluginHandler interface)
  - `log/`: Logging infrastructure
  - `etcd/`: Distributed configuration support

- **`endpoints/`**: Daemon implementations (depends on nhp module)
  - `agent/`: NHP-Agent - client that sends knock requests
  - `server/`: NHP-Server - authenticates and authorizes requests
  - `ac/`: NHP-AC - access controller that manages firewall rules
  - `db/`: NHP-DB - Data Broker for DHP
  - `kgc/`: Key Generation Center for IBC (Identity-Based Cryptography)
  - `relay/`: TCP relay functionality

### Core Concepts

**Device Types** (defined in `nhp/core/device.go`):
- `NHP_AGENT`: Client initiating access requests
- `NHP_SERVER`: Central authentication/authorization server
- `NHP_AC`: Access controller managing network rules
- `NHP_DB`: Data Broker for DHP
- `NHP_RELAY`: Packet relay

**Packet Types** (defined in `nhp/core/packet.go`):
- `NHP_KNK`: Agent knock request
- `NHP_ACK`: Server knock acknowledgment
- `NHP_AOP`: Server-to-AC operation request
- `NHP_ART`: AC operation result
- `NHP_REG`/`NHP_RAK`: Agent registration flow
- `DHP_*`: Data Hiding Protocol messages

**Cipher Schemes** (in `nhp/core/crypto.go`):
- `CIPHER_SCHEME_CURVE`: Curve25519 + AES-256-GCM + BLAKE2s
- `CIPHER_SCHEME_GMSM`: SM2 + SM4-GCM + SM3 (Chinese national standards)

### Configuration

All daemons use TOML configuration files in their respective `etc/` directories:
- `config.toml`: Base configuration (private key, listen address, log level)
- `server.toml`: Remote server/peer definitions
- `resource.toml`: Protected resources and auth service providers
- `http.toml`: HTTP server settings (for nhp-server)

### Plugin System

Server plugins implement the `PluginHandler` interface (`nhp/plugins/serverpluginhandler.go`) and are built as Go plugins (`.so` files). See `examples/server_plugin/` for reference implementation.

Key plugin methods:
- `AuthWithNHP()`: Handle NHP protocol authentication
- `AuthWithHttp()`: Handle HTTP-based authentication
- `RegisterAgent()`: Agent registration
- `ListService()`: Service discovery

### Key Generation

All daemons support the `keygen` command:
```bash
./nhp-serverd keygen --curve  # Generate Curve25519 keys
./nhp-serverd keygen --sm2    # Generate SM2 keys (default)
```

## Demo Deployment (AWS)

The `terraform/demo/` stack provisions the public demo (nhp-server, nhp-ac,
nhp-relay + nginx + Let's Encrypt) in `us-east-2` on the OpenNHP demo AWS
account. The state bucket is configured at `terraform init` time via
`-backend-config="bucket=$TF_STATE_BUCKET"` (workflows read the
`TF_STATE_BUCKET` repo variable) so the account ID is not committed in source.
All secrets live in a single AWS Secrets Manager secret: **`opennhp/demo`**.

> **The nhp-server has no public HTTP/HTTPS surface.** It used to serve a demo
> login page (`/plugins/example?action=login&resid=demo`) on 443 via nginx,
> reachable as `auth-plugin.opennhp.org` and the legacy alias
> `demologin.opennhp.org`. That surface is retired at three layers: the
> `tcp/443` ingress is gone from `aws_security_group.server`
> (`terraform/demo/security-groups.tf`), the `demologin` CNAME is gone from
> `terraform/demo/dns.tf`, and `deploy/config-templates/server/http.toml` sets
> `EnableHttp = false` so `nhp-serverd` never binds `127.0.0.1:8443`. The
> `deploy-server` job tears down any leftover `/etc/nginx/conf.d/server.conf`.
> The server host is reachable only over the NHP UDP knock port (plus SSH from
> the relay jump host); `deploy/nginx/server.conf.template` is kept for
> reference but is no longer deployed. Re-enabling the login page requires
> undoing all of these together.

#### nhp-server eBPF/XDP ingress filter

`nhp-serverd` attaches `nhp/ebpf/xdp/nhp_server_xdp.c` at startup and enforces
that same shape at the driver: UDP on the knock port (≥240 bytes, the curve
`NHP_KPL` header size — gmsm's 304 clears it too), SSH only from the relay's
addresses, everything else dropped. It parses no NHP protocol; identity is still
decided by the Noise handshake in user space. The AC's per-knock XDP program
and the shared loader (`nhp/utils/ebpf/engine_linux.go`) are the same code
path selected by `EngineLoadParams.Variant`.

**"Everything else" includes IPv6 and fragments**, and neither was true at
first. IPv6 was one `case ETH_P_IPV6: return XDP_PASS` justified by the demo
hosts having no v6 service — but sshd binds `[::]:22` by default, so the day a
v6 CIDR reaches the VPC every port would have been open over v6 with the
deploy's "filter attached" check still green. `handle_ipv6()` now runs the same
tree minus the whitelist (the LPM trie holds IPv4 prefixes, so no v6 source can
be recognised as the relay): ICMPv6 ND/MLD and "packet too big" pass so the host
stays on its subnet, DHCPv6 (547 → 546) and NTP pass for the same reasons as
their v4 spellings, replies to the host's own v6 flows pass via
`has_local_flow6()`, everything else is `V6_OTHER`. **SSH must therefore reach
the host over IPv4**; `warnOnGlobalIPv6` logs a `Warning` naming the address
when the filtered interface has a global v6 address, and `deploy-server`'s
`$SSH_CONNECTION` containment gate fails outright for a v6 peer address since no
IPv4 entry can contain it. Non-first IPv4 fragments have no L4 header, and the
TCP/UDP branches used to read one anyway at `ihl * 4`: every datagram over the
path MTU lost its tail to `UDP_OTHER` and died in reassembly, while a crafted
fragment whose payload read `67 → 68` took the DHCP exception. They are now
split out before the protocol dispatch and passed as `IPV4_FRAGMENT` — inert
without the first fragment, which carries the ports and takes the full tree. An
IPv6 fragment header with a non-zero offset gets the same answer
(`IPV6_FRAGMENT`).

The event record (`struct nhp_event_t`) carries an address family and both
address pairs, so a v6 drop logs its real source; `serverEventByteSize` in the
loader is `sizeof` that struct and has to move with it.

**Attaching is opt-in, and `etc/xdp.toml` is the opt-in.** This filter is not a
hardening tweak that is safe to turn on everywhere: it drops every inbound TCP
flow the host did not initiate, including SYNs to services this daemon listens
on, and on a host reachable only through the whitelist one bad list makes it
unreachable for good. So `loadXdpConfig` (`endpoints/server/config.go`) attaches
only when the file exists, parses, has `Enabled = true` **and** a `RelayIPs`
whose every entry parses as an IPv4 address or prefix (see the whitelist gates
below) — a server with no `xdp.toml` keeps exactly the exposure it had
before the filter existed, even though `make ebpf` puts the object in
`release/nhp-server/etc/` and many operators run the daemon as root. It also
refuses when the HTTP knock listener or an off-host metrics endpoint is enabled
(the shipped `endpoints/server/main/etc/http.toml` has `EnableHttp = true` on
443), because attaching in front of a listener would take that service off the
air with nothing in user space to say so. Loopback binds are not a conflict —
the program is on the default-route interface.

`NhpMinFrameBytes` and the knock port are *applied*, not documentation: the
loader rewrites `nhp_min_udp_len` and `nhp_listen_port` in the object's
`.rodata` before the verifier runs (`setServerConstants`). The port comes from
the daemon's own `ListenPort`, never from the TOML, so the filter and the
listener cannot disagree — a hard-coded 62206 would have dropped every knock on
any server configured elsewhere. An object too old to carry those constants is
refused rather than attached with its own defaults.

**The filter also has to let the host's own replies back in**, and this was
missed at first: the daemon is a client too (DNS, the auth plugin's SMTP
submission to SES, package mirrors), and dropping those answers broke OTP
email with a DNS `i/o timeout` while every knock still worked. `has_local_flow()`
answers it with `bpf_sk_lookup_{tcp,udp}` against the kernel's socket table —
an established TCP socket or a *connected* UDP socket admits the packet, a
listening socket never does — so the server needs none of the AC's
conn_track/TC-egress machinery and keeps no state that can drift. ICMP
fragmentation-needed (type 3 code 4) passes for the same reason; the rest of
ICMP stays dropped. `deploy-server` probes udp/53 and tcp/587 from the host
after every restart so this cannot regress silently again.

**Two clients the socket table cannot vouch for are passed by port pair, and
DHCP is the one that matters.** `has_local_flow()` only admits *connected*
sockets, and a DHCP client receives its lease on a raw `AF_PACKET` socket or on
one merely bound to udp/68 — so the renewal answer (udp/67 → udp/68) was
dropped. EC2 leases the primary private IPv4 for a finite time and
systemd-networkd renews it at half the lease; with the answer dropped the lease
expired, networkd removed the address, and roughly an hour after a deploy that
had verified perfectly the host went dark on **every** port at once — SSH, the
knock port, and with them every recovery path except a reboot. The knock port
staying open is what makes this look like "the filter closed the whole host"
rather than a filter bug. chronyd has the same shape one layer up (its poll
answer comes back to its own udp/123) and costs the clock instead of the
address, so it is passed too; nothing else is, because the alternative — admit
anything addressed to a bound port — is the "no unsolicited packet gets in"
property thrown away. The AC's program has always had the DHCP short-circuit
(`DHCP_PORT_R || DHCP_PORT_O`); the server's fork dropped it.

`deploy-server`'s `Verify the host can still renew its DHCP lease` step is the
standing gate: it runs `networkctl renew` and fails unless a new
`REASON=DHCP_CLIENT` line shows the answer crossed the filter. No other probe
can see this, because everything works for as long as the lease lasts. The
loader's second line of defence is a watchdog in
`endpoints/server/ebpf/serverengine.go`: if the filtered interface has no
routable IPv4 address for two minutes it logs `Critical` and detaches, so the
next DHCP exchange can restore the host instead of needing volume surgery. It
can only fire in a state where the host answers nothing anyway.

**The event log is bounded on purpose**, because every line in it is a packet
somebody else chose to send and `nhp/log` rotates by date without ever pruning:
letting it grow hands a scanner the ability to fill the root volume and kill
the daemon the filter protects. Three limits, in `record_packet()` and
`reportServerStats`: a per-action per-CPU token bucket (`NHP_EVENT_BURST`);
bulk classes that are counted but never written per packet — `TCP_ESTABLISHED`,
`UDP_ESTABLISHED` and every `SSH_RELAY` packet after the SYN, whose rate is the
host's throughput rather than its event rate; and a `[NHP-STAT]` summary line
per action per minute, fed by the `nhp_action_stats` counters that tick for
*every* packet, so the totals stay exact however much reporting was dropped.
`pruneServerEventLogs` then caps retention at 14 days / 256 MiB, oldest first,
never today's file.

The whitelist lives in `etc/xdp.toml` (`deploy/config-templates/server/xdp.toml`),
rendered from `$RELAY_IPS` (comma-separated; `scripts/generate-nhp-keys.sh`
quotes it into the TOML array) and hot-reloaded. Entries are host addresses or
CIDR prefixes, matched by longest prefix — `nhp_relay_ips` is an LPM trie. CI
renders three, and each is load-bearing:

- the relay's VPC **private** address, which is what SSH arrives from because CI
  jumps through the relay to the server's private address;
- the relay's **public** address, for traffic that reaches the host through the
  internet gateway. Whitelisting only the public address closes tcp/22 the moment
  the daemon restarts;
- **the VPC subnet**, because `aws_instance.relay` has no pinned `private_ip`.
  Replacing the relay for any reason (AMI or instance-type change, taint, AZ
  move) brings it back with a different private address. The server's live
  whitelist would still name the old one, SSH from the new relay would be
  dropped, and `deploy-server` — which reaches the server *through* the relay —
  could never push the correction. The dead-man watchdog does not fire either:
  the host still has its address. That is a root-volume-detach recovery, and the
  subnet prefix is what prevents it. The cost is that the AC and server share
  that subnet and are covered too.

**There is no break-glass SSH path**: if the list is rendered empty or without a
covering entry and the daemon restarts, the only way back in is to detach the
root volume or rebuild the instance. Several gates protect against that, all
before anything is scp'd or restarted:

- the `configure` job reads `relay_private_ip` and `subnet_cidr` from Terraform
  and resolves `relay.opennhp.org`, failing the whole run unless it gets an IPv4
  address, an IPv4 prefix, and exactly one A record, and unless the prefix
  *contains* the relay's private address — a prefix the relay is not in would
  not cover its replacement either. `terraform output` only sees outputs an
  apply has written into the state, so `subnet_cidr` falls back to
  `aws_subnet.public.cidr_block` read out of the state when the output is not
  there yet (this job never applies; without the fallback, adding an output
  leaves every deploy red until someone runs `infra-demo` by hand);
- `deploy-server`'s `Check this job's own SSH source is in the XDP whitelist`
  step asks the host what peer address it sees for the live connection
  (`$SSH_CONNECTION`) and fails unless some entry in the rendered `xdp.toml`
  *contains* it (a containment test, since entries may be prefixes);
- `startXdpFilter` refuses to attach at all for a file whose `RelayIPs` does not
  name at least one usable prefix, and `applyXdpConfig` refuses to apply one on
  reload and keeps the active whitelist — an unset `RELAY_IPS`, or a file caught
  half-written by the watcher, parses to valid TOML with an empty list, and
  applying that closes tcp/22 for every source. "Usable" is decided by
  `ParseRelayPrefixes` (`nhp/utils/ebpf/relayips.go`), not by counting strings:
  `["relay.opennhp.org"]`, `[""]`, an IPv6-only list or a typo like
  `["10.0.1.300"]` is a non-empty *list* that names not one address the LPM trie
  can hold, so a count-the-strings guard would attach with an empty map, or let
  a reload sweep a working whitelist away. One bad entry fails the whole list
  rather than being skipped — the entry that did not parse may be the one SSH
  arrives from, and nothing in user space can tell. An IPv4-mapped IPv6 prefix
  (`::ffff:10.0.1.0/120`) is refused for the same reason and is the one spelling
  that gets past a check written against `net`: `To4()` is non-nil for it while
  `Mask.Size()` is 120, a prefix length above the trie's `max_prefixlen` of 32
  that the kernel answers with `EINVAL` — an entry validation called usable and
  the map will not take. `ReplaceRelayIPs` repeats both refusals before its first
  map write, so the contract does not depend on its callers.
- `ReplaceRelayIPs` is also all-or-nothing against the *map*, not only against
  the file: it reads the live whitelist first, and a failed `Update` (ENOMEM on
  the no-prealloc trie, a key the kernel will not hold) rolls back the keys that
  call added and skips the stale sweep, so the map is left holding exactly the
  list it held before. Sweeping anyway is how a reload that merely failed to add
  the relay's *new* address removes its *old* one too — a map holding neither,
  i.e. tcp/22 closed to every source, while `applyXdpConfig` logs that it kept
  the active whitelist. A map it cannot even read is one it refuses to touch.

A red run with the demo still up is the intended outcome of any of them.

**The whitelist is installed before the program is attached**, not after: it
travels with the load as `EngineLoadParams.RelayIPs`, and `loadServerEngine`
writes it into `nhp_relay_ips` inside the deferred-cleanup region, before
`link.AttachXDP`. Writing it afterwards — which is what the first version did —
leaves two holes that no gate above can see, because both are on the host rather
than in the pipeline: a window after every restart in which SSH from the relay
is dropped, and, if that one write fails (an LPM-trie `Update` error, ENOMEM),
a filter left attached enforcing an *empty* whitelist, i.e. tcp/22 closed to
every source with no break-glass path — at the one moment user space has just
proved it cannot write the map it would need to fix. As a load parameter a bad
whitelist is an ordinary pre-attach error: nothing is attached, the pins are
swept, and the daemon runs fail-open. Startup therefore does not call
`applyXdpConfig` at all; that path exists for the config watcher, which always
has a live filter and a working list to fall back on.

If the host is already unreachable, **reboot the instance before anything
else** (`aws ec2 reboot-instances`, no SSH needed): DHCP runs before
`nhp-serverd` starts, so the address comes back and tcp/22 is open for the
window before the daemon attaches the filter again. Use that window to either
`systemctl stop nhp-serverd` (filter detached, host stays reachable) or re-run
the deploy. Volume surgery is only needed when the *whitelist* is wrong, i.e.
when the filter is attached and the host does have its address.

Loading is **fail-open**: a missing object, an old kernel or a missing
capability leaves the daemon running with no filter and a `Warning` in the log.
The `Verify the XDP ingress filter attached` step in `deploy-server` is what
turns that silent loss of protection into a failed deploy. Fail-open only means
anything if it is true, so `loadServerEngine` does every fallible step *before*
`link.AttachXDP` and detaches on any error after it: a filter left attached
behind a returned error would be enforcing with an empty SSH whitelist that
user space no longer holds a handle to fix, while the log says "fail-open". Capabilities
(`CAP_BPF CAP_NET_ADMIN CAP_PERFMON`) plus a root `ExecStartPre` that prepares
`/sys/fs/bpf` come from `terraform/demo/userdata/server.sh` on new hosts and
from a systemd drop-in installed by `deploy-server` on existing ones — keep the
two in sync.

Verification procedure (netns rehearsal, pre-deploy baseline, the same
commands after the deploy, recovery paths):
`terraform/demo/VERIFY-server-xdp.zh-cn.md`.

### `opennhp/demo` schema

The secret is JSON; fields are added idempotently by scripts and workflows.
Missing fields are auto-generated on the next `scripts/generate-nhp-keys.sh`
run (triggered by the `deploy-demo-v2` workflow).

| Field | Populated by | Used by |
| --- | --- | --- |
| `nhp_server_private_key` / `_public_key` | `scripts/generate-nhp-keys.sh` | server `config.toml`; peer tables on ac/relay |
| `nhp_ac_private_key` / `_public_key` | same | ac `config.toml`; peer table on server |
| `nhp_relay_private_key` / `_public_key` | same | relay `config.toml`; peer table on server |
| `nhp_agent_private_key` / `_public_key` | same | native nhp-agent clients; `agent.toml` on server |
| `nhp_jsagent_private_key` / `_public_key` | same | cluster 1 `endpoints/js-agent/` demo identity (rendered into `config.json` `clusters[0]` at deploy time); trusted by server cluster 1 only |
| `nhp_jsagent_sm2_public_key` | same (derived via `--both`) | SM2 peer entry in `server/agent.toml`; lets cluster 1 js-agent knock in gmsm mode |
| `demoapp_key_envelope_key` | `scripts/generate-nhp-keys.sh` (idempotent; generated via `openssl rand -base64 32` on first run) | demoapp `config.toml` `KeyEnvelopeKey`; AES-256 master that wraps each user's NHP private key at rest |
| `demoapp_session_key` | same | demoapp `config.toml` `SessionKey`; signs the session cookie |
| `cloudflare_api_token` | manually provisioned once | Terraform + certbot DNS-01 (`Zone:DNS:Edit` + `Zone:Zone:Read`) |
| `cloudflare_zone_id` | same | Terraform DNS records for `opennhp.org` |
| `stealth_ca_cert` | `infra-demo` workflow (from GitHub Secrets `STEALTH_CA_CERT`) | `tls_locally_signed_cert.demo_nhp` |
| `stealth_ca_key` | `infra-demo` workflow (from GitHub Secrets `STEALTH_CA_KEY`) | `tls_locally_signed_cert.demo_nhp` |
| `ssh_deploy_private_key` | manually bootstrapped (see `terraform/demo/RUNBOOK.md`); never enters Terraform state | CI SSH into EC2 hosts |
| `ssh_deploy_public_key` | derived in CI via `ssh-keygen -y` and passed as `TF_VAR_deploy_public_key` | `aws_key_pair.deploy` → `ec2-user` authorized keys |
| `ssh_host_keys` | `infra-demo` workflow on `apply` | CI `known_hosts` for strict host key checking |

> Cluster 2 has been decommissioned. The corresponding fields
> (`nhp_jsagent2_*`, `nhp_server2_*`, `nhp_ac2_*`) used to be listed here
> alongside `nhp_jsagent_sm2_public_key`. Legacy values may still be
> present in `opennhp/demo` for backwards compatibility but the
> deploy pipeline no longer reads or writes these keys, no Terraform
> resources reference them, and no config templates render them. Safe to
> delete from `opennhp/demo` if desired (does not affect any running
> cluster 1 host); the deploy script tolerates the field being absent.

### Key-generation flow

`scripts/generate-nhp-keys.sh`:

1. Reads existing values from `opennhp/demo`.
2. Uses each daemon's `keygen --curve --json` to fill any missing pair.
3. Writes the merged object back to `opennhp/demo` (preserving unrelated fields).
4. Renders `deploy/config-templates/` via `envsubst` into `deploy/configs/` for
   scp to the hosts.

Pass `--regenerate` to the script (or `regenerate_keys=yes` on the workflow) to
force a full rotation. This breaks every registered agent/ac/relay until their
peer tables are redeployed in lockstep, so use sparingly.

## Protocol Flow

1. Agent sends encrypted knock (`NHP_KNK`) to Server
2. Server validates, sends operation request (`NHP_AOP`) to AC
3. AC opens firewall, responds (`NHP_ART`) to Server
4. Server sends acknowledgment (`NHP_ACK`) with access info to Agent
5. Agent can now access the protected resource through AC
