[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![zh-tw](https://img.shields.io/badge/lang-zh--tw-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-tw.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![OpenNHP Logo](docs/images/logo11.png)

# OpenNHP：開源零信任安全工具套件

[![Build](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml/badge.svg)](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml)
[![Release](https://img.shields.io/github/v/tag/OpenNHP/opennhp?label=release)](https://github.com/OpenNHP/opennhp/tags)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
[![codecov](https://codecov.io/gh/OpenNHP/opennhp/branch/main/graph/badge.svg)](https://codecov.io/gh/OpenNHP/opennhp)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/OpenNHP/opennhp)

**OpenNHP** 是一套輕量級、以密碼學為基礎的開源工具套件，為基礎架構、應用程式與資料提供零信任安全保障。它是[**雲端安全聯盟（CSA）**](https://cloudsecurityalliance.org/) *[網路基礎架構隱藏協定（NHP）規範](https://cloudsecurityalliance.org/artifacts/stealth-mode-sdp-for-zero-trust-network-infrastructure)*的參考實作，包含兩項核心協定：

- **網路基礎架構隱藏協定（NHP）：** 隱藏伺服器連接埠、IP 位址與網域名稱，避免應用程式與基礎架構遭未授權存取。
- **資料內容隱藏協定（DHP）：** 透過加密與機密運算保障資料安全與隱私，使資料*「可用但不可見」*。

**[官方網站](https://opennhp.org) · [願景](https://opennhp.org/vision/) · [線上展示](https://opennhp.org/demo/) · [文件](https://docs.opennhp.org) · [Discord](https://discord.gg/CpyVmspx5x)**

---

<p align="center">
  <a href="https://www.atlascloud.ai/?utm_source=github&utm_medium=link&utm_campaign=opennhp">
    <img src="./assets/atlas-cloud-logo.png" alt="Atlas Cloud" width="200">
  </a>
</p>

> 🎁 **[Atlas Cloud](https://www.atlascloud.ai/?utm_source=github&utm_medium=link&utm_campaign=opennhp)** 贊助 OpenNHP 免費 API 額度 — 透過統一的 OpenAI 相容介面，使用 59+ 大型語言模型（Claude、GPT-5、Gemini、DeepSeek 等）為零信任安全自動化提供強大動力 — [coding plan](https://www.atlascloud.ai/console/coding-plan)

<details>
<summary>Atlas Cloud 全部 LLM 模型（59 個）</summary>

- Anthropic: `anthropic/claude-haiku-4.5-20251001`, `anthropic/claude-opus-4.8`, `anthropic/claude-sonnet-4.6`
- OpenAI: `openai/gpt-5.4`, `openai/gpt-5.5`
- Google Gemini: `google/gemini-3.1-flash-lite`, `google/gemini-3.1-pro-preview`, `google/gemini-3.5-flash`
- Qwen: `qwen/qwen2.5-7b-instruct`, `Qwen/Qwen3-235B-A22B-Instruct-2507`, `qwen/qwen3-235b-a22b-thinking-2507`, `qwen/qwen3-30b-a3b`, `Qwen/Qwen3-30B-A3B-Instruct-2507`, `qwen/qwen3-30b-a3b-thinking-2507`, `qwen/qwen3-32b`, `qwen/qwen3-8b`, `Qwen/Qwen3-Coder`, `qwen/qwen3-coder-next`, `qwen/qwen3-max-2026-01-23`, `Qwen/Qwen3-Next-80B-A3B-Instruct`, `Qwen/Qwen3-Next-80B-A3B-Thinking`, `Qwen/Qwen3-VL-235B-A22B-Instruct`, `qwen/qwen3-vl-235b-a22b-thinking`, `qwen/qwen3-vl-30b-a3b-instruct`, `qwen/qwen3-vl-30b-a3b-thinking`, `qwen/qwen3-vl-8b-instruct`, `qwen/qwen3.5-122b-a10b`, `qwen/qwen3.5-27b`, `qwen/qwen3.5-35b-a3b`, `qwen/qwen3.5-397b-a17b`, `qwen/qwen3.6-35b-a3b`, `qwen/qwen3.6-plus`
- DeepSeek: `deepseek-ai/deepseek-ocr`, `deepseek-ai/deepseek-r1-0528`, `deepseek-ai/DeepSeek-V3-0324`, `deepseek-ai/DeepSeek-V3.1`, `deepseek-ai/DeepSeek-V3.1-Terminus`, `deepseek-ai/deepseek-v3.2`, `deepseek-ai/DeepSeek-V3.2-Exp`, `deepseek-ai/deepseek-v4-flash`, `deepseek-ai/deepseek-v4-pro`
- Kimi: `moonshotai/Kimi-K2-Instruct`, `moonshotai/Kimi-K2-Instruct-0905`, `moonshotai/Kimi-K2-Thinking`, `moonshotai/kimi-k2.5`, `moonshotai/kimi-k2.6`
- GLM: `zai-org/GLM-4.6`, `zai-org/glm-4.7`, `zai-org/glm-5`, `zai-org/glm-5-turbo`, `zai-org/glm-5.1`, `zai-org/glm-5v-turbo`
- MiniMax: `MiniMaxAI/MiniMax-M2`, `minimaxai/minimax-m2.1`, `minimaxai/minimax-m2.5`, `minimaxai/minimax-m2.7`
- xAI: `xai/grok-4.3`
- KAT: `kwaipilot/kat-coder-pro-v2`
- Other: `owl`

</details>

---

## 為什麼選擇 OpenNHP

當今的網際網路是一座[黑暗森林](https://en.wikipedia.org/wiki/Dark_forest_hypothesis)。攻擊者——日益仰賴大型語言模型（LLM）透過[自主漏洞利用（AVE）](https://arxiv.org/abs/2404.08144)以機器速度進行掃描、指紋辨識與漏洞利用——將每一個可達的服務都視為攻擊目標。[Gartner 預測](https://www.gartner.com/en/newsroom/press-releases/2024-08-28-gartner-forecasts-global-information-security-spending-to-grow-15-percent-in-2025) AI 驅動的網路攻擊將快速增長。傳統防禦僅在*網路放行之後*才對使用者進行身分驗證，讓暴露的連接埠、IP 與網域成為永久的攻擊面。

> **在 AI 時代，可見性 = 脆弱性。**

OpenNHP 翻轉了這個模式：**驗證前不可見。** 所有連接埠、IP 與主機名稱皆置於預設拒絕的閘門之後。只有經加密簽署的「敲門」請求通過頻外認證與授權後，才會開放存取。攻擊者無法利用他們發現不了的東西。

### 第三代網路隱藏協定

NHP 是「先隱藏服務」這條設計路線的下一代演進：

| 世代 | 協定 | 限制 |
|---|---|---|
| 1 | 連接埠敲門（Port Knocking） | 明文傳輸，易受重放攻擊 |
| 2 | 單封包授權（SPA） | 共享金鑰、單向通訊、通常僅隱藏連接埠、通常以 C/C++ 實作 |
| **3** | **NHP** | 現代密碼學、具狀態的雙向通訊、同時隱藏網域 + IP + 連接埠、無狀態且可水平擴展、記憶體安全的 Go 實作 |

NHP 與既有的 IAM、DNS、FIDO 與零信任政策引擎並肩運作，而非取代它們——它是對既有技術棧的擴充，而非分支。

---

## 架構

OpenNHP 採模組化設計，包含三個核心元件，設計靈感來自 [NIST 零信任架構](https://www.nist.gov/publications/zero-trust-architecture)：

![OpenNHP architecture](docs/images/OpenNHP_Arch.gif)

| 核心元件 | 職責 |
|-----------|------|
| **NHP-Agent** | 用戶端，發送加密的「敲門」請求以取得存取權限 |
| **NHP-Server** | 認證與授權請求；獨立部署，在架構上與受保護主機解耦 |
| **NHP-AC** | 存取控制器，管理受保護伺服器上的防火牆規則 |

| 附加元件 | 職責 |
|-----------|------|
| **NHP-Relay** | HTTP 到 UDP 橋接，使瀏覽器代理能夠透過 HTTPS 發送 NHP 敲門請求 |
| **NHP-KGC** | 基於身分加密（IBC）的金鑰產生中心 |

### 協定流程

1. Agent 向 Server 發送加密的敲門請求（`NHP_KNK`）。
2. Server 驗證敲門請求，並向 AC 發送操作請求（`NHP_AOP`）。
3. AC 開啟防火牆，並回覆（`NHP_ART`）給 Server。
4. Server 向 Agent 回傳含存取資訊的確認（`NHP_ACK`）。
5. Agent 透過 AC 存取受保護資源。

### 密碼學

OpenNHP 提供兩套可互換的加密套件：

- **`CIPHER_SCHEME_CURVE`** —— Curve25519 + AES-256-GCM + BLAKE2s
- **`CIPHER_SCHEME_GMSM`** —— SM2 + SM4-GCM + SM3

兩者皆以 [Noise 協定框架](https://noiseprotocol.org/)為基礎。以身分為基礎的密碼學（IBC）模式透過金鑰產生中心（KGC）提供。

> 更多協定細節、部署模型與密碼學設計，請參閱[官方文件](https://docs.opennhp.org)。

---

## 儲存庫結構

```
opennhp/
├── nhp/              # 核心協定函式庫（Go 模組）
│   ├── core/         # 封包處理、密碼學、Noise 協定、裝置管理
│   ├── common/       # 共用型別與訊息定義
│   ├── utils/        # 工具函式
│   ├── plugins/      # 外掛處理介面
│   ├── log/          # 日誌基礎架構
│   └── etcd/         # 分散式設定支援
└── endpoints/        # 背景程式實作（Go 模組，依賴 nhp）
    ├── agent/        # NHP-Agent 背景程式
    ├── server/       # NHP-Server 背景程式
    ├── ac/           # NHP-AC（存取控制器）背景程式
    ├── db/           # NHP-DB（DHP 的資料經紀人）
    ├── kgc/          # NHP-KGC（金鑰產生中心）
    └── relay/        # NHP-Relay 背景程式
```

---

## 快速開始

### 先決條件

- Go 1.25.6+
- `make`
- Docker 與 Docker Compose（用於完整展示環境）

### 建置

```bash
# 建置所有元件
make

# 建置個別背景程式
make agentd    # NHP-Agent
make serverd   # NHP-Server
make acd       # NHP-AC
make db        # NHP-DB
make relayd    # NHP-Relay
make kgc       # NHP-KGC
```

### 測試

```bash
cd nhp && go test ./...
cd endpoints && go test ./...
```

### 以 Docker 執行

```bash
cd docker && docker-compose up --build
```

請依循[快速入門教學](https://docs.opennhp.org/nhp_quick_start/)，在 Docker 環境中模擬完整的認證工作流程。

---

## 貢獻

歡迎貢獻！送出 Pull Request 前請先閱讀 [CONTRIBUTING.md](CONTRIBUTING.md)。

**注意：** 所有 commit 必須以已驗證的 GPG 或 SSH 金鑰簽署。

```bash
git commit -S -m "your message"
```

---

## 安全

發現漏洞了嗎？請依循 [SECURITY.md](SECURITY.md) 所述的負責任揭露流程，請勿直接提交公開 issue。

---

## 贊助商

<a href="https://layerv.ai">
  <img src="docs/images/layerv_logo.png" height="40" alt="LayerV.ai logo">
</a>

---

## 授權條款

依 [Apache 2.0 授權條款](LICENSE)發布。

## 聯絡方式

- 電子郵件：[support@opennhp.org](mailto:support@opennhp.org)
- Discord：[加入我們的 Discord](https://discord.gg/CpyVmspx5x)
- 官方網站：[https://opennhp.org](https://opennhp.org)
