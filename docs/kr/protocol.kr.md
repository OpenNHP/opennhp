---
layout: page
title: 프로토콜 참조
nav_order: 13
has_children: true
permalink: /kr/protocol/
description: OpenNHP 온더와이어 형식 참조. CSA NHP 규격을 Go 구현에 매핑합니다.
---

# 프로토콜 참조
{: .fs-9 }

OpenNHP가 구현하는 네트워크 인프라 은닉 프로토콜(NHP)의 온더와이어 형식
참조 문서입니다. 새로운 클라이언트를 작성하거나, NHP를 다른 언어로
이식하거나, NHP 패킷을 디코딩해야 하는 도구를 개발하고 있다면 여기가
공식적인 출발점입니다.
{: .fs-6 .fw-300 }

*CSA 《Stealth Mode SDP for Zero Trust Network Infrastructure》백서 — §NHP Message Header, §NHP Message Types 및 Appendix 2 대응.*

## 이 섹션의 내용

- **[메시지 헤더]({{ '/kr/protocol/header/' | relative_url }})** — 모든 NHP 또는 DHP 패킷 앞에 붙는 240/304바이트 고정 헤더. 필드 레이아웃, 난독화 방식, 각 필드가 암호 프로토콜에서 하는 역할.
- **[메시지 타입]({{ '/kr/protocol/messages/' | relative_url }})** — 17가지 NHP 메시지 타입과 11가지 DHP 타입 전부를 발신자/수신자 역할, 페이로드 필드, 소스 코드 진입점과 함께 나열합니다.

## 규격 ↔ 구현

| CSA 규격 절 | OpenNHP 코드 |
|---|---|
| NHP Message Header (Table 3) | [`nhp/core/scheme/curve/header.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/scheme/curve/header.go), [`nhp/core/scheme/gmsm/header.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/scheme/gmsm/header.go), [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go) |
| NHP Message Types (Table 4, Appendix 2) | [`nhp/common/nhpmsg.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/nhpmsg.go) |
| 암호 알고리즘 및 프레임워크 | [`nhp/core/crypto.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/crypto.go), [`nhp/core/device.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/device.go) |
| NHP 워크플로 (Figure 3) | [`endpoints/agent/`](https://github.com/OpenNHP/opennhp/tree/main/endpoints/agent), [`endpoints/server/`](https://github.com/OpenNHP/opennhp/tree/main/endpoints/server), [`endpoints/ac/`](https://github.com/OpenNHP/opennhp/tree/main/endpoints/ac) |
| SDP 통합 | [배포]({{ '/kr/deploy/' | relative_url }}) 참고 |
| 로깅 | [`nhp/log/`](https://github.com/OpenNHP/opennhp/tree/main/nhp/log) |

## 규격 버전

OpenNHP는 CSA 제로 트러스트 워킹 그룹이 발표한 NHP 백서를 따릅니다. 해당
백서는 OpenNHP **v0.6.0**을 참조 구현으로 인용합니다. 규격과 코드가
일치하지 않는 경우 이 저장소는 코드를 기준으로 삼으며, 차이점은 워킹
그룹에 피드백합니다.

## 범위

이 섹션이 다루는 것은 온더와이어 형식입니다.

- 메시지 헤더 바이트 레이아웃
- 암호화 및 인증 봉투(Noise + AEAD)
- 공개 헤더 필드의 난독화 메커니즘
- 메시지 타입 ID와 그 페이로드 스키마
- Session-ID, 카운터(Counter), 재전송 방지 의미론

**이 섹션에서 다루지 않는 것:**

- 설정 파일 형식 → [OpenNHP 배포]({{ '/kr/deploy/' | relative_url }}) 참고
- 구성 요소 내부 구현 세부 사항 → [소스 코드 해설]({{ '/kr/code/' | relative_url }}) 참고
- 플러그인 인터페이스 → [서버 플러그인 개발]({{ '/kr/server_plugin/' | relative_url }}) 및 [클라이언트 SDK]({{ '/kr/agent_sdk/' | relative_url }}) 참고

## 용어

이 섹션에서 사용하는 표준 용어는 [용어집]({{ '/kr/glossary/' | relative_url }})에
정의되어 있습니다. 백서는 SDP 통합 맥락에서 가끔 동의어를 사용하지만
(예: "게이트웨이"로 NHP-AC를 지칭), Go 코드와 일치시키기 위해 이 문서는
항상 NHP 공식 명칭을 사용합니다.
