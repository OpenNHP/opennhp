---
layout: page
title: Understand the Code
nav_order: 8
permalink: /code/
---

# Understand the Source Code
{: .fs-9 }

This article explains the architecture and structure of the OpenNHP codebase.
{: .fs-6 .fw-300 }

[中文版](/zh-cn/code/){: .label .fs-4 }

---

## Repository Structure

OpenNHP uses a multi-module Go architecture:

```
opennhp/
├── nhp/                 # Core protocol library (github.com/OpenNHP/opennhp/nhp)
├── endpoints/           # Network daemons (github.com/OpenNHP/opennhp/endpoints)
├── examples/            # Example implementations
├── docs/                # Documentation (Jekyll)
├── docker/              # Container configurations
└── release/             # Build outputs
```

## Core Library (`nhp/`)

The `nhp` module contains the core NHP protocol implementation:

| Directory | Purpose |
|-----------|---------|
| `core/` | Protocol implementation: packets, encryption, device management |
| `common/` | Shared types, message structures, error definitions |
| `utils/` | Helper functions: IP utilities, iptables, crypto helpers |
| `plugins/` | Server plugin system interfaces |
| `log/` | Async logging framework |
| `etcd/` | etcd integration for distributed configuration |
| `ebpf/` | eBPF programs for XDP and traffic control |
| `test/` | Unit tests |

### Key Files in `core/`

- `device.go` - NHP device lifecycle and connection management
- `packet.go` - Packet structure and header definitions
- `crypto.go` - Cryptographic primitives (ECDH, AEAD)
- `initiator.go` - Client-side message encryption
- `responder.go` - Server-side message decryption

## Network Daemons (`endpoints/`)

The `endpoints` module contains executable daemons:

| Component | Binary | Purpose |
|-----------|--------|---------|
| `agent/` | `nhp-agent` | Client that sends knock requests |
| `server/` | `nhp-server` | Central server handling knock validation |
| `ac/` | `nhp-ac` | Access Controller managing firewall rules |
| `db/` | `nhp-db` | Data broker for DHP (Data Hiding Protocol) |
| `kgc/` | `nhp-kgc` | Key Generation Center for IBC keys |

Each daemon follows a similar structure:

```
agent/
├── main/           # Entry point and CLI
│   ├── main.go     # CLI commands
│   ├── export.go   # C FFI exports for SDK
│   └── etc/        # Configuration files
├── udpagent.go     # UDP transport implementation
├── config.go       # Configuration handling
└── msghandler.go   # Message processing
```

## Cryptographic Schemes

OpenNHP supports two cipher schemes:

| Scheme | Algorithms | Use Case |
|--------|-----------|----------|
| `CIPHER_SCHEME_CURVE` | Curve25519 + ChaCha20-Poly1305 + BLAKE2s | International |
| `CIPHER_SCHEME_GMSM` | SM2 + SM4-GCM + SM3 | Chinese standards |

See [Cryptography](/cryptography/) for detailed protocol documentation.

## Plugin System

Server plugins extend NHP server functionality:

```go
type PluginHandler interface {
    Init(helper *NhpServerPluginHelper) error
    Close() error
    AuthWithNHP(req *AuthRequest) (*AuthResponse, error)
    AuthWithHttp(req *HttpAuthRequest) (*HttpAuthResponse, error)
}
```

See [Server Plugin Development](/server_plugin/) for implementation guide.

## SDK Architecture

The agent provides SDKs for multiple platforms:

| Platform | Output | Build Target |
|----------|--------|--------------|
| Linux | `nhp-agent.so` | `make linuxagentsdk` |
| macOS | `nhp-agent.dylib` | `make macosagentsdk` |
| iOS | `nhpagent.xcframework` | `make iosagentsdk` |
| Android | `libnhpagent.so` | `make androidagentsdk` |

See [Agent SDK](/agent_sdk/) for integration documentation.

## Building from Source

```bash
# Initialize dependencies
make init

# Build all binaries
make

# Build specific components
make agentd      # nhp-agent
make serverd     # nhp-server
make acd         # nhp-ac

# Development commands
make test        # Run tests
make fmt         # Format code
make clean       # Clean build artifacts
```

## Testing

Tests are located in `nhp/test/` and `endpoints/test/`:

```bash
# Run all tests
make test

# Run with race detection
make test-race

# Run with coverage
make coverage
```

## Contributing

See [CONTRIBUTING.md](https://github.com/OpenNHP/opennhp/blob/main/CONTRIBUTING.md) for development guidelines.
