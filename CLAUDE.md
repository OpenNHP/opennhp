# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OpenNHP is a Go-based Zero Trust security toolkit implementing two core protocols:
- **NHP (Network-infrastructure Hiding Protocol)**: Conceals server ports, IPs, and domains from unauthorized access
- **DHP (Data-object Hiding Protocol)**: Ensures data security via encryption and confidential computing

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
  - `db/`: NHP-DB - data object management for DHP
  - `kgc/`: Key Generation Center for IBC (Identity-Based Cryptography)
  - `relay/`: TCP relay functionality

### Core Concepts

**Device Types** (defined in `nhp/core/device.go`):
- `NHP_AGENT`: Client initiating access requests
- `NHP_SERVER`: Central authentication/authorization server
- `NHP_AC`: Access controller managing network rules
- `NHP_DB`: Data object backend for DHP
- `NHP_RELAY`: Packet relay

**Packet Types** (defined in `nhp/core/packet.go`):
- `NHP_KNK`: Agent knock request
- `NHP_ACK`: Server knock acknowledgment
- `NHP_AOP`: Server-to-AC operation request
- `NHP_ART`: AC operation result
- `NHP_REG`/`NHP_RAK`: Agent registration flow
- `DHP_*`: Data Hiding Protocol messages

**Cipher Schemes** (in `nhp/core/crypto.go`):
- `CIPHER_SCHEME_CURVE`: Curve25519 + ChaCha20-Poly1305 + BLAKE2s
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

## Protocol Flow

1. Agent sends encrypted knock (`NHP_KNK`) to Server
2. Server validates, sends operation request (`NHP_AOP`) to AC
3. AC opens firewall, responds (`NHP_ART`) to Server
4. Server sends acknowledgment (`NHP_ACK`) with access info to Agent
5. Agent can now access the protected resource through AC
