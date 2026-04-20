---
layout: page
title: Message Header
parent: Protocol Reference
nav_order: 1
permalink: /protocol/header/
description: Byte layout and field semantics of the fixed NHP message header.
---

# NHP Message Header
{: .fs-9 }

Every NHP packet begins with a fixed-length header that carries identity,
ephemeral key material, replay-protection state, and an HMAC over the rest
of the message. This page documents each field with its byte offset, the
obfuscation scheme that hides public metadata in transit, and a pointer
into the Go code that parses it.
{: .fs-6 .fw-300 }

*Implements CSA Stealth Mode SDP §NHP Message Header (Table 3). Parsed in [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go).*

## Layout

- **Standard (international cipher suite)**: **160 bytes** (Curve25519, ChaCha20-Poly1305, BLAKE2s).
- **Extended (domestic / Chinese cipher suite)**: **224 bytes** (SM2, SM4-GCM, SM3).

The encrypted message payload (plaintext size + 16-byte AEAD tag) follows
the header. Total on-wire packet size is `header + ciphertext`, bounded by
the UDP max of 65,535 bytes. TCP short-connection transport is also
supported when UDP is unavailable.

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|         Leading Obfuscation (4, random, XOR mask source)      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|       Message Type (2)        |      Message Length (2)       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| Ver Major (1) | Ver Minor (1) |       Protocol Flags (2)      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       Reserved (4, zero)                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|                       Counter (8, big-endian)                 |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|        Ephemeral Public Key  (32 Curve25519 / 64 SM2)         |
|                            ...                                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|                      IBC Identity (80)                        |
|                            ...                                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|          Local Public Key Ciphertext (48 / 80, AEAD)          |
|                            ...                                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|             Timestamp Ciphertext (24, AEAD-wrapped)           |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|                          HMAC (32)                            |
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|              Encrypted message payload + tag (variable)       |
|                            ...                                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

## Fields

| Offset | Size (bytes) | Field | Description |
|---:|---:|---|---|
| 0 | 4 | **Leading Obfuscation** | 4 random bytes generated per packet. Used as an XOR mask for Message Type and Message Length so that passive observers cannot trivially identify the packet type or size. Plaintext on the wire; the receiver reads it first, then deobfuscates the two subsequent fields. |
| 4 | 2 | **Message Type** | NHP message type ID (see [Message Types]({{ '/protocol/messages/' | relative_url }})). XOR-obfuscated with part of the Leading Obfuscation. |
| 6 | 2 | **Message Length** | Length of the ciphertext that follows the header, **including** the 16-byte AEAD tag, big-endian. XOR-obfuscated. Max 65,535 (UDP limit). |
| 8 | 1 | **Protocol Major Version** | Plaintext. Receivers silently discard packets with an unsupported major version (default-deny). |
| 9 | 1 | **Protocol Minor Version** | Plaintext. Backward-compatible increments only. |
| 10 | 2 | **Protocol Flags** | Big-endian bit flags. Bit 0 = payload compressed (zlib). Other bits reserved for handshake-pattern selection, PQC hybrid mode, etc. Unused bits zero. |
| 12 | 4 | **Reserved** | Must be zero. Forward-compatibility padding. |
| 16 | 8 | **Counter** | 64-bit big-endian nonce and transaction tracker. Derived from the Noise Protocol Framework's `n` nonce. Monotonically increments per encryption. Replay protection: receivers reject counters older than or equal to the last-seen value for that session. |
| 24 | 32 or 64 | **Ephemeral Public Key** | Fresh per-packet ephemeral public key — 32 bytes for Curve25519/X25519 (standard), 64 bytes for SM2 (extended, X‖Y coordinates). Drives Noise handshake key derivation and provides forward secrecy even for single-shot message types (keepalives, logs). |
| 56 / 88 | 80 | **IBC Identity** | Sender identity used by Identity-Based Cryptography. Reserved / zero when PKI mode is in use. |
| 136 / 168 | 48 or 80 | **Local Public Key Ciphertext** | AEAD-encrypted long-term / static public key of the sender (ID-based or PKI). Decrypted with keys derived from the ephemeral DH. Size = plaintext key (32 or 64 bytes) + 16-byte tag. |
| 184 / 248 | 24 | **Timestamp Ciphertext** | AEAD-encrypted 8-byte UNIX-milliseconds timestamp + 16-byte tag. Freshness check: receivers enforce a tolerance window (default ±5 min) to defeat replay while surviving clock skew. |
| 208 / 272 | 32 | **HMAC** | HMAC over the entire header (with this field zeroed) plus the payload ciphertext. Keyed by the Noise chaining key. Validated **before** AEAD decryption — a mismatched HMAC causes a silent drop, aligning with default-deny. |

Total header length: **160** bytes (standard) or **224** bytes (extended).

## Obfuscation scheme

The first three fields — Leading Obfuscation, Message Type, Message Length —
are the only plaintext-adjacent fields. A naive observer scanning UDP traffic
for the NHP type code would learn nothing useful, because the Type and Length
bytes are XOR-masked by the per-packet random.

```
wire[4..6]  = type  XOR wire[0..2]
wire[6..8]  = len   XOR wire[2..4]
```

This is defence-in-depth — the payload itself is AEAD-encrypted and the static
key is AEAD-wrapped — but the obfuscation foils trivial traffic analysis and
fingerprinting.

## Version and flag handling

- **Version mismatch** → silent discard. Never respond to mismatched protocol
  versions; that would reveal server presence.
- **Unknown flags** → reserved bits must be zero; if a receiver sees a
  reserved bit set it may silently discard or log-and-drop depending on
  deployment policy. Future PQC-hybrid modes are planned to land in Protocol
  Flags rather than forcing a major-version bump.

## Concurrent sessions

Multiple simultaneous knocks from the same NHP-Agent (e.g., targeting different
Protected Resources) are distinguished by:

1. **Fresh ephemeral keys** per handshake — no collision even within the same
   second.
2. **Independent Noise state** (CipherState + chaining key) per session.
3. **Session ID** (payload-level, see the [NHP-ACK message]({{ '/protocol/messages/' | relative_url }})).

No sender-side coordination or locking is required.

## See also

- [Message Types]({{ '/protocol/messages/' | relative_url }}) — how each ID is handled after the header parses
- [Cryptography]({{ '/cryptography/' | relative_url }}) — the algorithms the AEAD and Noise state rely on
- [Glossary]({{ '/glossary/' | relative_url }}) — definitions for *counter*, *ephemeral key*, *HMAC*, *cipher scheme*
