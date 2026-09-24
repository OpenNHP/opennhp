---
layout: page
title: 용어집
nav_order: 14
permalink: /kr/glossary/
description: OpenNHP와 CSA NHP 규격이 공유하는 표준 용어.
---

# 용어집
{: .fs-9 }

OpenNHP와 CSA NHP 규격이 공유하는 표준 용어입니다. 이슈, PR, 생성되는 코드와
도구에서는 여기 표기를 우선 사용하세요. 백서는 SDP 통합 맥락에서 가끔 다른
이름을 쓰기도 하는데, 그런 경우 이 표에서는 "다른 이름" 칸에 함께 적어
두지만, 구현에서는 항상 표준 용어를 사용합니다.
{: .fs-6 .fw-300 }

*CSA 《Stealth Mode SDP》백서 §NHP Core Components 및 §Components That Interact with NHP 대응.*

## 프로토콜 역할

| 표준 용어 | 다른 이름 | 의미 |
|---|---|---|
| **NHP-Agent** | agent, 클라이언트, 요청자 | 클라이언트 구성 요소로, 호스트·SDK·브라우저·애플리케이션에 내장될 수 있습니다. 암호화된 노크 요청을 보내고, 인가가 통과되면 보호 대상 리소스에 연결합니다. |
| **NHP-Server** | server, 컨트롤러(SDP 맥락) | 제어 평면 구성 요소: 노크 패킷을 복호화하고, NHP-Agent를 인증하며, ASP를 통해 정책을 검증하고, NHP-AC에 지시합니다. NIST SP 800-207의 *Policy Engine/Administrator*에 대응합니다. |
| **NHP-AC** | 접근 제어기, 게이트웨이(SDP 맥락) | 데이터 평면 구성 요소: 기본 거부를 실행하고, NHP-Server의 지시에 따라 방화벽을 동적으로 개방합니다. NIST SP 800-207의 *Policy Enforcement Point*에 대응합니다. |
| **NHP-DB** | — | 데이터 콘텐츠 은닉 프로토콜(DHP)의 데이터 객체 백엔드로, 암호화된 제로 트러스트 데이터 객체를 저장합니다. |
| **NHP-KGC** | KGC | 신원 기반 암호(IBC)를 위한 키 생성 센터로, agent/server에 키를 발급합니다. |
| **NHP-Relay** | relay, 릴레이 | 네트워크 세그먼트를 넘나들며 중계하는 선택적 NHP 패킷 포워더입니다. `NHP-RLY` 메시지 타입을 통해 원본 주소를 보존합니다. |

## 관련 주체

| 용어 | 의미 |
|---|---|
| **리소스 요청자(Resource Requestor)** | NHP-Agent를 탑재하고 보호 대상 리소스에 대한 접근을 요청하는 주체(사용자, 디바이스, 애플리케이션, 또는 서버). |
| **인가 서비스 제공자(ASP, Authorization Service Provider)** | NHP-Server가 인증 시 조회하는 외부 IAM/정책 시스템(예: 기업 IAM, 또는 NHP가 SDP 배포 앞단에 위치할 경우 SDP Controller일 수도 있음). |
| **보호 대상 리소스(Protected Resource)** | NHP에 의해 숨겨지는 서비스, 호스트, 포트 또는 데이터 객체. |
| **로그 서버(Log Server)** | NHP-Server/NHP-AC의 접근 및 감사 로그를 받는 선택적 외부 SIEM 또는 로그 저장소. |

## 프로토콜 개념

| 용어 | 의미 |
|---|---|
| **노크(Knock)** | NHP-Agent가 NHP-Server로 보내는 최초의 암호화된 패킷(메시지 타입 `NHP-KNK`)으로, agent 신원, 디바이스 정보, 대상 리소스 식별자를 담습니다. |
| **연결 전 인증(Authenticate-before-connect)** | 보호 대상 리소스로의 TCP/UDP 연결이 수립되기 전에 신원 검증을 완료하는 것. NHP의 핵심 기여입니다. |
| **기본 거부(Default-deny)** | NHP-AC의 기본 자세: NHP-Server가 명시적으로 허용을 지시하지 않는 한 인바운드 트래픽을 항상 거부합니다. |
| **세션(Session)** | 인증되고 유효 기간이 있는 접근 창(window). `Session ID`로 `NHP-KNK`/`NHP-ACK`/`NHP-AOP`/`NHP-ART` 사이를 연관 짓습니다. |
| **개방(Open-door)** | `NHP-AOP`에 의해 트리거되어, NHP-AC에서 특정 "출발지→목적지" 트래픽을 동적으로 허용하는 동작. |
| **임시 키(Ephemeral key)** | 메시지마다 새로 생성되는 일회성 ECC 키 쌍으로, Noise 핸드셰이크를 구동하며 세션별 전방 비밀성(forward secrecy)을 제공합니다. |
| **정적 키(Static key)** | NHP-Agent/NHP-Server/NHP-AC의 장기 신원 키로, 대역 외 방식 또는 NHP-REG/NHP-RAK 절차를 통해 등록됩니다. |

## 암호학

| 용어 | 의미 |
|---|---|
| **ECC** | 타원 곡선 암호. NHP 국제 스위트는 기본적으로 Curve25519(256비트)를 사용하고, 확장/국가밀도(gmsm) 스위트는 SM2를 사용합니다. |
| **Noise Protocol Framework** | NHP가 양방향 인증과 세션별 키 유도에 사용하는 핸드셰이크 및 대칭 암호화 프레임워크. OpenNHP의 핸드셰이크 구조는 Noise `K` 패턴(`e`, `es`, `ss`)을 따르며, 정적 키 암호문 앞에 AEAD로 감싼 IBC 신원 블록(`MaximumIdentitySize`; PKI 모드에서는 이 블록이 전부 0)을 추가로 삽입합니다. |
| **IBC** | 신원 기반 암호(Identity-Based Cryptography). 사용자의 ID 문자열이 곧 공개키가 되어 키 배포를 단순화하지만, NHP-KGC의 키 위탁(key escrow) 문제를 야기합니다. |
| **CL-PKC** | 인증서 없는 공개키 암호(Certificateless Public Key Cryptography). IBC의 한 변형으로, 키 생성 책임을 KGC와 사용자 양쪽으로 나눠 키 위탁 문제를 완화합니다. |
| **PKI** | 전통적인 X.509 인증서 및 CA 체계. IBC와 병행하여 지원됩니다. |
| **암호 스위트(Cipher scheme)** | 암호 기본 요소들의 명명된 조합. `CIPHER_SCHEME_CURVE` = Curve25519 + AES-256-GCM + BLAKE2s(국제/표준); `CIPHER_SCHEME_GMSM` = SM2 + SM4-GCM + SM3(국가밀도). |

## 아키텍처 프레임워크

| 용어 | 의미 |
|---|---|
| **제로 트러스트 아키텍처(ZTA)** | NHP가 구현하는 보안 모델 — "절대 신뢰하지 말고, 항상 검증하라" — NIST SP 800-207에 의해 공식화되었습니다. |
| **소프트웨어 정의 경계(SDP)** | CSA의 프레임워크로, NHP는 원래 이를 보강하기 위해 설계되었습니다. SDP 용어로는 NHP-AC가 *Gateway*, NHP-Server가 *Controller*에 대응합니다. |
| **단일 패킷 인가(SPA)** | NHP가 대체하는 이전 세대 네트워크 은닉 기술. SPA는 단방향 암호화 노크를 사용하며, NHP는 여기에 양방향 인증, 명시적 피드백, 확장성을 더했습니다. |
| **기본 개방 vs 기본 거부** | TCP/IP는 기본적으로 *개방*되어 있습니다(어떤 상대든 연결을 시작할 수 있음). NHP는 인증이 성공하기 전까지 *거부* 상태를 유지합니다. |

## 메시지 타입 명명 접두사

모든 NHP 프로토콜 메시지 이름은 `NHP-`로 시작하고, 이어서 세 글자 니모닉이 붙습니다. 전체 목록은 [메시지 타입 참조]({{ '/kr/protocol/messages/' | relative_url }})를 참고하세요.
