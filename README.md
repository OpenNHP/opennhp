[![en](https://img.shields.io/badge/lang-en-red.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)

# OpenNHP: Zero Trust Network Hiding Protocol

![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Version](https://img.shields.io/badge/version-0.0.1-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)

![OpenNHP Logo](path_to_logo.png)

OpenNHP is an open-source implementation of the Network-infrastructure Hiding Protocol (NHP), a next-generation zero trust security protocol designed to secure network connections and hide infrastructure.

---

## Table of Contents

- [Introduction](#introduction)
- [Key Features](#key-features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage Example](#usage-example)
- [Dependencies](#dependencies)
- [Architecture](#architecture)
- [Components](#components)
- [Workflow](#workflow)
- [Deployment Models](#deployment-models)
- [Security Benefits](#security-benefits)
- [Cryptographic Foundations](#cryptographic-foundations)
- [Comparison to SPA](#comparison-to-spa)
- [Compatibility](#compatibility)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Acknowledgments](#acknowledgments)
- [License](#license)
- [Contact](#contact)

---

## Introduction

OpenNHP implements the Network-infrastructure Hiding Protocol (NHP), which operates at the session layer (layer 5) of the OSI model to enforce zero trust principles. NHP addresses key security challenges:

- Mitigates vulnerability exploitation by enforcing "deny-all" rules by default
- Prevents phishing attacks through encrypted DNS resolution
- Protects against DDoS attacks by hiding infrastructure
- Enables attack attribution through identity-based connections

NHP enhances security, reliability, and scalability compared to previous approaches like Single Packet Authorization (SPA).

> 🔐 OpenNHP is designed with security-first principles, implementing a true zero-trust architecture.

---

## Key Features

- Default-deny access control for all protected resources
- Identity and device-based authentication before network access
- Encrypted DNS resolution to prevent DNS hijacking
- Distributed infrastructure to mitigate DDoS attacks
- Scalable architecture with decoupled components
- Integration with existing identity and access management systems
- Support for various deployment models (client-to-gateway, client-to-server, etc)
- Cryptographically secure using modern algorithms (ECC, Noise Protocol, IBC)

<details>
<summary>Click to expand feature details</summary>

- **Default-deny access control**: All resources are hidden by default, only becoming accessible after authentication and authorization.
- **Identity and device-based authentication**: Ensures that only known users on approved devices can gain access.
- **Encrypted DNS resolution**: Prevents DNS hijacking and associated phishing attacks.
- **DDoS mitigation**: Distributed infrastructure design helps protect against Distributed Denial of Service attacks.
- **Scalable architecture**: Decoupled components allow for flexible deployment and scaling.
- **IAM integration**: Works with your existing Identity and Access Management systems.
- **Flexible deployment**: Supports various models including client-to-gateway, client-to-server, and more.
- **Strong cryptography**: Utilizes modern algorithms like ECC, Noise Protocol, and IBC for robust security.

</details>

---

## Quick Start

Get OpenNHP up and running in minutes:

```bash
git clone https://github.com/opennhp/nhp.git
cd nhp
make
./nhp-server run
```

---

## Installation

Detailed installation instructions:

1. Clone the repository:
   ```bash
   git clone https://github.com/opennhp/nhp.git
   ```
2. Navigate to the project directory:
   ```bash
   cd nhp
   ```
3. Build the project:
   ```bash
   make
   ```
4. Install (optional):
   ```bash
   sudo make install
   ```

> ⚠️ Note: Running `sudo make install` requires root privileges. Make sure you trust the source before running this command.

---

## Usage Example

Here's a simple example of using OpenNHP to secure a connection:

```python
from opennhp import NHPClient

client = NHPClient("config.yaml")
connection = client.connect("protected-resource-id")
response = connection.send("Hello, secure world!")
print(response)
```

---

## Dependencies

OpenNHP requires the following major dependencies:

| Dependency | Minimum Version |
|------------|-----------------|
| OpenSSL    | 1.1.1           |
| Boost      | 1.66            |
| Protocol Buffers | 3.0       |

---

## Architecture

OpenNHP follows a modular architecture with the following core components:

- NHP-Agent: Client-side component that initiates requests
- NHP-Server: Processes requests and manages authentication
- NHP-AC: Enforces access policies on protected resources

These interact with:

- Protected Resources: Applications, servers, network devices to be secured
- Authorization Service Provider: Validates access policies

---

## Components

### NHP-Agent

The NHP-Agent is a client-side component that initiates communication and requests access to protected resources. It can be implemented as:

- A standalone client application
- An SDK integrated into existing applications
- A browser plugin
- A mobile app

The agent is responsible for:

- Generating and sending knock requests to the NHP-Server
- Maintaining secure communication channels
- Handling authentication flows

### NHP-Server

The NHP-Server is the central controller that:

- Processes and validates knock requests from agents
- Interacts with the Authorization Service Provider for policy decisions
- Manages NHP-AC components to allow/deny access
- Handles key management and cryptographic operations

It can be deployed in a distributed or clustered configuration for high availability and scalability.

### NHP-AC

NHP-AC (Access Control) components enforce access policies on protected resources. Key functions:

- Implement default deny-all rules
- Open/close access based on NHP-Server instructions
- Ensure network invisibility of protected resources
- Log access attempts

---

## Workflow

1. `NHP-Agent` sends knock request to `NHP-Server`
2. `NHP-Server` validates request and retrieves agent info
3. `NHP-Server` queries Authorization Service Provider
4. If authorized, `NHP-Server` instructs `NHP-AC` to allow access
5. `NHP-AC` opens connection and notifies `NHP-Server`
6. `NHP-Server` provides resource access details to `NHP-Agent`
7. `NHP-Agent` can now access the protected resource
8. Access is logged for auditing purposes

---

## Deployment Models

OpenNHP supports multiple deployment models to suit different use cases:

- Client-to-Gateway: Secures access to multiple servers behind a gateway
- Client-to-Server: Directly secures individual servers/applications
- Server-to-Server: Secures communication between backend services
- Gateway-to-Gateway: Secures site-to-site connections

---

## Security Benefits

- Reduces attack surface by hiding infrastructure
- Prevents unauthorized network reconnaissance
- Mitigates vulnerability exploitation
- Stops phishing via encrypted DNS
- Protects against DDoS attacks
- Enables fine-grained access control
- Provides identity-based connection tracking

---

## Cryptographic Foundations

OpenNHP leverages state-of-the-art cryptographic algorithms:

- Elliptic Curve Cryptography (ECC): For efficient public key operations
- Noise Protocol Framework: For secure key exchange and identity verification
- Identity-Based Cryptography (IBC): For simplified key management at scale

---

## Comparison to SPA

NHP offers several advantages over Single Packet Authorization (SPA):

- Decoupled architecture improves scalability
- Bidirectional communication increases reliability
- Modern cryptographic primitives enhance security
- More comprehensive infrastructure hiding capabilities
- Extensible design supports various use cases
- Interoperable with existing protocols and systems

---

## Compatibility

OpenNHP is designed for broad compatibility:

### Cryptographic Algorithms

- International algorithms: Curve25519, AES, SHA256

### Operating Systems

- Windows
- macOS
- Linux (various distributions)
- iOS
- Android

### CPU Architectures

- x86
- ARM
- RISC-V
- MIPS

---

## Roadmap

Our plans for the near future include:

- 

---

## Contributing

We welcome contributions to OpenNHP! Please see our [Contributing Guidelines](CONTRIBUTING.md) for more information on how to get involved.

---

## Acknowledgments

We'd like to thank the Cloud Security Alliance for their work on the SDP specification, which inspired many aspects of NHP.

---

## License

OpenNHP is released under the [Apache 2.0 License](LICENSE).

---

## Contact

- Project Website: [https://opennhp.org](https://opennhp.org)
- Mailing List: community@opennhp.org
- Slack Channel: [Join our Slack](https://slack.opennhp.org)

For more detailed documentation, please visit our [Official Documentation](https://docs.opennhp.org).

---

🌟 Thank you for your interest in OpenNHP! We look forward to your contributions and feedback.
