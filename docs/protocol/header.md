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
ephemeral key material, replay-protection state, and an HMAC over the
header itself. This page documents each field with its byte offset, the
obfuscation scheme that hides public metadata in transit, and a pointer
into the Go code that parses it.
{: .fs-6 .fw-300 }

*Implements CSA Stealth Mode SDP §NHP Message Header (Table 3). Constants defined in [`nhp/core/scheme/curve/header.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/scheme/curve/header.go) and [`nhp/core/scheme/gmsm/header.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/scheme/gmsm/header.go); dispatch in [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go).*

{: .note }
The CSA whitepaper quotes 160 / 224 bytes, matching an earlier draft in which
the sender's static public key was the only encrypted identity material in the
header. The current Go implementation also wraps an 80-byte IBC identity
block, bringing the totals to **240 / 304 bytes**. When the spec and the code
disagree, this documentation tracks the code (per [Protocol Reference]({{ '/protocol/' | relative_url }}) §Spec version).

## Layout

- **Standard (international cipher suite)**: **240 bytes** — `CIPHER_SCHEME_CURVE` (Curve25519, AES-256-GCM, BLAKE2s).
- **Extended (domestic / Chinese cipher suite)**: **304 bytes** — `CIPHER_SCHEME_GMSM` (SM2, SM4-GCM, SM3).

The AEAD-encrypted message payload (plaintext size + 16-byte GCM tag) follows
the header. Total on-wire packet size is `header + ciphertext`, bounded by
the UDP max of 65,535 bytes. TCP short-connection transport is also
supported when UDP is unavailable.

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|         Leading Obfuscation (4, random, XOR mask source)      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|         Message Type (2) ‖ Message Length (2)  XOR-masked      |
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
|                IBC Identity Ciphertext (80, AEAD)             |
|                            ...                                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
|         Static Public Key Ciphertext (48 / 80, AEAD)          |
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
| 0 | 4 | **Leading Obfuscation** | 4 random bytes generated per packet. Used as a one-shot XOR mask for the 4-byte Type‖Length tuple that follows, so that passive observers cannot trivially identify the packet type or size. Plaintext on the wire; the receiver reads it first, then deobfuscates. |
| 4 | 2 | **Message Type** | NHP message type ID (see [Message Types]({{ '/protocol/messages/' | relative_url }})). Packed big-endian with Message Length into a single 32-bit word and XOR-masked against the Leading Obfuscation. |
| 6 | 2 | **Message Length** | Length of the ciphertext that follows the header, **including** the 16-byte AEAD tag, big-endian. XOR-masked together with Message Type. Max 65,535 (UDP limit). |
| 8 | 1 | **Protocol Major Version** | Plaintext. Receivers silently discard packets with an unsupported major version (default-deny). |
| 9 | 1 | **Protocol Minor Version** | Plaintext. Backward-compatible increments only. |
| 10 | 2 | **Protocol Flags** | Big-endian bit flags (see table below). Unused bits in the lower 12 should be zero; bits 12–15 carry the cipher-scheme selector. |
| 12 | 4 | **Reserved** | Zero in current senders; receivers ignore the contents. Forward-compatibility padding. |
| 16 | 8 | **Counter** | 64-bit big-endian nonce and transaction tracker. The low 8 bytes form the AEAD nonce (`NonceBytes` prepends 4 zero bytes so the 12-byte GCM nonce is derived from the counter). Monotonically increments per encryption; receivers reject stale counters for that session. |
| 24 | 32 or 64 | **Ephemeral Public Key** | Fresh per-packet ephemeral public key — 32 bytes for Curve25519/X25519 (standard), 64 bytes for SM2 (extended, uncompressed X‖Y coordinates). Drives Noise handshake key derivation and provides forward secrecy even for single-shot message types. |
| 56 / 88 | 80 | **IBC Identity Ciphertext** | AEAD-encrypted IBC identity block (64-byte plaintext + 16-byte tag). Zero-padded plaintext when PKI mode is used. |
| 136 / 168 | 48 or 80 | **Static Public Key Ciphertext** | AEAD-encrypted long-term static public key of the sender. Decrypted with keys derived from the ephemeral DH. Size = plaintext key (32 or 64 bytes) + 16-byte tag. |
| 184 / 248 | 24 | **Timestamp Ciphertext** | AEAD-encrypted 8-byte UNIX-milliseconds timestamp + 16-byte tag. Freshness check: receivers enforce a tolerance window to defeat replay while surviving clock skew. |
| 208 / 272 | 32 | **HMAC** | Keyed hash over every preceding header byte (offsets 0 through the byte immediately before this field) — i.e., the entire header **excluding** the HMAC itself; the payload ciphertext is **not** covered. For `NHP-RKN`, the previously issued cookie is appended to the hash input. Validated **before** AEAD decryption — a mismatched HMAC causes a silent drop, aligning with default-deny. See [`MsgAssemblerData.addHMAC`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/initiator.go). |

Total header length: **240** bytes (standard) or **304** bytes (extended).

### Protocol Flags (bit 0 = LSB)

| Bit | Name | Meaning |
|---:|---|---|
| 0 | `NHP_FLAG_EXTENDEDLENGTH` | Set for the 304-byte GMSM header; clear for the 240-byte Curve header. |
| 1 | `NHP_FLAG_COMPRESS` | Payload ciphertext plaintext was zlib-compressed before encryption. |
| 2 | `NHP_FLAG_CL_PKC` | CL-PKC (certificate-less public-key cryptography) mode. |
| 3–11 | — | Reserved. |
| 12 | cipher-scheme selector | `0` = `CIPHER_SCHEME_CURVE`; `1` = `CIPHER_SCHEME_GMSM`. |
| 13–15 | — | Reserved for additional cipher schemes. |

See [`nhp/common/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/packet.go) for the authoritative constants.

## Obfuscation scheme

The first eight bytes on the wire — Leading Obfuscation followed by the
Type‖Length tuple — are the only pre-decryption surface. A single 4-byte
XOR masks Type and Length together against the Leading Obfuscation word:

```
preamble      = wire[0..4]            // big-endian uint32 random
type_and_len  = wire[4..8]            // big-endian uint32, XOR-masked
decoded       = preamble XOR type_and_len
type          = (decoded >> 16) & 0xFFFF
length        = decoded        & 0xFFFF
```

This is defence-in-depth — the payload itself is AEAD-encrypted and the static
key is AEAD-wrapped — but the obfuscation foils trivial traffic analysis and
fingerprinting. See `SetTypeAndPayloadSize` / `TypeAndPayloadSize` in
[`nhp/core/scheme/curve/header.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/scheme/curve/header.go).

## Version and flag handling

- **Version mismatch** → silent discard. Never respond to mismatched protocol
  versions; that would reveal server presence.
- **Unknown flags** → the lower 12 bits that are not defined in the table
  above should be zero. Receivers may silently discard or log-and-drop
  depending on deployment policy. Future PQC-hybrid modes are planned to land
  in Protocol Flags rather than forcing a major-version bump.

## Concurrent sessions

Multiple simultaneous knocks from the same NHP-Agent (e.g., targeting different
Protected Resources) are distinguished by:

1. **Fresh ephemeral keys** per handshake — no collision even within the same
   second.
2. **Independent Noise state** (CipherState + chaining key) per session.
3. **Session / transaction ID** (payload-level, see the [NHP-ACK message]({{ '/protocol/messages/' | relative_url }})).

No sender-side coordination or locking is required.

## See also

- [Message Types]({{ '/protocol/messages/' | relative_url }}) — how each ID is handled after the header parses
- [Cryptography]({{ '/cryptography/' | relative_url }}) — the algorithms the AEAD and Noise state rely on
- [Glossary]({{ '/glossary/' | relative_url }}) — definitions for *counter*, *ephemeral key*, *HMAC*, *cipher scheme*
