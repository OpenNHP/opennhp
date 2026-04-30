[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![zh-tw](https://img.shields.io/badge/lang-zh--tw-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-tw.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![OpenNHP Logo](docs/images/logo11.png)

# OpenNHP：オープンソースのゼロトラスト・セキュリティ・ツールキット

[![Build](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml/badge.svg)](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml)
[![Release](https://img.shields.io/github/v/tag/OpenNHP/opennhp?label=release)](https://github.com/OpenNHP/opennhp/tags)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
[![codecov](https://codecov.io/gh/OpenNHP/opennhp/branch/main/graph/badge.svg)](https://codecov.io/gh/OpenNHP/opennhp)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/OpenNHP/opennhp)

**OpenNHP** は軽量かつ暗号技術を基盤としたオープンソースのツールキットであり、インフラ・アプリケーション・データに対してゼロトラスト・セキュリティを実現します。[**クラウドセキュリティアライアンス（CSA）**](https://cloudsecurityalliance.org/) の *[Network-infrastructure Hiding Protocol（NHP）仕様](https://cloudsecurityalliance.org/artifacts/stealth-mode-sdp-for-zero-trust-network-infrastructure)* のリファレンス実装であり、次の 2 つのコアプロトコルを備えています：

- **Network-infrastructure Hiding Protocol（NHP）：** サーバーのポート、IP アドレス、ドメイン名を隠蔽し、アプリケーションやインフラを不正アクセスから保護します。
- **Data-content Hiding Protocol（DHP）：** 暗号化とコンフィデンシャル・コンピューティングによりデータのセキュリティとプライバシーを確保し、データを*「使えるが見えない」*状態にします。

**[ウェブサイト](https://opennhp.org) · [ビジョン](https://opennhp.org/vision/) · [ライブデモ](https://opennhp.org/demo/) · [ドキュメント](https://docs.opennhp.org) · [Discord](https://discord.gg/CpyVmspx5x)**

---

## なぜ OpenNHP か

現代のインターネットは[暗黒森林](https://en.wikipedia.org/wiki/Dark_forest_hypothesis)です。攻撃者は—— LLM の力を得て [Autonomous Vulnerability Exploitation](https://arxiv.org/abs/2404.08144) により機械的なスピードでスキャン、フィンガープリンティング、エクスプロイトを実行し——到達可能なすべてのサービスを標的とみなします。[Gartner は](https://www.gartner.com/en/newsroom/press-releases/2024-08-28-gartner-forecasts-global-information-security-spending-to-grow-15-percent-in-2025) AI 駆動型サイバー攻撃が急増すると予測しています。従来の防御策はネットワークにユーザーを通した *後* に認証を行うため、露出したポート、IP、ドメインは永続的な攻撃面となり続けます。

> **AI 時代において、可視性 = 脆弱性。**

OpenNHP はこのモデルを反転させます：**信頼されるまで不可視**。すべてのポート、IP、ホスト名はデフォルト拒否のゲートの背後に置かれます。アクセスが許可されるのは、暗号署名された「ノック」がアウトオブバンドで認証・認可された後に限られます。攻撃者は発見できないものを悪用できません。

### 第 3 世代のネットワーク隠蔽プロトコル

NHP は「まずサービスを隠す」という設計思想の次なる一歩です：

| 世代 | プロトコル | 制限事項 |
|---|---|---|
| 1 | ポートノッキング（Port Knocking） | 平文、リプレイ攻撃に弱い |
| 2 | Single Packet Authorization（SPA） | 共有秘密、一方向通信、通常はポートのみを隠蔽、多くが C/C++ 実装 |
| **3** | **NHP** | 現代的な暗号、ステータス付きの双方向通信、ドメイン + IP + ポートを隠蔽、ステートレスで水平スケール可能、メモリ安全な Go |

NHP は既存の IAM、DNS、FIDO、ゼロトラスト・ポリシーエンジンを置き換えるのではなく、それらと並んで動作します——スタックをフォークせず拡張します。

---

## アーキテクチャ

OpenNHP は [NIST ゼロトラスト・アーキテクチャ](https://www.nist.gov/publications/zero-trust-architecture) を参考に、3 つのコアコンポーネントから成るモジュラー設計を採用しています：

![OpenNHP architecture](docs/images/OpenNHP_Arch.gif)

| コアコンポーネント | 役割 |
|-----------|------|
| **NHP-Agent** | 暗号化された「ノック」リクエストを送信し、アクセスを得るクライアント |
| **NHP-Server** | リクエストを認証・認可。独立して稼働し、保護対象ホストとアーキテクチャ上分離されている |
| **NHP-AC** | 保護対象サーバーのファイアウォール・ルールを管理するアクセス・コントローラ |

| アドオンコンポーネント | 役割 |
|-----------|------|
| **NHP-Relay** | ブラウザベースのエージェントが HTTPS 経由で NHP ノックを送信できるようにする HTTP-UDP ブリッジ |
| **NHP-KGC** | Identity-Based Cryptography（IBC）用の鍵生成センター |

### プロトコル・フロー

1. Agent が暗号化されたノック（`NHP_KNK`）を Server に送信する。
2. Server がノックを検証し、操作リクエスト（`NHP_AOP`）を AC に送る。
3. AC がファイアウォールを開き、Server に応答（`NHP_ART`）する。
4. Server が Agent にアクセス情報を含む確認応答（`NHP_ACK`）を返す。
5. Agent は AC を通して保護対象リソースに到達する。

### 暗号方式

OpenNHP は 2 つの互換可能な暗号スイートを提供します：

- **`CIPHER_SCHEME_CURVE`** —— Curve25519 + AES-256-GCM + BLAKE2s
- **`CIPHER_SCHEME_GMSM`** —— SM2 + SM4-GCM + SM3

いずれも [Noise Protocol Framework](https://noiseprotocol.org/) に基づきます。Identity-Based Cryptography（IBC）モードは Key Generation Center（KGC）経由で利用できます。

> プロトコルの詳細、デプロイメント・モデル、暗号設計については [ドキュメント](https://docs.opennhp.org) をご覧ください。

---

## リポジトリ構成

```
opennhp/
├── nhp/              # コアプロトコル・ライブラリ（Go モジュール）
│   ├── core/         # パケット処理、暗号、Noise プロトコル、デバイス管理
│   ├── common/       # 共有型とメッセージ定義
│   ├── utils/        # ユーティリティ関数
│   ├── plugins/      # プラグイン・ハンドラ・インタフェース
│   ├── log/          # ロギング基盤
│   └── etcd/         # 分散設定サポート
└── endpoints/        # デーモン実装（Go モジュール、nhp に依存）
    ├── agent/        # NHP-Agent デーモン
    ├── server/       # NHP-Server デーモン
    ├── ac/           # NHP-AC（アクセス・コントローラ）デーモン
    ├── db/           # NHP-DB（DHP のデータ・ブローカー）
    ├── kgc/          # NHP-KGC（Key Generation Center）
    └── relay/        # NHP-Relay デーモン
```

---

## クイックスタート

### 前提条件

- Go 1.25.6+
- `make`
- Docker と Docker Compose（フルスタック・デモ用）

### ビルド

```bash
# すべてのコンポーネントをビルド
make

# 個別のデーモンをビルド
make agentd    # NHP-Agent
make serverd   # NHP-Server
make acd       # NHP-AC
make db        # NHP-DB
make relayd    # NHP-Relay
make kgc       # NHP-KGC
```

### テスト

```bash
cd nhp && go test ./...
cd endpoints && go test ./...
```

### Docker で実行

```bash
cd docker && docker-compose up --build
```

[クイックスタート・チュートリアル](https://docs.opennhp.org/nhp_quick_start/)に従い、Docker 環境で完全な認証ワークフローをシミュレートしてください。

---

## コントリビューション

コントリビューションを歓迎します！ Pull Request を送る前に [CONTRIBUTING.md](CONTRIBUTING.md) をご一読ください。

**注意：** すべてのコミットは検証済みの GPG または SSH キーで署名されている必要があります。

```bash
git commit -S -m "your message"
```

---

## セキュリティ

脆弱性を発見した場合は、公開の issue を開くのではなく、[SECURITY.md](SECURITY.md) に記載された責任ある情報開示プロセスに従ってください。

---

## スポンサー

<a href="https://layerv.ai">
  <img src="docs/images/layerv_logo.png" height="40" alt="LayerV.ai logo">
</a>

---

## ライセンス

[Apache 2.0 ライセンス](LICENSE) の下で公開されています。

## お問い合わせ

- メール：[support@opennhp.org](mailto:support@opennhp.org)
- Discord：[Discord に参加](https://discord.gg/CpyVmspx5x)
- ウェブサイト：[https://opennhp.org](https://opennhp.org)
