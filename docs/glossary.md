---
layout: page
title: Glossary
nav_order: 12
permalink: /glossary/
description: Canonical terminology used across OpenNHP and the CSA NHP specification.
---

# Glossary
{: .fs-9 }

Canonical terms used across OpenNHP and the CSA NHP specification. Prefer these
spellings in issues, PRs, generated code, and tooling. The spec sometimes uses
alternate phrasings in SDP-integration contexts; where that happens, the
alternate is listed but the canonical term is what the implementation uses.
{: .fs-6 .fw-300 }

*Implements CSA Stealth Mode SDP §NHP Core Components and §Components That Interact with NHP.*

## Protocol roles

| Canonical term | Also seen as | Meaning |
|---|---|---|
| **NHP-Agent** | agent, client, initiator | Client-side component embedded in a host, SDK, browser, or app. Sends the encrypted knock request and, once access is granted, connects to the Protected Resource. |
| **NHP-Server** | server, controller (SDP) | Control-plane component that decrypts knocks, authenticates the NHP-Agent, checks policy via the ASP, and instructs the NHP-AC. Maps to the *Policy Engine/Administrator* in NIST SP 800-207. |
| **NHP-AC** | Access Controller, gateway (SDP) | Data-plane component that enforces default-deny and dynamically opens the firewall on the NHP-Server's instruction. Maps to the *Policy Enforcement Point* in NIST SP 800-207. |
| **NHP-DB** | — | Data-object backend for the Data-object Hiding Protocol (DHP). Stores encrypted zero-trust data objects. |
| **NHP-KGC** | KGC | Key Generation Center used by Identity-Based Cryptography (IBC) to issue agent/server keys. |
| **NHP-Relay** | relay | Optional forwarder for NHP packets across network segments; preserves source addresses via the NHP-RLY message type. |

## Interacting parties

| Term | Meaning |
|---|---|
| **Resource Requestor** | The entity (user, device, application, or server) that hosts an NHP-Agent and seeks access to a Protected Resource. |
| **Authorization Service Provider (ASP)** | External IAM / policy system the NHP-Server queries to authorize access (e.g., corporate IAM, SDP Controller when NHP fronts an SDP deployment). |
| **Protected Resource** | The service, host, port, or data object being concealed by NHP. |
| **Log Server** | Optional external SIEM or log store fed by NHP-Server / NHP-AC access and audit logs. |

## Protocol concepts

| Term | Meaning |
|---|---|
| **Knock** | The initial encrypted packet (message type `NHP-KNK`) an NHP-Agent sends to an NHP-Server. Contains the agent's identity, device information, and the target resource identifier. |
| **Authenticate-before-connect** | The design principle that identity is verified *before* any TCP/UDP session to the Protected Resource is allowed. NHP's core contribution. |
| **Default-deny** | The posture NHP-AC enforces by default: no inbound traffic is allowed until the NHP-Server explicitly instructs otherwise. |
| **Session** | An authenticated, time-bounded window during which an NHP-Agent may connect to the Protected Resource. Identified by a `Session ID` carried across `NHP-KNK`/`NHP-ACK`/`NHP-AOP`/`NHP-ART`. |
| **Open-door** | The action of dynamically permitting a specific source → destination flow in NHP-AC. Triggered by an `NHP-AOP` message. |
| **Ephemeral key** | A fresh ECC key pair generated per message to drive the Noise handshake and provide per-session forward secrecy. |
| **Static key** | The long-lived identity key for an NHP-Agent / NHP-Server / NHP-AC. Registered out-of-band or via the NHP-REG / NHP-RAK flow. |

## Cryptography

| Term | Meaning |
|---|---|
| **ECC** | Elliptic Curve Cryptography. NHP's default curve is Curve25519 (256-bit) for the international cipher suite; SM2 for the extended/domestic suite. |
| **Noise Protocol Framework** | The handshake + symmetric-crypto framework NHP uses for mutual authentication and per-session key derivation. OpenNHP's handshake follows the Noise `K` pattern structure (`e`, `es`, `ss`), extended with an AEAD-wrapped IBC identity block (`MaximumIdentitySize`, zeroed in PKI mode) inserted before the static-key ciphertext. |
| **IBC** | Identity-Based Cryptography. An alternative to PKI where a user's ID string *is* their public key. Simplifies key distribution at the cost of key escrow at the NHP-KGC. |
| **CL-PKC** | Certificate-Less Public Key Cryptography. A variant of IBC that mitigates key escrow by splitting key generation between the KGC and the user. |
| **PKI** | Traditional Public Key Infrastructure with X.509 certificates and a CA hierarchy. Supported alongside IBC. |
| **Cipher scheme** | A named bundle of crypto primitives. `CIPHER_SCHEME_CURVE` = Curve25519 + AES-256-GCM + BLAKE2s (international / standard). `CIPHER_SCHEME_GMSM` = SM2 + SM4-GCM + SM3 (Chinese national standard). |

## Architectural frames

| Term | Meaning |
|---|---|
| **Zero Trust Architecture (ZTA)** | The security model NHP operationalizes — "never trust, always verify" — as formalized in NIST SP 800-207. |
| **Software-Defined Perimeter (SDP)** | The CSA framework NHP was originally designed to enhance. In SDP terminology, NHP-AC aligns with the SDP *Gateway* and NHP-Server aligns with the SDP *Controller*. |
| **Single-Packet Authorization (SPA)** | The 2nd-generation network-hiding technology NHP supersedes. SPA uses one-way encrypted knocks; NHP adds mutual authentication, explicit feedback, and scalability. |
| **Default-open vs default-deny** | TCP/IP defaults to *open* (any peer can initiate a connection). NHP enforces *deny* until authentication succeeds. |

## Message-type prefix

All NHP protocol message names start with `NHP-` followed by a three-letter mnemonic. See the [Message Types reference]({{ '/protocol/messages/' | relative_url }}) for the full table.
