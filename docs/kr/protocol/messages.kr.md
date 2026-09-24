---
layout: page
title: 메시지 타입
parent: 프로토콜 참조
nav_order: 2
permalink: /kr/protocol/messages/
description: NHP와 DHP의 모든 메시지 타입에 대한 참조 — 방향, 페이로드 필드.
---

# NHP 메시지 타입
{: .fs-9 }

각 패킷은 메시지 헤더에 2바이트 메시지 타입 ID를 담고 있으며, 수신 측은 이를
근거로 해당 처리 로직으로 분배합니다. 현재 총 28개의 ID가 정의되어 있으며,
이 중 17개는 **NHP**(ID 0–16)에 속하며 노크/인증/접근 흐름, AC 생명주기,
릴레이, 등록, OTP 및 명시적 종료를 다룹니다. 나머지 11개는 **DHP**(ID
17–27)에 속하며, 데이터 객체 은닉 프로토콜에서 사용됩니다.
{: .fs-6 .fw-300 }

*CSA "Stealth Mode SDP" 백서 §NHP Message Types (Table 4) 및 Appendix 2에 대응. ID와 문자열 이름의 정의는 [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go)를, 페이로드 구조체는 [`nhp/common/nhpmsg.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/nhpmsg.go)를 참고하십시오.*

## 메시지 캡슐화

모든 메시지 타입은 동일한 메시지 헤더를 공유하며(참고: [메시지 헤더]({{ '/kr/protocol/header/' | relative_url }})),
그 뒤에 암호화된 페이로드가 이어집니다. 페이로드 형식은 JSON이며, 선택적으로
zlib 압축(`NHP_FLAG_COMPRESS` 플래그로 제어)을 거친 뒤, 메시지 헤더 파싱
과정에서 파생된 세션 키를 사용해 AEAD 암호화됩니다 — 단, 타입 설명에서 별도로
규정하는 경우는 예외입니다(예: `NHP-KPL`은 메시지 본문이 없음; `NHP-RLY`는
원본 내부 패킷 하나를 그대로 담음).

## 색인

NHP — 네트워크 인프라 은닉 프로토콜(Network-infrastructure Hiding Protocol)

| ID | 이름 | 방향 | 용도 |
|---:|---|---|---|
| 0 | [NHP-KPL](#nhp-kpl--keepalive) | Agent ↔ Server, AC ↔ Server | 킵얼라이브 패킷. 메시지 본문 없음. |
| 1 | [NHP-KNK](#nhp-knk--knock) | Agent → Server | 노크: 보호 대상 리소스에 대한 접근을 요청. |
| 2 | [NHP-ACK](#nhp-ack--acknowledge) | Server → Agent | 노크 응답. 성공 시 리소스 주소와 세션 파라미터를 포함. |
| 3 | [NHP-AOP](#nhp-aop--ac-operation) | Server → AC | 특정 agent → 리소스 흐름에 대해 접근을 개방(또는 거부)하도록 AC에 지시. |
| 4 | [NHP-ART](#nhp-art--ac-result) | AC → Server | `NHP-AOP`에 대한 AC의 응답. |
| 5 | [NHP-LST](#nhp-lst--list) | Agent → Server | 현재 agent가 접근 권한을 가진 서비스 목록 조회를 요청. |
| 6 | [NHP-LRT](#nhp-lrt--list-result) | Server → Agent | `NHP-LST`에 대한 응답. |
| 7 | [NHP-COK](#nhp-cok--cookie) | Server → Agent | 서버 과부하 시 하달되는 속도 제한 / 반(反) DDoS 쿠키. |
| 8 | [NHP-RKN](#nhp-rkn--re-knock) | Agent → Server | 재노크. HMAC 계산에 `NHP-COK`의 쿠키를 포함. |
| 9 | [NHP-RLY](#nhp-rly--relay) | Relay → Server | NHP-Relay가 전달하는 원본 NHP 패킷으로, 소스 주소 정보를 보존. |
| 10 | [NHP-AOL](#nhp-aol--ac-online) | AC → Server | AC가 컨트롤 플레인에 온라인 상태와 보호 중인 리소스를 통지. |
| 11 | [NHP-AAK](#nhp-aak--ac-acknowledge) | Server → AC | 서버가 AC 등록을 확인하고 AC의 공인 IP/포트를 반환. |
| 12 | [NHP-OTP](#nhp-otp--one-time-password) | Agent → Server | 등록에 사용할 일회용 비밀번호를 대역 외(out-of-band)로 하달해줄 것을 요청. |
| 13 | [NHP-REG](#nhp-reg--register) | Agent → Server | Agent가 OTP를 신원 증명 수단으로 사용하여 정적 공개키를 등록. |
| 14 | [NHP-RAK](#nhp-rak--register-acknowledge) | Server → Agent | `NHP-REG`에 대한 성공 또는 오류 응답. |
| 15 | [NHP-ACC](#nhp-acc--access) | Agent → AC | Agent가 AC의 임시 리스닝 포트에 접근 토큰을 제시. |
| 16 | [NHP-EXT](#nhp-ext--exit) | Agent → Server | 활성 세션을 조기에 종료할 것을 요청. 메시지 본문 없음. |

DHP — 데이터 객체 은닉 프로토콜(Data-object Hiding Protocol) *(색인 표의 완전성을 위해 여기 나열함. 상세한 DHP 의미는 별도의 DHP 참조 페이지에서 다룰 예정)*

{: .note }
DHP 항목은 현재 Go 상수 이름(예: `NHP_DRG`)으로만 표시되며, 아직 타입별
하위 절이 마련되어 있지 않습니다. 이후 전용 DHP 참조 페이지가 추가되어
위 표와 동일한 형식의 페이로드 필드 표를 각 행에 제공할 예정입니다.
**온라인 이름 차이:** [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go)에서
DHP 문자열 라벨은 일관되지 않습니다 — ID 17–26은 `NHP_DRG` … `NHP_DBA`로
출력되며(밑줄, NHP 접두사를 사용하여, 위 NHP 그룹이 하이픈을 사용하는 것과
다름), ID 27은 `"DHP-KNK"`(하이픈, DHP 접두사; Go 상수명은 `DHP_KNK`)로
출력됩니다. `HeaderTypeToString`을 사용하는 구현은 아래 표에 제시된
문자열과 정확히 일치해야 합니다.

| ID | 상수 | 방향 | 용도 |
|---:|---|---|---|
| 17 | `NHP_DRG` | DB → Server | 데이터 객체 하나를 NHP-Server에 등록. |
| 18 | `NHP_DAK` | Server → DB | `NHP_DRG`에 대한 확인 / 오류 코드. |
| 19 | `NHP_DAR` | DHP-Agent → Server | 등록된 데이터 객체에 대한 접근을 신청. |
| 20 | `NHP_DAG` | Server → DHP-Agent | `NHP_DAR`에 대한 승인 결정. |
| 21 | `NHP_DSA` | Server → DHP-Agent | 자체 증명(TEE / evidence) 제출을 요구. |
| 22 | `NHP_DAV` | DHP-Agent → Server | 서버에 원격 증명을 제출. |
| 23 | `NHP_DWR` | Server → DB | DB에 데이터 객체 비밀키의 래핑(wrap)을 요청. |
| 24 | `NHP_DWA` | DB → Server | DB가 래핑된 키(Key Access Object)를 반환. |
| 25 | `NHP_DOL` | DB → Server | DB 온라인 / 상태 통지. |
| 26 | `NHP_DBA` | Server → DB | `NHP_DOL`에 대한 확인. |
| 27 | `DHP_KNK` | DHP-Agent → Server | 원격 증명 증거를 포함한 DHP 방식의 노크 패킷. |

---

## NHP-KPL — Keepalive {#nhp-kpl--keepalive}

**ID:** `0` · **방향:** 양방향(Agent ↔ Server, AC ↔ Server) · **메시지 본문:** 없음(헤더만)

Agent/AC와 Server 사이의 NAT 바인딩 / TCP 연결을 유지하는 데 사용됩니다.
수신 측은 이 패킷을 수신하는 것 외에 별도의 처리를 하지 않습니다. Relay
노드는 킵얼라이브 패킷을 전달하지 **않습니다**.

## NHP-KNK — Knock {#nhp-knk--knock}

**ID:** `1` · **방향:** Agent → Server · **페이로드 구조체:** `common.AgentKnockMsg`

| JSON 키 | 필드 | 설명 |
|---|---|---|
| `headerType` | 메시지 타입 | 온라인 메시지 타입을 그대로 반환. |
| `usrId` | 사용자 ID | 요청을 시작한 자연인 또는 서비스 주체를 식별. |
| `devId` | 디바이스 ID | 다중 디바이스 / 단말 정책 검증을 지원. |
| `orgId` | 조직 ID | 선택 사항. 테넌트 범위. |
| `aspId` | ASP ID | 어떤 인가 서비스 제공자(ASP)가 인증을 수행할지 지정. |
| `resId` | 리소스 ID | 요청 대상인 보호 대상 리소스(도메인, 서비스명 또는 사용자 정의 문자열). |
| `results` | 검사 결과 | 선택 사항. `{checkID: result}` 매핑으로, 클라이언트 단말 환경 검사 결과를 반영. |
| `usrData` | 사용자 데이터 | 선택 사항. ASP 플러그인에 그대로 전달되는 자유 필드. |

노크 패킷은 콜드 스타트 시점부터 시작되는 NHP 상호작용의 **유일한** 트리거 메시지입니다.

## NHP-ACK — Acknowledge {#nhp-ack--acknowledge}

**ID:** `2` · **방향:** Server → Agent · **페이로드 구조체:** `common.ServerKnockAckMsg`

| JSON 키 | 필드 | 설명 |
|---|---|---|
| `errCode` / `errMsg` | 오류 코드 / 메시지 | `"0"`(또는 빈 값)은 성공을 의미하고, 그 외 값은 실패 사유를 나타냄. |
| `resHost` | 리소스 호스트 | 리소스 이름 → 호스트 주소 매핑. |
| `opnTime` | 개방 시간 | 이번 접근 인가의 유효 시간(초). |
| `aspToken` | ASP 토큰 | AC 측 백엔드 2차 검증에 사용(선택 사항). |
| `agentAddr` | Agent 주소 | 서버가 관측한 Agent의 공인 네트워크 튜플. |
| `acTokens` | AC 토큰 | 리소스 이름 → 단기 토큰. `NHP-ACC`와 함께 사용. |
| `preActions` | 사전 접근 동작 | 선택 사항. 리소스별로 하나의 `PreAccessInfo`(AC IP, 포트, 공개키, 토큰, 암호 스위트)를 포함하며, `NHP-ACC` 단계를 필요로 하는 배포에서 사용. |
| `redirectUrl` | 리다이렉트 URL | 선택 사항. 예: 직접 연결하지 않고 HTTPS 로그인 흐름으로 리다이렉트. |

## NHP-AOP — AC Operation {#nhp-aop--ac-operation}

**ID:** `3` · **방향:** Server → AC · **페이로드 구조체:** `common.ServerACOpsMsg`

| JSON 키 | 필드 | 설명 |
|---|---|---|
| `usrId` / `devId` / `orgId` | Agent 신원 | 소속 필드. |
| `aspId` / `resId` | ASP + 리소스 | 원본 노크와 연관됨. |
| `srcAddrs` | 소스 주소 목록 | AC가 허용해야 할 `NetAddress` 배열. |
| `dstAddrs` | 목적지 주소 목록 | 보호 대상 리소스의 튜플. |
| `opnTime` | 개방 시간 | 초 단위. `0`은 **거부** / 종료를 의미. |

NAT 참고: CGNAT 또는 출구를 공유하는 시나리오에서는 소스 IP가 Agent마다
고유하지 않습니다. 배포자는 AC와 보호 대상 리소스 사이에 애플리케이션
계층 토큰/쿠키를 한 겹 더 추가하거나(`NHP-ACC` 경로가 바로 이 방식입니다),
가능한 한 엔드투엔드 IPv6를 사용해야 합니다.

## NHP-ART — AC Result {#nhp-art--ac-result}

**ID:** `4` · **방향:** AC → Server · **페이로드 구조체:** `common.ACOpsResultMsg`

`errCode` / `errMsg`, 실제 개방 시간 `opnTime`(0은 거부를 의미), AC가
배포한 `token`, 그리고 선택적인 `preAct`(`PreAccessInfo`)를 포함합니다.
서버는 ART를 수신한 후에만 Agent에게 NHP-ACK를 회신합니다.

## NHP-LST — List {#nhp-lst--list}

**ID:** `5` · **방향:** Agent → Server · **페이로드 구조체:** `common.AgentListMsg`

`usrId`, `devId`, 선택적인 `orgId`, `aspId` 및 자유 형식의 `usrData`를 포함합니다.

## NHP-LRT — List Result {#nhp-lrt--list-result}

**ID:** `6` · **방향:** Server → Agent · **페이로드 구조체:** `common.ServerListResultMsg`

`errCode` / `errMsg`와 `list` 매핑을 포함하며, 구체적인 스키마는 ASP
플러그인이 결정합니다.

## NHP-COK — Cookie {#nhp-cok--cookie}

**ID:** `7` · **방향:** Server → Agent · **페이로드 구조체:** `common.ServerCookieMsg`

서버가 그대로 반환하는 `trxId`와 서버가 생성한 `cookie`를 포함합니다.
서버 과부하 시 하달됩니다. Agent는 이후 NHP-RKN을 전송해야 하며, HMAC
계산에 이 쿠키를 포함시켜 완전한 왕복 과정을 증명하고 조기에 폐기되는
것을 방지해야 합니다.

## NHP-RKN — Re-Knock {#nhp-rkn--re-knock}

**ID:** `8` · **방향:** Agent → Server · **페이로드:** NHP-KNK와 동일.

NHP-KNK와의 차이점: 일반적인 체인 키 외에도, HMAC에 NHP-COK의 쿠키가
추가로 혼합됩니다([`nhp/core/initiator.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/initiator.go)의
`addHMAC(sumCookie: true)` 참고).

## NHP-RLY — Relay {#nhp-rly--relay}

**ID:** `9` · **방향:** Relay → Server · **페이로드:** 소스 측에서 전송된 원본 내부 NHP 패킷.

Relay를 경유하는 소스 주소를 보존합니다. 다른 메시지 타입은 이 처리가
필요하지 않으며 투명하게 전달됩니다.

## NHP-AOL — AC Online {#nhp-aol--ac-online}

**ID:** `10` · **방향:** AC → Server · **페이로드 구조체:** `common.ACOnlineMsg`

`aspId`, 보호 중인 `resIds` 목록, 그리고 선택적인 `acId`를 포함합니다.
컨트롤 플레인에 AC 온라인 상태를 통지하는 데 사용됩니다.

## NHP-AAK — AC Acknowledge {#nhp-aak--ac-acknowledge}

**ID:** `11` · **방향:** Server → AC · **페이로드 구조체:** `common.ServerACAckMsg`

`errCode` / `errMsg` 및 `acAddr`을 포함합니다. AC 등록을 확인하고 AC의
공인 주소를 반환합니다(NAT 뒤에 있어 서버로부터 자신의 외부 튜플을
알아야 하는 AC에 특히 유용합니다).

## NHP-OTP — One-Time Password {#nhp-otp--one-time-password}

**ID:** `12` · **방향:** Agent → Server · **페이로드 구조체:** `common.AgentOTPMsg`

`usrId`, `devId`, 선택적인 `orgId`, `aspId`, 사전 공유된 `pass`, 그리고
자유 형식의 `usrData`를 포함합니다. ASP 플러그인이 자체 대역 외 채널을
통해 OTP(문자메시지, 이메일, QR코드)를 하달하도록 트리거합니다. 서버는
별도의 전용 타입으로 회신하지 않으며, ASP가 사이드 채널로 직접 OTP를
전달하는 것 자체가 성공으로 간주됩니다.

## NHP-REG — Register {#nhp-reg--register}

**ID:** `13` · **방향:** Agent → Server · **페이로드 구조체:** `common.AgentRegisterMsg`

`usrId`, `devId`, 선택적인 `orgId`, `aspId`, 대역 외로 획득한 `otp`, 그리고
자유 형식의 `usrData`를 포함합니다. Agent의 정적 공개키(암호화된 메시지
헤더에 담겨 있음)를 해당 신원과 결합하여 등록합니다.

## NHP-RAK — Register Acknowledge {#nhp-rak--register-acknowledge}

**ID:** `14` · **방향:** Server → Agent · **페이로드 구조체:** `common.ServerRegisterAckMsg`

`errCode`(`"0"`은 성공), `errMsg`, `aspId`를 포함합니다. 등록 과정 중의
실패는 0이 아닌 `errCode`로 명시적으로 보고됩니다. 자세한 내용은
[`HandleRegisterRequest`](https://github.com/OpenNHP/opennhp/blob/main/endpoints/server/msghandler.go)를 참고하십시오.

## NHP-ACC — Access {#nhp-acc--access}

**ID:** `15` · **방향:** Agent → AC · **페이로드 구조체:** `common.AgentAccessMsg`

`usrId`, `devId`, 선택적인 `orgId`, NHP-ACK에서 받은 `acToken`(`acTokens`에서
추출), 그리고 `usrData`를 포함합니다. 장기적인 허용 목록 대신 세션별 임시
엔드포인트(`PreAccessInfo`)를 사용하는 배포에서는, Agent가 이 메시지를
AC의 단기 리스너에 직접 제시합니다. AC는 토큰을 검증한 후 관측된 소스
주소에 대해 방화벽 허용 규칙을 설치하고 즉시 임시 포트를 닫습니다 —
[`tcpTempAccessHandler` / `udpTempAccessHandler`](https://github.com/OpenNHP/opennhp/blob/main/endpoints/ac/msghandler.go)
참고. 이 경로는 어떤 NHP 계층 응답도 생성하지 **않으며**, 성공은 이후
데이터 평면 연결이 보호 대상 리소스에 정상적으로 도달하는 것으로 암묵적으로
드러납니다. `common.ACAccessAckMsg` 구조체는 정의되어 있지만 현재 전송
경로에서는 사용되지 않으며, 예약된 필드로 간주됩니다.

## NHP-EXT — Exit {#nhp-ext--exit}

**ID:** `16` · **방향:** Agent → Server · **메시지 본문:** 없음.

Agent가 활성 세션을 조기에 종료할 것을 능동적으로 요청합니다. 서버는
즉시 AC에 `opnTime = 0`인 NHP-AOP를 전송하여 해당 허용 규칙을 닫습니다.

---

## DHP 메시지 타입 {#dhp}

DHP는 NHP의 온라인 포맷을 재사용하여 독립된 일련의 흐름 — 데이터 객체 등록,
접근, 원격 증명, 키 래핑 등 — 을 처리합니다. 여기서는 ID 표의 완전성을
위해 나열만 하였으며, 각 타입의 페이로드 필드(`DRGMsg`, `DAKMsg`,
`DARMsg`, `DAGMsg`, `DSAMsg`, `DAVMsg`, `DWRMsg`, `DWAMsg`, `DBOnlineMsg`,
`ServerDBAckMsg`, `DHPKnockMsg`)는
[`nhp/common/nhpmsg.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/nhpmsg.go)에
정의되어 있습니다. 전용 DHP 참조 페이지에서 이를 상세히 다룰 예정이며,
그때까지는 [DHP 빠른 시작]({{ '/kr/dhp_quick_start/' | relative_url }})을 참고하십시오.

---

## 함께 보기

- [메시지 헤더]({{ '/kr/protocol/header/' | relative_url }}) — 모든 타입이 공유하는 봉투(envelope)
- [암호화 알고리즘]({{ '/kr/cryptography/' | relative_url }}) — 페이로드가 암호화되는 방식
- [용어집]({{ '/kr/glossary/' | relative_url }}) — 이 페이지에서 등장하는 각 역할의 표준 명칭
