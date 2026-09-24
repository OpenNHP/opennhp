---
layout: page
title: OpenNHP 소개
nav_order: 1
description: "OpenNHP: 제로 트러스트 네트워크 은닉 프로토콜"
permalink: /kr/overview/
---

# OpenNHP: 제로 트러스트 네트워크 은닉 프로토콜
{: .fs-9 }


---

## 1장: 안내

**NHP**는 중국계산기학회(CCF)가 발표한 제로 트러스트 네트워크 은닉 프로토콜로, 제로 트러스트 SDP(소프트웨어 정의 경계)의 **단일 패킷 인가 프로토콜(SPA)**과 비교했을 때 더 뛰어난 은닉성, 더 강력한 성능과 신뢰성, 더 유연한 확장성, 그리고 우수한 신뢰 가능한 국산화(信创) 호환성을 갖추고 있습니다.

**OpenNHP**는 NHP 표준을 기반으로 개발된 오픈소스 소프트웨어 프로젝트입니다. OpenNHP 프로젝트를 시작하기 전에 다음 글을 먼저 읽어보시길 권장합니다.

OpenNHP의 코드 구조와 기술적 세부 사항에 대해서는 다음을 읽어보세요.

- [《OpenNHP 소스 코드 해설》](../code/)

## 2장: OpenNHP의 호환성

OpenNHP는 우수한 호환성을 갖추고 있으며, 특히 신뢰 가능한 국산화(信创) 생태계를 잘 지원합니다. 다음은 OpenNHP가 호환하는 암호 알고리즘 및 소프트웨어/하드웨어 목록입니다.

### 2.1 암호 알고리즘

| 국가 암호 알고리즘 |  *SM2、SM3、SM4*  |  
|---|---|
| **국제 암호 알고리즘**   |  ***Curve25519、AES、SHA256***  |

### 2.2 운영체제

| 운영체제 |  호환성  |
|---|:---:|
| Windows   |  ✅ |
| Apple macOS   |  ✅ |
| UOS   |  ✅ |
| Kylin OS   |  ✅ |
| CETC Puhua OS   |  ✅ |  
| Apple iOS   |  ✅ |
| Android   |  ✅ |

### 2.3 CPU 명령어 집합

| CPU 종류 |  호환성  |
|---|:---:|
| x86   |  ✅ |
| ARM(Huawei Kunpeng)   |  ✅ |
|  LoongArch(龙芯)   |  ✅ |
|  SW(申威)   |  ✅ |
