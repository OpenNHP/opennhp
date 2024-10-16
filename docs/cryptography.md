---
layout: page
title: Cryptography
nav_order: 4
permalink: /cryptography/
---

# Cryptographic Algorithms in OpenNHP
{: .fs-9 }

[中文版](/zh-cn/cryptography/){: .label .fs-4 }

---

Cryptography is at the heart of OpenNHP, providing robust security, excellent performance, and scalability by utilizing cutting-edge cryptographic algorithms. This article explains how OpenNHP takes advantages of modern cryptographic algorithms in several critical areas:

1. Public Key Cryptography
2. Key Exchange, Data Encryption and Identity Verification
3. Key Distribution and Management

## 1) Public Key Cryptography

### 1.1 Introduction
In the evolving landscape of cybersecurity, securing communications and protecting network resources are essential, especially with the increasing sophistication of cyber threats. The Network Infrastructure Hiding Protocol (NHP), a zero-trust security mechanism, stands at the forefront of efforts to address these concerns by concealing network infrastructure details from attackers and ensuring that only trusted entities can interact with network resources. A key component of NHP's security model is the use of Elliptic Curve Cryptography (ECC) for public key cryptography. In this article, we explore how ECC integrates into the NHP Zero Trust protocol to provide robust and efficient security.

### 1.2 What is Elliptic Curve Cryptography?

Elliptic Curve Cryptography (ECC) is a modern approach to public key cryptography that provides equivalent levels of security with significantly smaller key sizes compared to traditional methods such as RSA. ECC relies on the mathematical properties of elliptic curves over finite fields, providing a powerful balance of security and performance. Due to its reduced computational overhead, ECC is particularly suitable for resource-constrained environments such as embedded systems or mobile devices.

The advantages of ECC include:

- **Smaller Key Sizes**: ECC achieves a high level of security with smaller keys, which translates to faster operations, less bandwidth usage, and reduced computational requirements.
- **Enhanced Security**: The underlying problem of elliptic curve discrete logarithms is computationally complex, making ECC resistant to common forms of cryptographic attack.
- **Efficiency**: With less processing power needed compared to traditional methods, ECC can handle encryption and decryption more efficiently, which is crucial for zero trust environments requiring frequent cryptographic operations.


### 1.3 How NHP Uses ECC for Secure Communication

NHP uses ECC in key exchange, data encryption, identity verification with the Noise protocol framework, and key distribution and management with Certificateless Public Key Cryptography (CL-PKC).


#### 1. **Key Exchange Mechanism**

The secure exchange of encryption keys between communicating entities is the backbone of any secure communication protocol. NHP uses Elliptic Curve Diffie-Hellman (ECDH) for its key exchange mechanism. In the ECDH key exchange, both communicating parties generate a public-private key pair using elliptic curves. The public keys are then exchanged, allowing both parties to compute a shared secret without ever having to transmit it directly over the network.

The benefit of using ECDH in NHP is twofold: first, it provides forward secrecy, meaning that even if the private key of one party is compromised in the future, previously established session keys remain secure. Secondly, because of ECC's efficiency, the key exchange process is computationally lightweight, ensuring that key establishment is performed quickly without a large computational footprint.

#### 2. **Authentication with Digital Signatures**

In a zero-trust environment, authentication is paramount. NHP utilizes Elliptic Curve Digital Signature Algorithm (ECDSA) to verify the authenticity of entities attempting to access network resources. ECDSA, an ECC-based digital signature scheme, allows devices to prove their identity without revealing sensitive private keys.

In the NHP protocol, when an entity wants to communicate with the network, it must provide a digital signature generated with its private key. The receiving entity can then use the corresponding public key to verify the validity of the signature. This ensures that only legitimate entities can participate in the network, effectively implementing the zero-trust model's "never trust, always verify" principle.

#### 3. **Encryption for Data Confidentiality**

NHP employs symmetric encryption for data confidentiality during communication, but symmetric keys must be securely distributed and shared between entities. ECC plays a role in the secure distribution of these symmetric keys through ECDH, providing an encrypted communication channel where symmetric keys are exchanged securely.

Once these keys are exchanged, NHP switches to symmetric encryption for data transfer, benefiting from the speed and efficiency of symmetric encryption algorithms. ECC ensures that the symmetric key exchange is both secure and resource-efficient.

#### 4. **Key Distribution and Management with Certificateless Public Key Cryptography (CL-PKC)**

NHP also leverages ECC for key distribution and management using Certificateless Public Key Cryptography (CL-PKC). In traditional public key infrastructure, certificates are used to validate public keys, which introduces complexity in terms of certificate management. CL-PKC eliminates the need for certificates by allowing entities to generate partial private keys in collaboration with a trusted authority, while also generating their own key pairs independently.

This approach simplifies key management and ensures that public keys can be used securely without the overhead of certificate issuance and validation. By using ECC in CL-PKC, NHP provides a lightweight and secure means of key distribution, further enhancing the zero-trust model by removing dependencies on centralized certificate authorities.

### The Advantages of Using ECC in NHP

The use of ECC within the NHP zero-trust protocol offers numerous advantages that make it well-suited to its security objectives:

1. **Scalable Security**: ECC's smaller key sizes provide strong security, which scales well with the increasing computational power of adversaries. With NHP's goal of providing a zero-trust environment for diverse network deployments, ECC's scalability is a critical asset.

2. **Resource Efficiency**: ECC reduces the computational burden on network devices compared to traditional public key cryptography. In environments where network resources may be constrained—such as edge devices or IoT components—this efficiency is essential for maintaining high performance without sacrificing security.

3. **Improved Performance**: The combination of ECDH for key exchange, ECDSA for authentication, and efficient symmetric encryption provides a balanced solution for secure communications. This balanced approach allows NHP to achieve the goals of zero trust while keeping latency low, which is crucial in time-sensitive network applications.

### Conclusion

The integration of Elliptic Curve Cryptography into the NHP Zero Trust Protocol provides a powerful means of securing network communications with minimal performance impact. By leveraging ECDH for secure key exchanges, ECDSA for robust authentication, and efficient symmetric encryption for data transfer, ECC supports the zero-trust model's goals of concealing network infrastructure, ensuring only trusted entities can access resources, and maintaining security with low overhead.

As cyber threats become more sophisticated, leveraging advanced cryptographic techniques like ECC in protocols like NHP is vital to staying ahead of attackers. The synergy between ECC and NHP not only helps protect critical network infrastructure but also ensures that security measures are both robust and efficient—a key combination for the success of any modern cybersecurity initiative.



## 2) Key Exchange, Data Encryption and Identity Verification
### 2.1 Introduction
The Network Infrastructure Hiding Protocol (NHP) is built around a zero-trust security model, ensuring secure communications even in the presence of potential attackers. To achieve this, NHP integrates the Noise Protocol Framework, a cryptographic framework designed for secure and flexible key exchange, data encryption, and identity verification. This combination provides robust security with minimal computational overhead.

### 2.2 Key Exchange with Noise Protocol
NHP utilizes the Noise Protocol's key exchange mechanism to ensure secure, authenticated communication channels between parties. The key exchange begins with a handshake phase where both communicating entities exchange Diffie-Hellman (DH) public keys. In Noise, each party generates an ephemeral key pair, and the exchanged keys are used to derive a shared secret, which is then used to encrypt the following communication.

Noise allows NHP to support both long-term static keys and ephemeral keys for enhanced security. The flexibility of the Noise framework's handshake patterns enables NHP to customize how the handshake occurs based on the specific use case, providing options for mutual authentication, anonymous initiators, or encryption of the initial handshake itself. By leveraging Noise's simple yet powerful token-based handshake system, NHP can precisely control the sequence of key exchange messages while keeping identity information confidential.

### 2.3 Data Encryption
Once the shared key is derived during the handshake, the Noise framework uses symmetric encryption to secure data. NHP takes advantage of the Noise CipherState and SymmetricState objects, which are core components of Noise's state machine, to manage encryption and decryption keys for the communication session.

In particular, the shared key is used to initialize a symmetric encryption key (k) along with a nonce (n) for encrypting data. Noise supports advanced encryption schemes like ChaCha20-Poly1305 or AESGCM, providing authenticated encryption with associated data (AEAD) to maintain data confidentiality and integrity. The chaining key (ck) and the handshake hash (h) are used to continuously derive fresh keys during the session, enhancing the forward secrecy and ensuring that a compromise of one key does not jeopardize other parts of the communication.

NHP benefits from these cryptographic properties by providing encrypted tunnels for network data, ensuring that any intercepted data cannot be decrypted without knowledge of the derived keys, which are securely exchanged during the handshake.

### 2.4 Identity Verification
Noise provides mechanisms for identity verification by combining the exchange of static keys with Diffie-Hellman operations. In NHP, identity verification occurs during the handshake, where static keys are encrypted and verified through shared DH operations, effectively binding the public keys of both parties to the derived session key.

During the handshake, Noise uses tokens such as "s" (static) and "e" (ephemeral) to indicate which keys are being exchanged and verified. This token-based approach allows NHP to selectively authenticate one or both parties depending on the specific use case. For example, the "XX" pattern in Noise provides mutual authentication, while the "NK" pattern allows for a one-sided authenticated handshake, giving NHP flexibility in how strictly identity verification is enforced.

To further protect identity information, Noise can encrypt static keys during the handshake. NHP leverages this feature to prevent an eavesdropper from discovering the identities of the participants, thus supporting the zero-trust model by ensuring that the identity of any participant is revealed only to the intended counterpart and not to third parties.

### 2.5 Algorithms and Formulas
The cryptographic strength of NHP's integration with the Noise Protocol Framework is built on the use of well-defined algorithms and mathematical formulas. Here, we provide an overview of the key algorithms and their corresponding formulas that are used in NHP for key exchange, encryption, and identity verification.

#### 2.5.1 Diffie-Hellman Key Exchange
The Diffie-Hellman (DH) key exchange is used to derive a shared secret between two parties, \( A \) and \( B \). Each party generates a private key (\( a \) for \( A \), \( b \) for \( B \)) and computes a public key by exponentiating a common generator \( g \) to their private key in a finite cyclic group of prime order \( p \):

- \( A \) computes its public key: \( A_{pub} = g^a \mod p \)
- \( B \) computes its public key: \( B_{pub} = g^b \mod p \)

The shared secret \( s \) is then computed by both parties using the other party's public key:

- \( A \) computes: \( s = B_{pub}^a \mod p \)
- \( B \) computes: \( s = A_{pub}^b \mod p \)

The resulting shared secret \( s \) is identical for both parties and is used to derive encryption keys.

#### 2.5.2 Symmetric Encryption
NHP uses symmetric encryption for data confidentiality. The key \( k \) and nonce \( n \) are used in the encryption function. For authenticated encryption with associated data (AEAD), the ChaCha20-Poly1305 algorithm is commonly used, which combines a stream cipher (ChaCha20) for encryption and a MAC (Poly1305) for authentication.

- Encryption: \( c = 	ext{ChaCha20}(k, n, 	ext{plaintext}) \)
- Authentication: \( 	ext{tag} = 	ext{Poly1305}(k, 	ext{associated data} || c) \)

The ciphertext \( c \) and the tag are transmitted together, ensuring both confidentiality and integrity.

#### 2.5.3 Key Derivation and Hashing
Noise uses a key derivation function (KDF) based on HMAC (Hash-based Message Authentication Code) to derive keys. The HKDF (HMAC-based Key Derivation Function) is used to produce multiple keys from the shared secret \( s \).

- HKDF steps:
  - \( 	ext{temp\_key} = 	ext{HMAC}(	ext{chaining\_key}, 	ext{input\_key\_material}) \)
  - \( 	ext{output1} = 	ext{HMAC}(	ext{temp\_key}, 0x01) \)
  - \( 	ext{output2} = 	ext{HMAC}(	ext{temp\_key}, 	ext{output1} || 0x02) \)

The derived keys are used for encryption and maintaining the chaining key (\( ck \)) that evolves with each message, ensuring forward secrecy.

#### 2.5.4 Identity Verification
Identity verification in NHP involves using static and ephemeral keys to authenticate parties. The Diffie-Hellman operations between static (\( s \)) and ephemeral (\( e \)) keys produce unique shared values that verify the identity of the participants.

- For identity verification, a combination of DH operations is performed:
  - \( 	ext{ss} = DH(s_A, s_B) \)
  - \( 	ext{es} = DH(e_A, s_B) \) or \( DH(s_A, e_B) \)
  - \( 	ext{ee} = DH(e_A, e_B) \)

These values are hashed together to derive the final session key, effectively binding the identities to the key exchange process and ensuring that only the intended parties can derive the correct session key.

### 2.6 Summary
NHP's implementation of the Noise Protocol Framework strengthens its zero-trust architecture by leveraging robust and well-tested cryptographic mechanisms for key exchange, data encryption, and identity verification. The modular nature of Noise allows NHP to adapt the handshake and encryption process based on the threat model, providing a high level of security against active and passive attacks. By incorporating Noise, NHP can maintain secure and authenticated communication channels while hiding the network infrastructure from attackers, achieving its goal of protecting network resources in hostile environments.

## 3) Key Distribution and Management

### 3.1 Introduction

Certificateless Public Key Cryptography, originally introduced by Al-Riyami and Paterson in 2003, provides a hybrid solution that eliminates the need for a conventional Certificate Authority (CA) while ensuring strong cryptographic assurances. This article explores how NHP utilizes CL-PKC for efficient and secure key management without relying on certificates, which are typically seen as a vulnerability in many cryptographic systems.

In traditional Public Key Infrastructure (PKI), certificate authorities (CAs) serve as trusted third parties that issue and manage public key certificates to verify the authenticity of users' keys. While effective, this model introduces complexities and risks, such as dependency on the CA and exposure to attacks targeting these central entities. Certificateless Public Key Cryptography aims to mitigate these issues by eliminating the use of certificates while still ensuring the authenticity of public keys.

In a CL-PKC system, a trusted third party called the Key Generation Center (KGC) is responsible for generating partial private keys for users. However, unlike a CA, the KGC does not have access to the complete private keys, which makes it impossible for the KGC to impersonate users. Each user combines the partial key from the KGC with their own secret value to generate their full private key and public key. This approach reduces the trust placed on any single entity and provides an extra layer of security.

### 3.2 Advantages of Using CL-PKC in NHP

1. **Reduced Trust Requirements**: Unlike PKI, where trust in the CA is critical, CL-PKC reduces this trust requirement. The KGC cannot generate complete private keys on its own, meaning it cannot impersonate users or decrypt their communications.

2. **Simplified Key Distribution**: There is no need for users to request or renew certificates, which eliminates many administrative burdens associated with traditional PKI.

3. **Resistance to Key Compromise**: Since the user's full private key is generated in part by the user, compromising the KGC does not allow an adversary to fully recover user keys. This mitigates the impact of a successful attack on the key distribution infrastructure.

4. **Scalability**: The certificateless nature of the system removes the need for managing large certificate databases, which simplifies scalability. This is particularly useful for IoT and other large-scale deployments where the overhead of certificate management would be prohibitive.

### 3.3 Key Management in NHP Using Certificateless Cryptography

The NHP Zero Trust protocol integrates CL-PKC to manage the distribution and verification of keys for its secure communication framework. Below, we explain how the various mechanisms of CL-PKC contribute to the key management process in NHP.

#### 3.3.1 Key Generation and Distribution

In NHP, the Key Generation Center (KGC) is responsible for creating system-wide parameters, including the master public-private key pair. The master private key is kept confidential by the KGC, while the master public key is distributed to all participants. When a new user wants to join the network, the KGC performs the following steps:

1. **Partial Key Generation**: The KGC generates a partial private key for the user using their unique identifier (e.g., an email or other identity information). This ensures that each user’s partial key is bound to their identity, providing identity-based security.

2. **User-Specific Key Pair Generation**: The user then selects their own secret value and combines it with the partial private key from the KGC to generate their full private key. The public key is computed from the combined secret, which means that while the KGC contributes to the key generation, it does not possess the complete private key.

This key distribution method ensures that the KGC cannot unilaterally determine a user's private key, mitigating the risks associated with compromised key generation authorities. Additionally, the lack of a need for traditional certificates means that users do not need to rely on external certificate authorities to validate keys, reducing the attack surface for man-in-the-middle (MITM) attacks.

#### 3.3.2 Public Key Verification Without Certificates

In certificateless systems like the one implemented in NHP, the authenticity of public keys is verified through implicit methods, rather than relying on certificates signed by a CA. Specifically, the user's public key is computed using the system parameters, the user's identifier, and the KGC's master public key. This computation is deterministic and allows any party to verify the authenticity of a public key without needing to trust a CA or store a large database of certificates.

By removing the need for traditional certificates, NHP is able to streamline the key verification process, eliminating the need for certificate revocation lists (CRLs) and other PKI complexities. This approach not only reduces the communication overhead but also enhances the security by removing dependencies on a trusted third party that could be targeted by attackers.

### 3.4 Algorithms in CL-PKC for NHP

To provide a clearer understanding of how the NHP protocol leverages Certificateless Public Key Cryptography, we describe the key algorithms involved along with their respective formulas.

#### 3.4.1 System Parameter Generation

The Key Generation Center (KGC) is responsible for generating the system parameters that will be used across the network. These parameters include an elliptic curve \( E \) defined over a finite field \( \mathbb{F}_q \), a base point \( G \) of prime order \( n \), and the master secret key \( ms \). The KGC computes the master public key \( P_{pub} \) as:

\[
P_{pub} = [ms]G
\]

where \( [ms]G \) denotes scalar multiplication of the base point \( G \) by the master secret \( ms \).

#### 3.4.2 Partial Private Key Generation

For each user with a unique identifier \( ID_A \), the KGC generates a partial private key. First, the KGC computes a hash value \( H_A \) based on the user's identifier and system parameters:

\[
H_A = H(ENTL_A \parallel ID_A \parallel a \parallel b \parallel x_G \parallel y_G \parallel x_{P_{pub}} \parallel y_{P_{pub}})
\]

where \( ENTL_A \) is a length value derived from the identifier, and \( (x_G, y_G) \) and \( (x_{P_{pub}}, y_{P_{pub}}) \) are the coordinates of points \( G \) and \( P_{pub} \), respectively.

The KGC then selects a random value \( w \in [1, n-1] \) and computes:

\[
W_A = [w]G + U_A
\]

where \( U_A = [d'_A]G \) is a point generated by the user with their own secret value \( d'_A \).

The partial private key \( t_A \) is computed as:

\[
t_A = (w + l \cdot ms) \mod n
\]

where \( l \) is a hash value computed from the point \( W_A \) and the hash \( H_A \).

#### 3.4.3 User Full Private Key Generation

The user generates their full private key \( d_A \) by combining the partial private key \( t_A \) with their secret value \( d'_A \):

\[
d_A = (t_A + d'_A) \mod n
\]

This ensures that only the user knows their complete private key.

#### 3.4.4 Public Key Computation

The user’s public key \( P_A \) is computed as:

\[
P_A = W_A + [l]P_{pub}
\]

This public key can be verified by anyone using the system parameters, the user’s identifier, and the KGC’s public key.

#### 3.4.5 Signature Generation and Verification

To generate a digital signature on a message \( M \), the user computes a hash \( e \) as follows:

\[
e = H(H_A \parallel x_{W_A} \parallel y_{W_A} \parallel M)
\]

The signature \( (r, s) \) is generated using the user’s private key \( d_A \) and a random value \( k \):

\[
[r]G = (x_1, y_1)
\]
\[
r = x_1 \mod n
\]
\[
s = (k^{-1}(e + d_A \cdot r)) \mod n
\]

To verify the signature, the verifier computes \( P_A \) and then checks whether:

\[
[r]G = [s]G + [e + r]P_A
\]

If the equality holds, the signature is valid.

### 3.5 Conclusion

NHP’s implementation of Certificateless Public Key Cryptography provides a powerful and efficient approach to key management in Zero Trust environments. By leveraging CL-PKC, NHP is able to mitigate the risks associated with traditional PKI, reduce the reliance on centralized trusted authorities, and simplify the key distribution process. The result is a more secure and scalable system that is well-suited for protecting critical network infrastructure in the face of evolving cyber threats.

The combination of certificateless cryptography and the Zero Trust principles of NHP makes it an attractive solution for securing network resources while minimizing the risks introduced by centralized authorities.


