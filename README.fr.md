[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![zh-tw](https://img.shields.io/badge/lang-zh--tw-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-tw.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![OpenNHP Logo](docs/images/logo11.png)

# OpenNHP : Boîte à outils open source de sécurité Zero Trust

[![Build](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml/badge.svg)](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml)
[![Release](https://img.shields.io/github/v/tag/OpenNHP/opennhp?label=release)](https://github.com/OpenNHP/opennhp/tags)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
[![codecov](https://codecov.io/gh/OpenNHP/opennhp/branch/main/graph/badge.svg)](https://codecov.io/gh/OpenNHP/opennhp)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/OpenNHP/opennhp)

**OpenNHP** est une boîte à outils open source légère et basée sur la cryptographie, qui met en œuvre la sécurité Zero Trust pour les infrastructures, les applications et les données. C'est l'implémentation de référence de la *[spécification Network-infrastructure Hiding Protocol (NHP)](https://cloudsecurityalliance.org/artifacts/stealth-mode-sdp-for-zero-trust-network-infrastructure)* de la [**Cloud Security Alliance (CSA)**](https://cloudsecurityalliance.org/), et elle comprend deux protocoles principaux :

- **Network-infrastructure Hiding Protocol (NHP) :** dissimule les ports serveur, les adresses IP et les noms de domaine pour protéger les applications et l'infrastructure contre les accès non autorisés.
- **Data-content Hiding Protocol (DHP) :** assure la sécurité et la confidentialité des données grâce au chiffrement et au confidential computing, rendant les données *« utilisables mais invisibles »*.

**[Site web](https://opennhp.org) · [Vision](https://opennhp.org/vision/) · [Démo en direct](https://opennhp.org/demo/) · [Documentation](https://docs.opennhp.org) · [Discord](https://discord.gg/CpyVmspx5x)**

---

## Pourquoi OpenNHP

L'internet moderne est une [forêt sombre](https://en.wikipedia.org/wiki/Dark_forest_hypothesis). Les attaquants — de plus en plus soutenus par des LLM qui scannent, identifient et exploitent à la vitesse de la machine via l'[Autonomous Vulnerability Exploitation](https://arxiv.org/abs/2404.08144) — considèrent chaque service accessible comme une cible. [Gartner prévoit](https://www.gartner.com/en/newsroom/press-releases/2024-08-28-gartner-forecasts-global-information-security-spending-to-grow-15-percent-in-2025) une hausse rapide des cyberattaques pilotées par l'IA. Les défenses traditionnelles authentifient les utilisateurs *après* que le réseau les ait laissés entrer, laissant les ports, IP et domaines exposés comme une surface d'attaque permanente.

> **À l'ère de l'IA, VISIBILITÉ = VULNÉRABILITÉ.**

OpenNHP inverse ce modèle : **invisible jusqu'à la confiance**. Chaque port, IP et nom d'hôte est placé derrière une porte refusant tout par défaut. L'accès n'est accordé qu'après qu'un « toc-toc » cryptographiquement signé a été authentifié et autorisé hors bande. Les attaquants ne peuvent pas exploiter ce qu'ils ne peuvent pas découvrir.

### Le protocole de masquage réseau de troisième génération

NHP est la prochaine étape dans la lignée des conceptions « cacher le service d'abord » :

| Génération | Protocole | Limitations |
|---|---|---|
| 1 | Port Knocking | Texte clair, vulnérable au rejeu |
| 2 | Single Packet Authorization (SPA) | Secrets partagés, unidirectionnel, cache généralement uniquement les ports, souvent en C/C++ |
| **3** | **NHP** | Cryptographie moderne, bidirectionnel avec statut, cache domaine + IP + ports, sans état et scalable horizontalement, Go memory-safe |

NHP s'intègre aux moteurs IAM, DNS, FIDO et aux policy engines Zero Trust existants au lieu de les remplacer — il étend votre stack sans la forker.

---

## Architecture

OpenNHP adopte une conception modulaire avec trois composants principaux, inspirée de l'[architecture Zero Trust du NIST](https://www.nist.gov/publications/zero-trust-architecture) :

![OpenNHP architecture](docs/images/OpenNHP_Arch.gif)

| Composant principal | Rôle |
|-----------|------|
| **NHP-Agent** | Client qui envoie des requêtes « toc-toc » chiffrées pour obtenir l'accès |
| **NHP-Server** | Authentifie et autorise les requêtes ; s'exécute séparément et est architecturalement découplé de l'hôte protégé |
| **NHP-AC** | Contrôleur d'accès qui gère les règles de pare-feu sur le serveur protégé |

| Composant additionnel | Rôle |
|-----------|------|
| **NHP-Relay** | Pont HTTP vers UDP permettant aux agents basés sur navigateur d'envoyer des knocks NHP via HTTPS |
| **NHP-KGC** | Centre de génération de clés pour la cryptographie basée sur l'identité (IBC) |

### Flux protocolaire

1. L'Agent envoie un knock chiffré (`NHP_KNK`) au Server.
2. Le Server valide le knock et envoie une requête d'opération (`NHP_AOP`) à l'AC.
3. L'AC ouvre le pare-feu et répond (`NHP_ART`) au Server.
4. Le Server renvoie un acquittement (`NHP_ACK`) avec les informations d'accès à l'Agent.
5. L'Agent atteint la ressource protégée via l'AC.

### Cryptographie

OpenNHP fournit deux suites cryptographiques interchangeables :

- **`CIPHER_SCHEME_CURVE`** — Curve25519 + AES-256-GCM + BLAKE2s
- **`CIPHER_SCHEME_GMSM`** — SM2 + SM4-GCM + SM3

Toutes deux reposent sur le [Noise Protocol Framework](https://noiseprotocol.org/). Un mode Identity-Based Cryptography (IBC) est disponible via le Key Generation Center (KGC).

> Pour les détails du protocole, les modèles de déploiement et la conception cryptographique, consultez la [documentation](https://docs.opennhp.org).

---

## Structure du dépôt

```
opennhp/
├── nhp/              # Bibliothèque principale du protocole (module Go)
│   ├── core/         # Traitement des paquets, cryptographie, protocole Noise, gestion des périphériques
│   ├── common/       # Types partagés et définitions de messages
│   ├── utils/        # Fonctions utilitaires
│   ├── plugins/      # Interfaces des gestionnaires de plugins
│   ├── log/          # Infrastructure de journalisation
│   └── etcd/         # Support de configuration distribuée
└── endpoints/        # Implémentations des daemons (module Go, dépend de nhp)
    ├── agent/        # Daemon NHP-Agent
    ├── server/       # Daemon NHP-Server
    ├── ac/           # Daemon NHP-AC (contrôleur d'accès)
    ├── db/           # NHP-DB (Data Broker pour DHP)
    ├── kgc/          # NHP-KGC (Key Generation Center)
    └── relay/        # Daemon NHP-Relay
```

---

## Démarrage rapide

### Prérequis

- Go 1.25.6+
- `make`
- Docker et Docker Compose (pour la démo complète)

### Construction

```bash
# Construire tous les composants
make

# Construire les daemons individuellement
make agentd    # NHP-Agent
make serverd   # NHP-Server
make acd       # NHP-AC
make db        # NHP-DB
make relayd    # NHP-Relay
make kgc       # NHP-KGC
```

### Tests

```bash
cd nhp && go test ./...
cd endpoints && go test ./...
```

### Exécution avec Docker

```bash
cd docker && docker-compose up --build
```

Suivez le [tutoriel de démarrage rapide](https://docs.opennhp.org/nhp_quick_start/) pour simuler le workflow d'authentification complet dans un environnement Docker.

---

## Contribuer

Les contributions sont les bienvenues ! Veuillez lire [CONTRIBUTING.md](CONTRIBUTING.md) avant de soumettre une Pull Request.

**Remarque :** tous les commits doivent être signés avec une clé GPG ou SSH vérifiée.

```bash
git commit -S -m "your message"
```

---

## Sécurité

Vous avez trouvé une vulnérabilité ? Merci de suivre le processus de divulgation responsable décrit dans [SECURITY.md](SECURITY.md) plutôt que d'ouvrir un ticket public.

---

## Sponsors

<a href="https://layerv.ai">
  <img src="docs/images/layerv_logo.png" height="40" alt="LayerV.ai logo">
</a>

---

## Licence

Publié sous [licence Apache 2.0](LICENSE).

## Contact

- E-mail : [support@opennhp.org](mailto:support@opennhp.org)
- Discord : [Rejoindre notre Discord](https://discord.gg/CpyVmspx5x)
- Site web : [https://opennhp.org](https://opennhp.org)
