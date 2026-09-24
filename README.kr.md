[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/main/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/main/README.zh-cn.md)
[![zh-tw](https://img.shields.io/badge/lang-zh--tw-green.svg)](https://github.com/OpenNHP/opennhp/blob/main/README.zh-tw.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/main/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/main/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/main/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/main/README.es.md)
[![id](https://img.shields.io/badge/lang-id-green.svg)](https://github.com/OpenNHP/opennhp/blob/main/README.id.md)
[![kr](https://img.shields.io/badge/lang-kr-green.svg)](https://github.com/OpenNHP/opennhp/blob/main/README.kr.md)

![OpenNHP Logo](docs/images/logo11.png)

# OpenNHP: 오픈소스 제로 트러스트 보안 툴킷

[![Build](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml/badge.svg)](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml)
[![Release](https://img.shields.io/github/v/tag/OpenNHP/opennhp?label=release)](https://github.com/OpenNHP/opennhp/tags)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
[![codecov](https://codecov.io/gh/OpenNHP/opennhp/branch/main/graph/badge.svg)](https://codecov.io/gh/OpenNHP/opennhp)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/OpenNHP/opennhp)

**OpenNHP**는 인프라, 애플리케이션, 데이터를 위한 제로 트러스트 보안을 구현하는 경량의 암호 기술 기반 오픈소스 툴킷입니다. [**클라우드 보안 연합(CSA)**](https://cloudsecurityalliance.org/)의 *[네트워크 인프라 은닉 프로토콜(NHP) 사양](https://cloudsecurityalliance.org/artifacts/stealth-mode-sdp-for-zero-trust-network-infrastructure)*에 대한 참조 구현이며, 두 가지 핵심 프로토콜을 제공합니다.

- **네트워크 인프라 은닉 프로토콜(NHP, Network-infrastructure Hiding Protocol):** 서버 포트, IP 주소, 도메인 이름을 숨겨 애플리케이션과 인프라를 무단 접근으로부터 보호합니다.
- **데이터 콘텐츠 은닉 프로토콜(DHP, Data-content Hiding Protocol):** 암호화와 기밀 컴퓨팅을 통해 데이터 보안과 프라이버시를 보장하여, 데이터를 *"사용할 수는 있지만 볼 수는 없게"* 만듭니다.

**[웹사이트](https://opennhp.org) · [비전](https://opennhp.org/vision/) · [라이브 데모](https://opennhp.org/demo/) · [문서](https://docs.opennhp.org) · [Discord](https://discord.gg/CpyVmspx5x)**

---

## 왜 OpenNHP인가

현대 인터넷은 [어두운 숲](https://en.wikipedia.org/wiki/Dark_forest_hypothesis)입니다. 공격자들은 [자율 취약점 공격(Autonomous Vulnerability Exploitation)](https://arxiv.org/abs/2404.08144)을 통해 기계 속도로 스캔, 핑거프린팅, 공격을 수행하는 LLM의 지원을 점점 더 많이 받으며, 접근 가능한 모든 서비스를 표적으로 삼습니다. [Gartner는](https://www.gartner.com/en/newsroom/press-releases/2024-08-28-gartner-forecasts-global-information-security-spending-to-grow-15-percent-in-2025) AI 기반 사이버 공격이 빠르게 증가할 것으로 전망합니다. 기존 방어 체계는 네트워크가 사용자를 들여보낸 *후에* 인증을 수행하기 때문에, 노출된 포트, IP, 도메인이 영구적인 공격 표면으로 남게 됩니다.

> **AI 시대에는 가시성 = 취약성(VISIBILITY = VULNERABILITY)입니다.**

OpenNHP는 이 모델을 뒤집습니다: **신뢰되기 전까지는 보이지 않습니다.** 모든 포트, IP, 호스트 이름은 기본 거부(default-deny) 게이트 뒤에 위치합니다. 접근은 암호학적으로 서명된 노크(knock)가 대역 외(out-of-band)에서 인증 및 인가된 후에만 허용됩니다. 공격자는 발견할 수 없는 것을 공격할 수 없습니다.

### 3세대 네트워크 은닉 프로토콜

NHP는 "서비스를 먼저 숨긴다"는 설계 계보의 다음 단계입니다.

| 세대 | 프로토콜 | 한계 |
|---|---|---|
| 1 | 포트 노킹(Port Knocking) | 평문 전송, 재전송 공격에 취약 |
| 2 | 단일 패킷 인가(SPA, Single Packet Authorization) | 공유 비밀키, 단방향, 일반적으로 포트만 숨김, 주로 C/C++ |
| **3** | **NHP** | 최신 암호 기술, 상태 정보를 포함한 양방향 통신, 도메인 + IP + 포트 은닉, 무상태(stateless) 및 수평 확장 가능, 메모리 안전한 Go |

NHP는 기존 IAM, DNS, FIDO, 제로 트러스트 정책 엔진을 대체하지 않고 함께 동작합니다. 즉, 기존 스택을 갈아엎는 것이 아니라 확장합니다.

---

## 아키텍처

OpenNHP는 [NIST 제로 트러스트 아키텍처](https://www.nist.gov/publications/zero-trust-architecture)에서 영감을 받아 세 가지 핵심 구성 요소로 이루어진 모듈식 설계를 따릅니다.

![OpenNHP architecture](docs/images/OpenNHP_Arch.gif)

| 핵심 구성 요소 | 역할 |
|-----------|------|
| **NHP-Agent** | 접근 권한을 얻기 위해 암호화된 노크 요청을 보내는 클라이언트 |
| **NHP-Server** | 요청을 인증하고 인가함. 별도로 실행되며 보호 대상 호스트와 아키텍처적으로 분리됨 |
| **NHP-AC** | 보호 대상 서버의 방화벽 규칙을 관리하는 접근 제어기 |

| 추가 구성 요소 | 역할 |
|-----------|------|
| **NHP-Relay** | 브라우저 기반 에이전트가 HTTPS를 통해 NHP 노크를 보낼 수 있게 하는 HTTP-to-UDP 브리지 |
| **NHP-KGC** | 신원 기반 암호(IBC, Identity-Based Cryptography)를 위한 키 생성 센터 |

### 프로토콜 흐름

1. Agent가 암호화된 노크(`NHP_KNK`)를 Server로 보냅니다.
2. Server가 노크를 검증하고 AC에 작업 요청(`NHP_AOP`)을 보냅니다.
3. AC가 방화벽을 열고 Server에 응답(`NHP_ART`)합니다.
4. Server가 접근 정보가 담긴 확인 응답(`NHP_ACK`)을 Agent에 반환합니다.
5. Agent가 AC를 통해 보호된 리소스에 접근합니다.

### 암호 기술

OpenNHP는 상호 교체 가능한 두 가지 암호 스위트를 제공합니다.

- **`CIPHER_SCHEME_CURVE`** — Curve25519 + AES-256-GCM + BLAKE2s
- **`CIPHER_SCHEME_GMSM`** — SM2 + SM4-GCM + SM3

두 스위트 모두 [Noise Protocol Framework](https://noiseprotocol.org/)를 기반으로 동작합니다. 신원 기반 암호(IBC) 모드는 키 생성 센터(KGC)를 통해 사용할 수 있습니다.

> 프로토콜 세부 사항, 배포 모델, 암호 설계에 대해서는 [문서](https://docs.opennhp.org)를 참고하세요.

---

## 저장소 구조

```
opennhp/
├── nhp/              # 핵심 프로토콜 라이브러리 (Go 모듈)
│   ├── core/         # 패킷 처리, 암호화, Noise Protocol, 디바이스 관리
│   ├── common/       # 공유 타입 및 메시지 정의
│   ├── utils/        # 유틸리티 함수
│   ├── plugins/      # 플러그인 핸들러 인터페이스
│   ├── log/          # 로깅 인프라
│   └── etcd/         # 분산 구성 지원
└── endpoints/        # 데몬 구현 (Go 모듈, nhp에 의존)
    ├── agent/        # NHP-Agent 데몬
    ├── server/       # NHP-Server 데몬
    ├── ac/           # NHP-AC (접근 제어기) 데몬
    ├── db/           # NHP-DB (DHP용 데이터 브로커)
    ├── kgc/          # NHP-KGC (키 생성 센터)
    └── relay/        # NHP-Relay 데몬
```

---

## 빠른 시작

### 사전 요구 사항

- Go 1.26+
- `make`
- Docker 및 Docker Compose (전체 스택 데모용)

### 빌드

```bash
# 모든 구성 요소 빌드
make

# 개별 데몬 빌드
make agentd    # NHP-Agent
make serverd   # NHP-Server
make acd       # NHP-AC
make db        # NHP-DB
make relayd    # NHP-Relay
make kgc       # NHP-KGC

```

### 테스트

```bash
cd nhp && go test ./...
cd endpoints && go test ./...
```

### Docker로 실행

```bash
cd docker && docker-compose up --build
```

[빠른 시작 튜토리얼](https://docs.opennhp.org/nhp_quick_start/)을 따라 Docker 환경에서 전체 인증 워크플로를 시뮬레이션해 보세요.

---

## 기여하기

기여를 환영합니다! 풀 리퀘스트를 제출하기 전에 [CONTRIBUTING.md](CONTRIBUTING.md)를 읽어 주세요.

**참고:** 모든 커밋은 검증된 GPG 또는 SSH 키로 서명되어야 합니다.

```bash
git commit -S -m "your message"
```

---

## 보안

취약점을 발견하셨나요? 공개 이슈를 여는 대신 [SECURITY.md](SECURITY.md)에 안내된 책임 있는 공개(responsible disclosure) 절차를 따라 주세요.

---

## 후원사

<a href="https://layerv.ai">
  <img src="docs/images/layerv_logo.png" height="40" alt="LayerV.ai logo">
</a>
&nbsp;&nbsp;
<a href="https://www.atlascloud.ai/">
  <img src="docs/images/atlascloud_logo.png" height="40" alt="Atlas Cloud logo">
</a>
&nbsp;&nbsp;
<a href="https://cloud.tencent.com/">
  <img src="docs/images/tencentcloud_logo.svg" height="40" alt="Tencent Cloud logo">
</a>

---

## 라이선스

[Apache 2.0 라이선스](LICENSE)에 따라 배포됩니다.

## 연락처

- 이메일: [support@opennhp.org](mailto:support@opennhp.org)
- Discord: [Discord 참여하기](https://discord.gg/CpyVmspx5x)
- 웹사이트: [https://opennhp.org](https://opennhp.org)
