---
title: Architecture
---

# OpenNHP Architecture

The OpenNHP architecture follows the **NIST Zero Trust Architecture** standard with a modular design. Hover over components to learn more, or click **Show Flow** to see the protocol in action.

<ArchDiagram />

## Protocol Highlights

- **Default Deny-All**: All resources are hidden until authenticated access is granted
- **Encrypted UDP Knock**: Uses Noise Protocol Framework for secure communication
- **Time-Limited Access**: Opened paths automatically expire after the configured TTL
