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
the same shape at the driver: UDP on the knock port (≥ `NhpMinFrameBytes`), SSH
only from the addresses in `etc/xdp.toml`, the replies to flows the host itself
opened, and nothing else — IPv6 and non-first fragments included. It parses no
NHP protocol; identity is still decided by the Noise handshake in user space.
The AC's per-knock program and the shared loader
(`nhp/utils/ebpf/engine_linux.go`) are the same code path, selected by
`EngineLoadParams.Variant`.

Invariants. Each one is load-bearing and most were paid for once already; the
reasoning, and the incident behind each, is in
[terraform/demo/RUNBOOK.md](terraform/demo/RUNBOOK.md) under *nhp-server ingress
filter (XDP)*.

- **Attaching is opt-in; `etc/xdp.toml` is the opt-in.** No file, an unparsable
  one, `Enabled = false`, or a `RelayIPs` whose entries do not all parse as IPv4
  addresses or prefixes ⇒ nothing is attached and the host keeps the exposure it
  had before the filter existed. `loadXdpConfig`/`startXdpFilter` in
  `endpoints/server/config.go` also refuse in front of a non-loopback HTTP or
  metrics listener, and `updateHttpConfig` refuses the mirror case — starting
  such a listener while the filter is attached.
- **The whitelist must cover the SSH source, and there is no break-glass path.**
  A list rendered empty, or without an entry containing the address the host
  sees SSH arrive from, means detaching the root volume or rebuilding the
  instance. Entries are host addresses or CIDR prefixes in an LPM trie; one
  unparsable entry fails the whole list rather than being skipped.
- **The whitelist is installed before the program is attached**, as
  `EngineLoadParams.RelayIPs`, never written afterwards. Startup therefore never
  calls `applyXdpConfig`; that path is the config watcher's, and it always has a
  live filter and a working list to fall back on.
- **A reload never weakens a working whitelist.** `applyXdpConfig` keeps the
  active list on any refusal, and `ReplaceRelayIPs` is all-or-nothing against
  the map as well as the file: it adds before it sweeps, rolls back its own adds
  on a failed `Update`, and refuses a map it cannot read.
- **Loading is fail-open**, and `deploy-server`'s `Verify the XDP ingress filter
  attached` is what stops that from being a silent loss of protection. Every
  fallible step happens before `link.AttachXDP`, and any error after it detaches.
- **The filter must let the host's own traffic back in.** Replies to the
  daemon's own flows via `bpf_sk_lookup_{tcp,udp}` (connected sockets only,
  never listening ones); DHCP (67→68) and NTP by port pair, because their
  clients have no connected socket. Dropping the DHCP answer costs the host its
  address and every port with it, about an hour after a deploy that verified
  clean.
- **The capabilities go back on every path, including the ones that never
  load.** `dropLoaderPrivileges` (`nhp/utils/ebpf/caps_linux.go`) clears the
  ambient set and empties permitted and effective — keeping CAP_BPF only where
  `kernel.unprivileged_bpf_disabled` makes bpf(2) privileged and the reload path
  still needs to write the map — process-wide, not just on the calling thread.
  The unit grants all three as *ambient* capabilities before any of this code
  runs, so a host that declines to filter is holding them too: `loadXdpConfig`
  therefore defers `DropLoaderPrivileges` across all of its returns, not just
  the loader's. A load that *failed*, and every refusal that never reached the
  loader, keeps nothing, CAP_BPF included — there is no map to reload, and that
  is the state the daemon stays in for the life of the process. The drop is
  once-guarded and the loader goes first, so the attach path's answer is the one
  that decides CAP_BPF.
- **The event log is bounded**, because every line in it is a packet someone
  else chose to send: per-class and global token buckets in the C, bulk classes
  counted but never written per packet, a daily byte budget in the writer, exact
  `[NHP-STAT]` totals regardless, and 14-day / 256 MiB retention.
- **Recovery starts with a reboot** (`aws ec2 reboot-instances`, no SSH needed):
  DHCP runs before `nhp-serverd`, so tcp/22 is open for the window before the
  filter is re-attached. Volume surgery is only for a wrong *whitelist* on a
  host that does have its address.

The gates run before anything is scp'd or restarted — the `configure` job's
Terraform/DNS resolution and containment check, `deploy-server`'s
`Check this job's own SSH source is in the XDP whitelist`, and the refusals in
`startXdpFilter`, `applyXdpConfig` and `ParseRelayPrefixes`
(`nhp/utils/ebpf/relayips.go`). A red run with the demo still up is the intended
outcome of any of them.

Capabilities (`CAP_BPF CAP_NET_ADMIN CAP_PERFMON`) plus a root `ExecStartPre`
that prepares `/sys/fs/bpf` come from `terraform/demo/userdata/server.sh` on new
hosts and from a systemd drop-in installed by `deploy-server` on existing
ones — keep the two in sync.

Verification procedure (netns rehearsal, pre-deploy baseline, the same commands
after the deploy, recovery paths):
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
