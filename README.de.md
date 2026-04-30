[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![zh-tw](https://img.shields.io/badge/lang-zh--tw-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-tw.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![OpenNHP Logo](docs/images/logo11.png)

# OpenNHP: Open-Source-Werkzeugkasten für Zero-Trust-Sicherheit

[![Build](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml/badge.svg)](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml)
[![Release](https://img.shields.io/github/v/tag/OpenNHP/opennhp?label=release)](https://github.com/OpenNHP/opennhp/tags)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
[![codecov](https://codecov.io/gh/OpenNHP/opennhp/branch/main/graph/badge.svg)](https://codecov.io/gh/OpenNHP/opennhp)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/OpenNHP/opennhp)

**OpenNHP** ist ein schlankes, kryptographiegetriebenes, quelloffenes Toolkit, das Zero-Trust-Sicherheit für Infrastruktur, Anwendungen und Daten umsetzt. Es ist die Referenzimplementierung der [**Cloud Security Alliance (CSA)**](https://cloudsecurityalliance.org/) *[Network-infrastructure Hiding Protocol (NHP) Spezifikation](https://cloudsecurityalliance.org/artifacts/stealth-mode-sdp-for-zero-trust-network-infrastructure)* und umfasst zwei Kernprotokolle:

- **Network-infrastructure Hiding Protocol (NHP):** Verbirgt Server-Ports, IP-Adressen und Domainnamen, um Anwendungen und Infrastruktur vor unbefugtem Zugriff zu schützen.
- **Data-content Hiding Protocol (DHP):** Sorgt für Datensicherheit und -vertraulichkeit durch Verschlüsselung und Confidential Computing, sodass Daten *„nutzbar, aber nicht sichtbar"* werden.

**[Website](https://opennhp.org) · [Vision](https://opennhp.org/vision/) · [Live-Demo](https://opennhp.org/demo/) · [Dokumentation](https://docs.opennhp.org) · [Discord](https://discord.gg/CpyVmspx5x)**

---

## Warum OpenNHP

Das moderne Internet ist ein [dunkler Wald](https://en.wikipedia.org/wiki/Dark_forest_hypothesis). Angreifer – zunehmend unterstützt durch LLMs, die mittels [Autonomous Vulnerability Exploitation](https://arxiv.org/abs/2404.08144) in Maschinengeschwindigkeit scannen, Fingerprinting betreiben und ausnutzen – behandeln jeden erreichbaren Dienst als Ziel. [Gartner prognostiziert](https://www.gartner.com/en/newsroom/press-releases/2024-08-28-gartner-forecasts-global-information-security-spending-to-grow-15-percent-in-2025), dass KI-gestützte Cyberangriffe rasant zunehmen werden. Herkömmliche Verteidigungsmaßnahmen authentifizieren Nutzer erst *nachdem* das Netzwerk sie hereingelassen hat, sodass offene Ports, IPs und Domains zur dauerhaften Angriffsfläche werden.

> **Im KI-Zeitalter gilt: SICHTBARKEIT = VERWUNDBARKEIT.**

OpenNHP dreht dieses Modell um: **unsichtbar bis vertrauenswürdig.** Jeder Port, jede IP und jeder Hostname liegt hinter einem Default-Deny-Gate. Zugriff wird erst nach einem kryptographisch signierten Knock gewährt, der außerhalb des Datenkanals authentifiziert und autorisiert wurde. Angreifer können nicht ausnutzen, was sie nicht entdecken können.

### Das Netzwerkverstecker-Protokoll der dritten Generation

NHP ist der nächste Schritt in einer Linie von „Dienst zuerst verstecken"-Entwürfen:

| Generation | Protokoll | Einschränkungen |
|---|---|---|
| 1 | Port Knocking | Klartext, anfällig für Replay-Angriffe |
| 2 | Single Packet Authorization (SPA) | Geteilte Geheimnisse, Einwegkommunikation, verbirgt typischerweise nur Ports, meist in C/C++ |
| **3** | **NHP** | Moderne Kryptographie, bidirektional mit Statusrückmeldung, verbirgt Domain + IP + Ports, zustandslos und horizontal skalierbar, speichersicheres Go |

NHP fügt sich neben bestehenden IAM-, DNS-, FIDO- und Zero-Trust-Policy-Engines ein, statt sie zu ersetzen – es erweitert Ihren Stack, statt ihn zu forken.

---

## Architektur

OpenNHP folgt einem modularen Design mit drei Kernkomponenten, inspiriert von der [NIST Zero Trust Architecture](https://www.nist.gov/publications/zero-trust-architecture):

![OpenNHP architecture](docs/images/OpenNHP_Arch.gif)

| Kernkomponente | Rolle |
|-----------|------|
| **NHP-Agent** | Client, der verschlüsselte Knock-Anfragen sendet, um Zugriff zu erhalten |
| **NHP-Server** | Authentifiziert und autorisiert Anfragen; läuft separat und ist architektonisch vom geschützten Host entkoppelt |
| **NHP-AC** | Zugriffs-Controller, der die Firewall-Regeln auf dem geschützten Server verwaltet |

| Zusatzkomponente | Rolle |
|-----------|------|
| **NHP-Relay** | HTTP-zu-UDP-Brücke, die browserbasierte Agents ermöglicht, NHP-Knocks über HTTPS zu senden |
| **NHP-KGC** | Key Generation Center für Identity-Based Cryptography (IBC) |

### Protokollablauf

1. Agent sendet einen verschlüsselten Knock (`NHP_KNK`) an den Server.
2. Server validiert den Knock und schickt eine Operations-Anfrage (`NHP_AOP`) an den AC.
3. AC öffnet die Firewall und antwortet (`NHP_ART`) dem Server.
4. Server schickt eine Bestätigung (`NHP_ACK`) mit Zugriffsinformationen an den Agent.
5. Agent erreicht die geschützte Ressource über den AC.

### Kryptographie

OpenNHP liefert zwei austauschbare Cipher-Suiten mit:

- **`CIPHER_SCHEME_CURVE`** – Curve25519 + AES-256-GCM + BLAKE2s
- **`CIPHER_SCHEME_GMSM`** – SM2 + SM4-GCM + SM3

Beide bauen auf dem [Noise Protocol Framework](https://noiseprotocol.org/) auf. Ein Identity-Based Cryptography (IBC)-Modus ist über das Key Generation Center (KGC) verfügbar.

> Für Protokolldetails, Bereitstellungsmodelle und kryptographisches Design siehe die [Dokumentation](https://docs.opennhp.org).

---

## Repository-Struktur

```
opennhp/
├── nhp/              # Kernprotokoll-Bibliothek (Go-Modul)
│   ├── core/         # Paketverarbeitung, Kryptographie, Noise-Protokoll, Geräteverwaltung
│   ├── common/       # Gemeinsame Typen und Nachrichtendefinitionen
│   ├── utils/        # Hilfsfunktionen
│   ├── plugins/      # Plugin-Handler-Schnittstellen
│   ├── log/          # Logging-Infrastruktur
│   └── etcd/         # Unterstützung für verteilte Konfiguration
└── endpoints/        # Daemon-Implementierungen (Go-Modul, abhängig von nhp)
    ├── agent/        # NHP-Agent-Daemon
    ├── server/       # NHP-Server-Daemon
    ├── ac/           # NHP-AC-(Zugriffs-Controller)-Daemon
    ├── db/           # NHP-DB (Data Broker für DHP)
    ├── kgc/          # NHP-KGC (Key Generation Center)
    └── relay/        # NHP-Relay-Daemon
```

---

## Schnellstart

### Voraussetzungen

- Go 1.25.6+
- `make`
- Docker und Docker Compose (für die Full-Stack-Demo)

### Bauen

```bash
# Alle Komponenten bauen
make

# Einzelne Daemons bauen
make agentd    # NHP-Agent
make serverd   # NHP-Server
make acd       # NHP-AC
make db        # NHP-DB
make relayd    # NHP-Relay
make kgc       # NHP-KGC
```

### Testen

```bash
cd nhp && go test ./...
cd endpoints && go test ./...
```

### Mit Docker starten

```bash
cd docker && docker-compose up --build
```

Folgen Sie dem [Schnellstart-Tutorial](https://docs.opennhp.org/nhp_quick_start/), um den vollständigen Authentifizierungs-Workflow in einer Docker-Umgebung zu simulieren.

---

## Mitwirken

Beiträge sind willkommen! Bitte lesen Sie [CONTRIBUTING.md](CONTRIBUTING.md), bevor Sie Pull Requests einreichen.

**Hinweis:** Alle Commits müssen mit einem verifizierten GPG- oder SSH-Schlüssel signiert sein.

```bash
git commit -S -m "your message"
```

---

## Sicherheit

Eine Schwachstelle gefunden? Bitte folgen Sie dem Responsible-Disclosure-Prozess in [SECURITY.md](SECURITY.md), statt ein öffentliches Issue zu eröffnen.

---

## Sponsoren

<a href="https://layerv.ai">
  <img src="docs/images/layerv_logo.png" height="40" alt="LayerV.ai logo">
</a>

---

## Lizenz

Veröffentlicht unter der [Apache 2.0 Lizenz](LICENSE).

## Kontakt

- E-Mail: [support@opennhp.org](mailto:support@opennhp.org)
- Discord: [Discord beitreten](https://discord.gg/CpyVmspx5x)
- Website: [https://opennhp.org](https://opennhp.org)
