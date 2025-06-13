[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![OpenNHP Logo](docs/images/logo11.png)
# OpenNHP: ゼロトラストネットワークインフラストラクチャ隠蔽プロトコル
攻撃者からサーバーとデータを隠すためのOSI第5層に位置する、軽量の暗号化駆動型ゼロトラストネットワークプロトコルです。

![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Version](https://img.shields.io/badge/version-1.0.0-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)

---

## セキュリティの利点

OpenNHPは*OSIセッション層*でゼロトラストの原則を実装しているため、次のような大きな利点があります。

- インフラの隠蔽による攻撃面の削減
- 不正なネットワーク偵察の防止
- 脆弱性の悪用を防ぐ
- 暗号化されたDNSによるフィッシング防止
- DDoS攻撃に対する防御
- 細粒度のアクセス制御を実現
- アイデンティティベースの接続追跡
- 攻撃の帰属

## アーキテクチャ

OpenNHPのアーキテクチャは[NISTゼロトラストアーキテクチャ標準](https://www.nist.gov/publications/zero-trust-architecture)に触発されています。以下の図に示すように、3つの主要なコンポーネント（**NHP-Server**、**NHP-AC**、**NHP-Agent**）を持つモジュール設計に従います。

![OpenNHP architecture](docs/images/OpenNHP_Arch.png)

> アーキテクチャとワークフローの詳細については、[OpenNHPドキュメント](https://opennhp.org/)を参照してください。

## コア: 暗号化アルゴリズム

暗号化はOpenNHPの中心にあり、強力なセキュリティ、高いパフォーマンス、およびスケーラビリティを提供するために最新の暗号化アルゴリズムを利用しています。以下は、OpenNHPで使用されている主要な暗号化アルゴリズムとフレームワークです。

- **[楕円曲線暗号（ECC）](https://en.wikipedia.org/wiki/Elliptic-curve_cryptography)**：効率的な公開鍵暗号に使用されています。

> RSAと比較して、ECCは短い鍵長で強力な暗号化を提供し、ネットワーク伝送と計算パフォーマンスを向上させます。以下の表は、RSAとECCのセキュリティ強度、鍵長、および鍵長の比率の違いを示し、それぞれの有効期間を示しています。

| セキュリティ強度（ビット） | DSA/RSA鍵長（ビット） | ECC鍵長（ビット） | 比率：ECC対DSA/RSA | 有効期限 |
|:------------------------:|:-------------------------:|:---------------------:|:----------------------:|:--------:|
| 80                       | 1024                      | 160-223               | 1:6                    | 2010年まで |
| 112                      | 2048                      | 224-255               | 1:9                    | 2030年まで |
| 128                      | 3072                      | 256-383               | 1:12                   | 2031年以降 |
| 192                      | 7680                      | 384-511               | 1:20                   | |
| 256                      | 15360                     | 512+                  | 1:30                   | |

- **[ノイズプロトコルフレームワーク](https://noiseprotocol.org/)**：安全な鍵交換、メッセージの暗号化/復号化、および相互認証を可能にします。

> ノイズプロトコルは[ディフィー・ヘルマン鍵共有](https://en.wikipedia.org/wiki/Diffie%E2%80%93Hellman_key_exchange)に基づいており、相互およびオプションの認証、アイデンティティの隠蔽、前方秘匿性、ゼロラウンドトリップ暗号化などの最新の暗号化ソリューションを提供します。そのセキュリティとパフォーマンスは、[WhatsApp](https://www.whatsapp.com/security/WhatsApp-Security-Whitepaper.pdf)、[Slack](https://github.com/slackhq/nebula)、および[WireGuard](https://www.wireguard.com/)などの人気アプリケーションで既に証明されています。

- **[アイデンティティベース暗号（IBC）](https://en.wikipedia.org/wiki/Identity-based_cryptography)**：大規模な鍵配布を簡素化します。

> 効率的な鍵配布は、ゼロトラストの実装に不可欠です。OpenNHPはPKIとIBCの両方をサポートしています。PKIは数十年にわたって広く使用されてきましたが、アイデンティティの確認と鍵管理に中央集権的な認証局（CA）に依存しており、時間とコストがかかることがあります。一方、IBCは、アイデンティティの確認と鍵管理を分散型で自己管理可能な方法で行うことができ、リアルタイムで何十億ものデバイスやサーバーを保護し、オンボーディングする必要があるOpenNHPのゼロトラスト環境において、よりコスト効率的です。

- **[証明書レス公開鍵暗号（CL-PKC）](https://en.wikipedia.org/wiki/Certificateless_cryptography)**：推奨されるIBCアルゴリズム

> CL-PKCは、鍵エスクローを回避し、アイデンティティベース暗号（IBC）の制限に対処することでセキュリティを強化するスキームです。ほとんどのIBCシステムでは、ユーザーの秘密鍵は鍵生成センター（KGC）によって生成され、これは重大なリスクをもたらします。KGCが侵害された場合、すべてのユーザーの秘密鍵が公開される可能性があり、KGCへの完全な信頼が必要です。CL-PKCは鍵生成プロセスを分割し、KGCは部分的な秘密鍵のみを知っているため、CL-PKCはPKIとIBCの両方の強みを組み合わせ、中央集権的な鍵管理の欠点なしに強力なセキュリティを提供します。

詳細について：

> OpenNHPで使用されている暗号化アルゴリズムの詳細な説明については、[OpenNHPドキュメント](https://opennhp.org/cryptography/)を参照してください。

## 主な機能

- デフォルトで「すべて拒否」ルールを適用することにより、脆弱性の悪用を軽減
- 暗号化されたDNS解決を通じてフィッシング攻撃を防止
- インフラの隠蔽によるDDoS攻撃の防御
- アイデンティティベースの接続による攻撃の帰属
- 保護されたリソースに対するすべてのアクセスをデフォルトで拒否
- ネットワークアクセス前にアイデンティティおよびデバイスベースの認証
- DNSハイジャックを防止するための暗号化されたDNS解決
- DDoS攻撃を緩和するための分散インフラ
- 分離されたコンポーネントによるスケーラブルなアーキテクチャ
- 既存のアイデンティティおよびアクセス管理システムとの統合
- さまざまな展開モデルをサポート（クライアント対ゲートウェイ、クライアント対サーバーなど）
- 最新のアルゴリズム（ECC、ノイズプロトコル、IBC）を使用した暗号化によるセキュリティの確保

<details>
<summary>機能の詳細を表示</summary>

- **デフォルト拒否のアクセス制御**：すべてのリソースはデフォルトで隠蔽され、認証と認可が行われた後にのみアクセス可能になります。
- **アイデンティティおよびデバイスベースの認証**：既知のユーザーと承認されたデバイスのみがアクセス可能です。
- **暗号化されたDNS解決**：DNSハイジャックとそれに伴うフィッシング攻撃を防止します。
- **DDoS緩和**：分散型インフラ設計により、分散型サービス拒否攻撃を防御します。
- **スケーラブルなアーキテクチャ**：分離されたコンポーネントにより柔軟な展開とスケーリングが可能です。
- **IAM統合**：既存のアイデンティティおよびアクセス管理システムと連携します。
- **柔軟な展開**：クライアント対ゲートウェイ、クライアント対サーバーなど、さまざまなモデルをサポートします。
- **強力な暗号化**：ECC、ノイズプロトコル、IBCなどの最新アルゴリズムを使用して強力なセキュリティを提供します。
</details>

## 展開

OpenNHPは、さまざまなユースケースに合わせた複数の展開モデルをサポートしています。

- クライアント対ゲートウェイ：ゲートウェイの背後にある複数のサーバーへのアクセスを保護します
- クライアント対サーバー：個々のサーバー/アプリケーションを直接保護します
- サーバー対サーバー：バックエンドサービス間の通信を保護します
- ゲートウェイ対ゲートウェイ：サイト間接続を保護します

> 詳細な展開手順については、[OpenNHPドキュメント](https://opennhp.org/deploy/)を参照してください。

## SPAとNHPの比較
[クラウドセキュリティアライアンス（CSA）](https://cloudsecurityalliance.org/)がリリースした[ソフトウェア定義境界（SDP）仕様](https://cloudsecurityalliance.org/artifacts/software-defined-perimeter-zero-trust-specification-v2)には、シングルパケット認証（SPA）プロトコルが含まれています。NHPは、最新の暗号化フレームワークとアーキテクチャを通じてセキュリティ、信頼性、スケーラビリティ、拡張性を向上させ、[AHAC研究論文](https://www.mdpi.com/2076-3417/14/13/5593)で示されているように従来の技術の限界を克服しています。

| - | SPA |NHP | NHPの利点  |
|:---|:---|:---|:---|
| **アーキテクチャ** | SPAサーバーのパケット復号化およびユーザー/デバイス認証コンポーネントがネットワークアクセス制御コンポーネントと結合されています。 | NHP-Server（パケット復号化およびユーザー/デバイス認証コンポーネント）とNHP-AC（アクセス制御コンポーネント）が分離されています。NHP-Serverは別のホストに展開でき、水平スケーリングをサポートします。 | <ul><li>パフォーマンス：リソース消費の多いコンポーネントNHP-Serverが保護されたサーバーから分離されています。</li><li>スケーラビリティ：NHP-Serverは分散またはクラスター化モードで展開可能です。</li><li>セキュリティ：認証が成功するまでは、保護されたサーバーのIPアドレスがクライアントには見えません。</li></ul>|
| **通信** | 単方向 | 双方向 | アクセス制御のステータス通知による信頼性の向上 |
| **暗号化フレームワーク** | 共有シークレット | PKIまたはIBC、ノイズフレームワーク |<ul><li>セキュリティ：MITM脅威を軽減する証明された安全な鍵交換メカニズム</li><li>低コスト：ゼロトラストモデルにおける効率的な鍵配布</li><li>パフォーマンス：高パフォーマンスの暗号化/復号化</li></ul>|
| **ネットワークインフラストラクチャ隠蔽能力** | サーバーポートのみ | ドメイン、IP、ポート | 脆弱性、DNSハイジャック、DDoS攻撃など、さまざまな攻撃に対する強力な防御 |
| **拡張性** | なし、SDP専用 | 汎用 | あらゆるサービス暗黒化の必要があるシナリオに対応 |
| **相互運用性** | 利用不可 | カスタマイズ可能| NHPは既存のプロトコル（例：DNS、FIDOなど）とシームレスに統合可能 |

## コントリビューション

OpenNHPへの貢献を歓迎します！貢献方法の詳細については、[コントリビューションガイドライン](CONTRIBUTING.md)を参照してください。

## ライセンス

OpenNHPは[Apache 2.0ライセンス](LICENSE)の下でリリースされています。

## 連絡先

- プロジェクトウェブサイト：[https://github.com/OpenNHP/opennhp](https://github.com/OpenNHP/opennhp)
- メール：[opennhp@gmail.com](mailto:opennhp@gmail.com)
- Slackチャンネル：[Slackに参加する](https://slack.opennhp.org)

詳細なドキュメントについては、[公式ドキュメント](https://opennhp.org)をご覧ください。

## 参考文献

- [ソフトウェア定義境界（SDP）仕様 v2.0](https://cloudsecurityalliance.org/artifacts/software-defined-perimeter-zero-trust-specification-v2)。Jason Garbis、Juanita Koilpillai、Junaid Islam、Bob Flores、Daniel Bailey、Benfeng Chen、Eitan Bremler、Michael Roza、Ahmed Refaey Hussein。[*クラウドセキュリティアライアンス（CSA）*](https://cloudsecurityalliance.org/)。2022年3月。
- [AHAC：高度なネットワーク隠蔽アクセス制御フレームワーク](https://www.mdpi.com/2076-3417/14/13/5593)。Mudi Xu、Benfeng Chen、Zhizhong Tan、Shan Chen、Lei Wang、Yan Liu、Tai Io San、Sou Wang Fong、Wenyong Wang、Jing Feng。*応用科学ジャーナル*。2024年6月。
- [STALE ：電子メールと ECDH 鍵交換を活用したスケーラブルでセキュアなクロスボーダー認証スキーム](https://www.mdpi.com/2079-9292/14/12/2399) Zhizhong Tan， Mudi Xu， Benfeng Chen， Anyu Wang， Shuo Zhang， Yan Liu， Jiexin Zheng， Kevin Qi Zhang， and Wenyong Wang.*電子ジャーナル*。2025 年 6 月。
- ノイズプロトコルフレームワーク。https://noiseprotocol.org/
- 脆弱性管理フレームワークプロジェクト。https://phoenix.security/web-vuln-management/

---

✨ OpenNHPにご関心をお寄せいただき、ありがとうございます！皆様の貢献とフィードバックをお待ちしております。

