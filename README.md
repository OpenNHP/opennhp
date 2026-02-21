[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![OpenNHP Logo](docs/images/logo11.png)

# OpenNHP: Open Source Zero Trust Security Toolkit

![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Version](https://img.shields.io/badge/version-1.0.0-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
[![codecov](https://codecov.io/gh/OpenNHP/opennhp/branch/main/graph/badge.svg)](https://codecov.io/gh/OpenNHP/opennhp)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/OpenNHP/opennhp)

**OpenNHP** is a lightweight, cryptography-powered, open-source toolkit implementing Zero Trust security for infrastructure, applications, and data. It features two core protocols:

- **Network-infrastructure Hiding Protocol (NHP):** Conceals server ports, IP addresses, and domain names to protect applications and infrastructure from unauthorized access.
- **Data-object Hiding Protocol (DHP):** Ensures data security and privacy via encryption and confidential computing, making data *"usable but not visible."*

**[Website](https://opennhp.org) · [Documentation](https://docs.opennhp.org) · [Live Demo](https://acdemo.opennhp.org) · [Discord](https://discord.gg/CpyVmspx5x)**

---

## Architecture

OpenNHP follows a modular design with three core components, inspired by the [NIST Zero Trust Architecture](https://www.nist.gov/publications/zero-trust-architecture):

![OpenNHP architecture](docs/images/OpenNHP_Arch.png)

| Component | Role |
|-----------|------|
| **NHP-Agent** | Client that sends encrypted knock requests to gain access |
| **NHP-Server** | Authenticates and authorizes requests; decoupled from protected resources |
| **NHP-AC** | Access controller that manages firewall rules on the protected server |

> For protocol details, deployment models, and cryptographic design, see the [documentation](https://docs.opennhp.org).

---

## Repository Structure

```
opennhp/
├── nhp/              # Core protocol library (Go module)
│   ├── core/         # Packet handling, cryptography, Noise Protocol, device management
│   ├── common/       # Shared types and message definitions
│   ├── utils/        # Utility functions
│   ├── plugins/      # Plugin handler interfaces
│   ├── log/          # Logging infrastructure
│   └── etcd/         # Distributed configuration support
└── endpoints/        # Daemon implementations (Go module, depends on nhp)
    ├── agent/        # NHP-Agent daemon
    ├── server/        # NHP-Server daemon
    ├── ac/           # NHP-AC (access controller) daemon
    ├── db/           # NHP-DB (data object backend for DHP)
    ├── kgc/          # Key Generation Center (IBC)
    └── relay/        # TCP relay
```

---

## Quick Start

### Prerequisites

- Go 1.25.6+
- `make`
- Docker and Docker Compose (for the full-stack demo)

### Build

```bash
# Build all components
make

# Build individual daemons
make agentd    # NHP-Agent
make serverd   # NHP-Server
make acd       # NHP-AC
make db        # NHP-DB
make kgc       # Key Generation Center
```

### Test

```bash
cd nhp && go test ./...
cd endpoints && go test ./...
```

### Run with Docker

```bash
cd docker && docker-compose up --build
```

Follow the [Quick Start tutorial](https://docs.opennhp.org/nhp_quick_start/) to simulate the full authentication workflow in a Docker environment.

---

## Contributing

We welcome contributions! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting pull requests.

**Note:** All commits must be signed with a verified GPG or SSH key.

```bash
git commit -S -m "your message"
```

---

## Sponsors

<a href="https://layerv.ai">
  <img src="docs/images/layerv_logo.png" width="80" alt="LayerV.ai">
  <br>
  <img src="docs/images/layerv_text.svg" width="120" alt="LayerV.ai">
</a>

---

## License

Released under the [Apache 2.0 License](LICENSE).

## Contact

- Email: [support@opennhp.org](mailto:support@opennhp.org)
- Discord: [Join our Discord](https://discord.gg/CpyVmspx5x)
- Website: [https://opennhp.org](https://opennhp.org)
