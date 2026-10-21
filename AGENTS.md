# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

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
> The server host *serves* only the NHP UDP knock port (plus SSH from the relay
> jump host) — the security group admits more than that, and the XDP filter drops
> the rest (see *All-ports exposure* below); `deploy/nginx/server.conf.template` is kept for
> reference but is no longer deployed. Re-enabling the login page requires
> undoing all of these together.

**NHP knock port.** The UDP port `nhp-serverd` listens on for knocks is the
single source of truth `var.nhp_listen_port` in `terraform/demo/variables.tf`.
`terraform/demo/security-groups.tf` derives both `aws_security_group.server`
and `aws_security_group.ac` from it, and the `deploy-demo-v2` `configure` job
reads it via `terraform output` and renders it into `server/config.toml`,
`ac/server.toml`, and `relay/config.toml` via `${NHP_SERVER_PORT}`. To change
the port: edit `var.nhp_listen_port`, run `infra-demo apply` first (so the
security group rule is in place), then run `deploy-demo-v2` — the order
matters, because the rendered config tracks the applied state, never the
source-default.

**All-ports exposure (deliberate).** `aws_security_group.server` is opened to
`0.0.0.0/0` on every tcp port **except 22**, every udp port and icmp while
`var.server_open_all_ports` is `true` (opt-in, default `false`; set it through the `infra-demo` `open_all_ports`
input once the attach gate is green, and `infra-demo` refuses it in a plan that
creates or replaces `aws_instance.server`). That is intentional and is
for observation: scans must reach the NIC so the XDP filter can count them
(`[NHP-STAT]`, `[NHP-SCAN]`). In this mode **the XDP filter is the only inbound
control on everything but tcp/22** — the security group is no longer the
narrower layer — and it is fail-open and its link is not pinned, so the filter
is detached on *every* `deploy-demo-v2` run (the daemon is stopped and
restarted), failed deploy, crash, `systemctl restart`, reboot and failed load,
not only on a new or replaced instance. During each of those windows the host
is exposed to the internet on every udp port, every tcp port except 22, and
icmp: chronyd, the SSM/cloud-init agents and anything else listening. **tcp/22
is deliberately carved out of the open range** and stays restricted to the relay
SG at the security-group layer, so sshd never depends on a fail-open filter;
the cost is that tcp/22 scans never reach the NIC and are not counted. Other IP
protocols (GRE, ESP, ...) are not opened. What the filter still lets through to the host from any source in this
mode: the knock port, SSH from the whitelist, UDP 67→68 (DHCP client), UDP
123→123 (chronyd), non-first IPv4 fragments and ICMP frag-needed. Those are
accepted exposure, not narrowed. Order is the *reverse* of the knock-port rule: run `deploy-demo-v2` first
and confirm `Verify the XDP ingress filter attached` is green, then
`infra-demo apply`. Rollback is `server_open_all_ports = false` plus
`infra-demo apply`, which returns the security group exactly to udp/knock-port +
tcp/22 from the relay SG; recovery of an unfiltered host is still
`aws ec2 reboot-instances`. `aws_instance.{server,ac,relay}` carry
`ignore_changes = [ami]` so a new al2023 release never plans a replace; change
an AMI deliberately with `terraform apply -replace=aws_instance.<x>`, one host
at a time (see `terraform/demo/RUNBOOK.md`).

**Binding a privileged port from `User=ec2-user`.** Because `nhp-serverd` runs
unprivileged on the demo host and `443 < 1024`, the unit does not rely on
`CAP_NET_BIND_SERVICE` — granting it would leave the daemon with a non-empty
ambient cap set and break the `Verify the XDP loader gave its capabilities
back` gate (`CapAmb` must be 0, `CapPrm/CapEff` must be `v & ~BPF == 0`).
Instead, the drop-in adds
`ExecStartPre=+/sbin/sysctl -q -w net.ipv4.ip_unprivileged_port_start=443`
and writes `/etc/sysctl.d/99-nhp-server.conf` for persistence; the same line
is in `terraform/demo/userdata/server.sh` for new hosts, and the two must
stay in sync (see the `ignore_changes = [user_data]` note in `CLAUDE.md`).
The XDP filter allows only the knock port and whitelisted SSH, so widening
the unprivileged-port range to 443 does not broaden the host's exposure
beyond what the filter already permits.

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
| `stats_collector_ssh_private_key` | `scripts/generate-nhp-keys.sh` (idempotent; ed25519 via `ssh-keygen -t ed25519 -N ""`) | relay host's `~/.ssh/stats_collector_key` (installed by `deploy-xdpstats`); used by `xdpstats-collect.sh` to pull from nhp-server |
| `stats_collector_ssh_public_key` | same | nhp-server's `~ec2-user/.ssh/authorized_keys` (appended by `deploy-server` with `command=…/nhp-xdplog-export.sh`,no-*); verified to match the private half before install |

> Cluster 2 has been decommissioned. The corresponding fields
> (`nhp_jsagent2_*`, `nhp_server2_*`, `nhp_ac2_*`) used to be listed here
> alongside `nhp_jsagent_sm2_public_key`. Legacy values may still be
> present in `opennhp/demo` for backwards compatibility but the
> deploy pipeline no longer reads or writes these keys, no Terraform
> resources reference them, and no config templates render them. Safe
> to delete from `opennhp/demo` if desired (does not affect any
> running cluster 1 host); the deploy script tolerates the field being
> absent.

#### XDP attack stats site (stats.opennhp.org)

Daily summaries of the nhp-server XDP ingress filter's `[NHP-DROP]` events are
published at `https://stats.opennhp.org/`. Static HTML + JSON, no third-party
resources, fully offline-built from the same
`logs/nhp_server_xdp-YYYY-MM-DD.log` files the daemon writes. The homepage
leads with a hero total of all-time blocked packets (plus a `Since` date) that
`countup.js` counts up; that script is generated into the web root by
`xdplog-report site`, loaded same-origin, and is the only one the CSP permits
(`script-src 'self'`). It is progressive enhancement — the number is already
in the HTML for readers without JS, without that CSP, or with reduced motion.

**Data flow**

- **Server (writer)**: `nhp-serverd` keeps writing the daily log as it always
  has. Same 0600 mode, same `[NHP-DROP]` / `[NHP-STAT]` / `[NHP-SCAN]` lines.
- **Server (exporter)**: `deploy/scripts/nhp-xdplog-export.sh` is the SSH
  *forced command* attached to the `stats_collector_ssh_public_key` entry in
  `/home/ec2-user/.ssh/authorized_keys` (the path sshd reads, not
  `~/nhp-server/.ssh/...`). It accepts only `list` and `get <YYYY-MM-DD>`;
  anything else (shell, scp, sftp, paths with `/`, malformed dates) is
  rejected at the script. The `#` marker goes on its own line above the entry
  — sshd treats a line *starting* with `#` as a comment.
- **Relay (collector)**: a systemd timer (`deploy/scripts/xdpstats.timer`,
  `OnCalendar=*-*-* 00:20:00 UTC`, `Persistent=true`) runs
  `xdpstats-collect.sh`, which calls `ssh -i ~/.ssh/stats_collector_key -o
  IdentitiesOnly=yes ... list` then `... get <date>` per day, gzip-stores
  them under
  `/home/ec2-user/xdpstats/raw/server/nhp_server_xdp-<date>.log.gz`, then
  invokes `xdplog-report site`. Host keys are verified against a known_hosts
  seeded from `ssh_host_keys`; never TOFU.
- **Site**: `endpoints/xdplog-report` has a `site` subcommand that emits
  `index.html` + `index.json` plus one `<day>.html` / `<day>.json` per day
  under `/var/www/stats/`. nginx serves that directory at
  `https://stats.opennhp.org/` using the same Let's Encrypt cert lineage as
  `relay.*` / `agent.*` / `reg.*` / `demo.*`.

**Three directories, three permission stories.** The web root is
`/var/www/stats` (0755), not under `$HOME`: `/home/ec2-user` is 0700 on
AL2023 and nginx runs as `nginx`. `raw/` stays 0700 under `$HOME` — it is
the only place full, unmasked source IPs exist on the relay. `state.json` is
written outside the web root via `-state`, and the vhost refuses
`/state.json` as a second line of defence.

**What's deliberately NOT changed**

- `xdp.toml`: untouched. Relay addresses were already in the SSH whitelist
  before the stats site existed.
- The XDP filter, capability grants, `aws_security_group.server`, and the
  `aws_instance.server` lifecycle are all untouched. The collector is read-only.
- `deploy-demo-v2` job order. `deploy-xdpstats` chains AFTER `deploy-demoapp`
  so certbot's `deploy-relay → deploy-jsagent → deploy-demoapp → deploy-xdpstats`
  serialisation keeps the relay host's exclusive certbot lock sequential.
  `deploy-server` appends the stats collector public key to `authorized_keys`
  BEFORE `deploy-xdpstats` runs.

**Deploy order: `infra-demo apply` first.** The `stats.opennhp.org` A record
is `cloudflare_record.stats` in `terraform/demo/dns.tf`, and `deploy-demo-v2`
only reads Terraform state, so the release that introduces the site has to be
`infra-demo action=apply` → `deploy-demo-v2 confirm=deploy`.
`deploy-xdpstats`'s first step requires `stats.opennhp.org` to resolve to
exactly one A record equal to `aws_eip.relay.public_ip`, so a forgotten apply
fails before the relay is touched. DNS-01 cert issuance would not catch it:
it writes a TXT record through the Cloudflare API and never looks up the A
record.

**Default public behaviour**

- Source IPs are masked: IPv4 truncated to `a.b.c.x`, IPv6 to the first 48 bits.
  `site` mode of `xdplog-report` defaults `-mask-ip` to `true` (different from
  the single-file CLI, which defaults `false`). The raw log cache on the relay
  keeps full IPs for `-force` rebuilds; it is not under the nginx vhost.
- No login, no rate limit, no geo-restriction. The site is a deliberately
  public read-only summary of attack telemetry.
- Detail pages add a `← Back to index` link; the homepage lists every day with
  partial / budget / restart / unknown flags inline.
- `tcp/22` scans never reach the host's NIC and are not in the totals; the
  filter being detached across deploys/reboots is shown as missing minutes,
  not as zero attacks. RUNBOOK.md § "XDP attack stats site" has the full
  recovery flow including key rotation via `opennhp/demo` and a forced rebuild
  via `xdplog-report site -force`.

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
