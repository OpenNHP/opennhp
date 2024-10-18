---
layout: page
title: Features
nav_order: 2
permalink: /features/
---

# OpenNHP Feature List
{: .fs-9 }

OpenNHP offers robust security, excellent performance, and scalability to protect your network resources.
{: .fs-6 .fw-300 }

[中文版](/zh-cn/features/){: .label .fs-4 }

---

- **Mitigate vulnerability risk:** The openness of TCP/IP protocols leads to a "trust by default" connection model, allowing anyone to establish a connection to a server port that provides services. Attackers exploit this openness to target server vulnerabilities. The NHP protocol implements the zero trust principle "never trust, always verify" by enforcing "deny-all" rules by default on the server side, only allowing authorized hosts to establish connections. This effectively mitigates vulnerability exploitation, particularly zero-day exploits.
- **Mitigate phishing attacks:** DNS hijacking is a serious threat to internet security and is used for malicious purposes such as phishing, stealing sensitive information, or spreading malware. The NHP protocol can function as an encrypted DNS resolution service to mitigate this problem. When the NHP-Agent on the client side sends a knock request to the controller component NHP-Server with the identifier (e.g., the domain name) of the protected resource, the NHP-Server will return the IP address and port number of the protected resource if the NHP-Agent is successfully authenticated. Since NHP communication is encrypted and mutually verified, the risk of DNS hijacking is effectively mitigated.
- **Mitigate DDoS attacks:** As mentioned above, a client cannot obtain the IP address and port number of protected resources without authentication. If the protected resources are distributed across multiple locations, the NHP server may return different IP addresses to different clients, making DDoS attacks significantly more difficult and expensive to execute.
- **Attack attribution:** The connection model of TCP/IP protocols is IP-based. With NHP, the connection model becomes identity (ID)-based. The connection initiator's identity must be authenticated before establishing the connection, making attacks much more identifiable and traceable. 
- **Default-deny access control**: All resources are hidden by default, only becoming accessible after authentication and authorization.
- **Identity and device-based authentication**: Ensures that only known users on approved devices can gain access.
- **Encrypted DNS resolution**: Prevents DNS hijacking and associated phishing attacks.
- **DDoS mitigation**: Distributed infrastructure design helps protect against Distributed Denial of Service attacks.
- **Scalable architecture**: Decoupled components allow for flexible deployment and scaling.
- **IAM integration**: Works with your existing Identity and Access Management systems.
- **Flexible deployment**: Supports various models including client-to-gateway, client-to-server, and more.
- **Strong cryptography**: Utilizes modern algorithms like ECC, Noise Protocol, and IBC for robust security.
- Mitigates vulnerability exploitation by enforcing "deny-all" rules by default
- Prevents phishing attacks through encrypted DNS resolution
- Protects against DDoS attacks by hiding infrastructure
- Enables attack attribution through identity-based connections
- Default-deny access control for all protected resources
- Identity and device-based authentication before network access
- Encrypted DNS resolution to prevent DNS hijacking
- Distributed infrastructure to mitigate DDoS attacks
- Scalable architecture with decoupled components
- Integration with existing identity and access management systems
- Support for various deployment models (client-to-gateway, client-to-server, etc)
- Cryptographically secure using modern algorithms (ECC, Noise Protocol, IBC)

