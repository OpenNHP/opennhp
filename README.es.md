[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![zh-tw](https://img.shields.io/badge/lang-zh--tw-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-tw.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![OpenNHP Logo](docs/images/logo11.png)

# OpenNHP: Kit de herramientas de seguridad Zero Trust de código abierto

[![Build](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml/badge.svg)](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml)
[![Release](https://img.shields.io/github/v/tag/OpenNHP/opennhp?label=release)](https://github.com/OpenNHP/opennhp/tags)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
[![codecov](https://codecov.io/gh/OpenNHP/opennhp/branch/main/graph/badge.svg)](https://codecov.io/gh/OpenNHP/opennhp)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/OpenNHP/opennhp)

**OpenNHP** es un kit de herramientas ligero, de código abierto e impulsado por criptografía, que implementa seguridad Zero Trust para infraestructuras, aplicaciones y datos. Es la implementación de referencia de la *[especificación Network-infrastructure Hiding Protocol (NHP)](https://cloudsecurityalliance.org/artifacts/stealth-mode-sdp-for-zero-trust-network-infrastructure)* de la [**Cloud Security Alliance (CSA)**](https://cloudsecurityalliance.org/) e incluye dos protocolos principales:

- **Network-infrastructure Hiding Protocol (NHP):** oculta puertos del servidor, direcciones IP y nombres de dominio para proteger aplicaciones e infraestructura frente a accesos no autorizados.
- **Data-content Hiding Protocol (DHP):** garantiza la seguridad y la privacidad de los datos mediante cifrado y confidential computing, haciendo que los datos sean *«utilizables pero no visibles»*.

**[Sitio web](https://opennhp.org) · [Visión](https://opennhp.org/vision/) · [Demo en vivo](https://opennhp.org/demo/) · [Documentación](https://docs.opennhp.org) · [Discord](https://discord.gg/CpyVmspx5x)**

---

## Por qué OpenNHP

La internet moderna es un [bosque oscuro](https://en.wikipedia.org/wiki/Dark_forest_hypothesis). Los atacantes —respaldados cada vez más por LLMs que escanean, toman huellas y explotan a velocidad de máquina mediante [Autonomous Vulnerability Exploitation](https://arxiv.org/abs/2404.08144)— tratan todo servicio alcanzable como un objetivo. [Gartner prevé](https://www.gartner.com/en/newsroom/press-releases/2024-08-28-gartner-forecasts-global-information-security-spending-to-grow-15-percent-in-2025) un rápido aumento de los ciberataques impulsados por IA. Las defensas tradicionales autentican a los usuarios *después* de que la red los deja entrar, dejando los puertos, IPs y dominios expuestos como una superficie de ataque permanente.

> **En la era de la IA, VISIBILIDAD = VULNERABILIDAD.**

OpenNHP invierte ese modelo: **invisible hasta la confianza**. Cada puerto, IP y nombre de host se sitúa detrás de una puerta de denegación por defecto. El acceso se concede solo después de que un «golpe en la puerta» firmado criptográficamente haya sido autenticado y autorizado fuera de banda. Los atacantes no pueden explotar lo que no pueden descubrir.

### El protocolo de ocultación de red de tercera generación

NHP es el siguiente paso en la línea de diseños «ocultar el servicio primero»:

| Generación | Protocolo | Limitaciones |
|---|---|---|
| 1 | Port Knocking | Texto plano, vulnerable a repetición |
| 2 | Single Packet Authorization (SPA) | Secretos compartidos, unidireccional, normalmente oculta solo puertos, típicamente en C/C++ |
| **3** | **NHP** | Criptografía moderna, bidireccional con estado, oculta dominio + IP + puertos, sin estado y escalable horizontalmente, Go con seguridad de memoria |

NHP se integra junto a los motores IAM, DNS, FIDO y de política Zero Trust existentes, en lugar de reemplazarlos: extiende tu stack en vez de bifurcarlo.

---

## Arquitectura

OpenNHP sigue un diseño modular con tres componentes principales, inspirado en la [Arquitectura Zero Trust del NIST](https://www.nist.gov/publications/zero-trust-architecture):

![OpenNHP architecture](docs/images/OpenNHP_Arch.gif)

| Componente principal | Rol |
|-----------|------|
| **NHP-Agent** | Cliente que envía solicitudes «knock» cifradas para obtener acceso |
| **NHP-Server** | Autentica y autoriza las solicitudes; se ejecuta por separado y está arquitectónicamente desacoplado del host protegido |
| **NHP-AC** | Controlador de acceso que gestiona las reglas del cortafuegos en el servidor protegido |

| Componente adicional | Rol |
|-----------|------|
| **NHP-Relay** | Puente HTTP a UDP que permite a los agentes basados en navegador enviar knocks NHP a través de HTTPS |
| **NHP-KGC** | Centro de generación de claves para criptografía basada en identidad (IBC) |

### Flujo del protocolo

1. El Agent envía un knock cifrado (`NHP_KNK`) al Server.
2. El Server valida el knock y envía una solicitud de operación (`NHP_AOP`) al AC.
3. El AC abre el cortafuegos y responde (`NHP_ART`) al Server.
4. El Server devuelve un reconocimiento (`NHP_ACK`) con la información de acceso al Agent.
5. El Agent alcanza el recurso protegido a través del AC.

### Criptografía

OpenNHP incluye dos suites criptográficas intercambiables:

- **`CIPHER_SCHEME_CURVE`** — Curve25519 + AES-256-GCM + BLAKE2s
- **`CIPHER_SCHEME_GMSM`** — SM2 + SM4-GCM + SM3

Ambas se basan en el [Noise Protocol Framework](https://noiseprotocol.org/). Un modo de criptografía basada en identidad (IBC) está disponible a través del Key Generation Center (KGC).

> Para detalles del protocolo, modelos de despliegue y diseño criptográfico, consulta la [documentación](https://docs.opennhp.org).

---

## Estructura del repositorio

```
opennhp/
├── nhp/              # Biblioteca principal del protocolo (módulo Go)
│   ├── core/         # Manejo de paquetes, criptografía, protocolo Noise, gestión de dispositivos
│   ├── common/       # Tipos compartidos y definiciones de mensajes
│   ├── utils/        # Funciones utilitarias
│   ├── plugins/      # Interfaces de manejadores de plugins
│   ├── log/          # Infraestructura de logging
│   └── etcd/         # Soporte de configuración distribuida
└── endpoints/        # Implementaciones de daemons (módulo Go, depende de nhp)
    ├── agent/        # Daemon NHP-Agent
    ├── server/       # Daemon NHP-Server
    ├── ac/           # Daemon NHP-AC (controlador de acceso)
    ├── db/           # NHP-DB (Data Broker para DHP)
    ├── kgc/          # NHP-KGC (Key Generation Center)
    └── relay/        # Daemon NHP-Relay
```

---

## Inicio rápido

### Requisitos previos

- Go 1.25.6+
- `make`
- Docker y Docker Compose (para la demo completa)

### Compilación

```bash
# Compilar todos los componentes
make

# Compilar daemons individuales
make agentd    # NHP-Agent
make serverd   # NHP-Server
make acd       # NHP-AC
make db        # NHP-DB
make relayd    # NHP-Relay
make kgc       # NHP-KGC
```

### Pruebas

```bash
cd nhp && go test ./...
cd endpoints && go test ./...
```

### Ejecución con Docker

```bash
cd docker && docker-compose up --build
```

Sigue el [tutorial de inicio rápido](https://docs.opennhp.org/nhp_quick_start/) para simular el flujo completo de autenticación en un entorno Docker.

---

## Contribuir

¡Las contribuciones son bienvenidas! Por favor, lee [CONTRIBUTING.md](CONTRIBUTING.md) antes de enviar un Pull Request.

**Nota:** todos los commits deben estar firmados con una clave GPG o SSH verificada.

```bash
git commit -S -m "your message"
```

---

## Seguridad

¿Has encontrado una vulnerabilidad? Sigue el proceso de divulgación responsable descrito en [SECURITY.md](SECURITY.md) en lugar de abrir un issue público.

---

## Patrocinadores

<a href="https://layerv.ai">
  <img src="docs/images/layerv_logo.png" height="40" alt="LayerV.ai logo">
</a>

---

## Licencia

Publicado bajo la [Licencia Apache 2.0](LICENSE).

## Contacto

- Correo: [support@opennhp.org](mailto:support@opennhp.org)
- Discord: [Únete a nuestro Discord](https://discord.gg/CpyVmspx5x)
- Sitio web: [https://opennhp.org](https://opennhp.org)
