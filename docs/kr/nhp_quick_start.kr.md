---
layout: page
title: NHP 빠른 시작
nav_order: 2
permalink: /kr/nhp_quick_start/
---

# NHP 빠른 시작
{: .fs-9 }

nhp-server, nhp-ac, traefik, web-app 등을 시뮬레이션하는 로컬 Docker 디버깅 환경입니다. 이 환경은 다음 용도로 사용할 수 있습니다:
{: .fs-6 .fw-300 }

- opennhp의 동작 방식을 빠르게 이해하기
- 플러그인 디버깅
- 기본 로직 검증
- 소규모 성능 부하 테스트
{: .fs-6 .fw-300 }


---

## 1. 개요

빠른 시작 가이드는 개발자가 OpenNHP Docker 환경을 신속하게 구축하고, 소스 코드를 빌드하며, OpenNHP의 핵심 기능을 테스트할 수 있도록 도와줍니다. OpenNHP가 인가되지 않은 스캔으로부터 서버를 "보이지 않게" 만드는 방식을 살펴보든, 기존 제로 트러스트 아키텍처에 통합하려 하든, 이 가이드는 빠르게 시작하고 실행하는 데 필요한 기본 단계를 제공합니다.

### 1.1 네트워크 토폴로지

![Workflow](https://docs.opennhp.org/images/infrastructure.jpg)

| 컨테이너 이름 | IP           | 설명                                                                 |
| ------------- | ------------ | -------------------------------------------------------------------- |
| NHP-Agent     | 177.7.0.8    | nhp-agentd & nginx(기본적으로 둘 다 실행되지 않음), 443->AC:80, 80-> NHP-Server:62206 |
| NHP-Server    | 177.7.0.9    | nhp-serverd, 62206 포트 개방                                          |
| NHP-AC        | 177.7.0.10   | nhp-acd & traefik, 어떠한 포트 접근도 금지                            |
| Web app       | 177.7.0.11   | 보호 대상 Web app, NHP-AC만 8080 포트 접근 허용                       |

### 1.2 테스트 시나리오

| 시나리오   | 상태                          | 기대 결과                                              |
| ---------- | ----------------------------- | ------------------------------------------------------- |
| 시나리오 1 | 은닉(인가되지 않은 사용자 대상) | ping 또는 NHP-AC Server가 프록시하는 Web-app 접근 실패 |
| 시나리오 2 | NHP-Agent 노크 후             | NHP-AC로 보호되는 Web-app에 정상적으로 접근 가능        |
| 시나리오 3 | 웹 신원 인증을 통한 노크 후    | NHP-AC로 보호되는 Web-app에 정상적으로 접근 가능        |

## 2. Docker 환경 설치

### 2.1 Docker Desktop for Mac

```shell
brew install --cask docker
```

또는 Docker 공식 웹사이트에서 .dmg 파일을 다운로드하여 수동으로 설치:
<https://www.docker.com/products/docker-desktop/>

### 2.2 Docker Desktop for Windows

- 시스템 요구사항:
  - Windows 10/11(64비트, Pro/Enterprise/Home)
  - WSL 2(권장) 또는 Hyper-V 활성화
- 설치 단계
  - Docker Desktop 다운로드: 공식 웹사이트에서 다운로드
  - 설치 프로그램을 실행하고 안내에 따라 설치를 완료합니다.

설치가 완료되면 Docker Desktop을 실행합니다.

## 3. 최신 코드 기반으로 기본 이미지 빌드

### 3.1 최신 코드 가져오기

```shell
git clone https://github.com/OpenNHP/opennhp.git
```

### 3.2 opennhp-base 이미지 빌드

***주의: 먼저 docker 디렉터리로 이동하세요(cd ./docker)***

```shell
cd ./docker
docker build --no-cache -t opennhp-base:latest -f Dockerfile.base ../..
```

## 4. 실행 및 테스트

다음 시작 명령을 실행하면 시작 과정에서 nhp-server, nhp-ac, web-app, nhp-agent 이미지가 자동으로 빌드됩니다. 실제 디버깅 과정에서는 ``` docker compose build [container_name] ```(예: ```docker compose build nhp-ac```로 nhp-ac를 빌드)을 사용하여 서비스를 개별적으로 빌드할 수 있습니다.

### 4.1 모든 서비스 시작

***주의: 먼저 docker 디렉터리로 이동하세요(cd ./docker)***

```shell
cd ./docker
docker compose up -d
```

### 4.2 시나리오 1: 은닉(인가되지 않은 사용자 대상)

nhp-agentd 컨테이너에 진입하여 검증합니다.
***주의: 먼저 docker 디렉터리로 이동하세요(cd ./docker)***

```shell
cd ./docker
docker exec -it nhp-agent bash
```

기본적으로 NHP-AC에 curl을 실행하면 다음과 같은 오류가 발생합니다(보호 중):

```shell
root@68a230812459:/workdir# curl -i  http://177.7.0.10
curl: (28) Failed to connect to 177.7.0.10 port 80: Connection timed out
```

포트 스캔 검증을 위해 NHP-Agent 컨테이너에 진입하여 nmap을 설치합니다.

```shell
root@ee88ec992447:/# docker exec -it nhp-agent bash
root@ee88ec992447:/# apt-get update && apt-get install -y nmap
```

NHP-Agent를 통해 NHP-AC를 스캔하면 어떠한 포트도 발견되지 않습니다.

```shell
root@ee88ec992447:/# nmap 177.7.0.10
Starting Nmap 7.93 ( https://nmap.org ) at 2025-07-03 07:33 UTC
Nmap scan report for nhp-ac.docker_nginx (177.7.0.10)
Host is up (0.000044s latency).
All 1000 scanned ports on nhp-ac.docker_nginx (177.7.0.10) are in ignored states.
Not shown: 1000 filtered tcp ports (no-response)
MAC Address: 12:B4:5C:EB:72:F4 (Unknown)

Nmap done: 1 IP address (1 host up) scanned in 21.84 seconds
```

### 4.3 시나리오 2: nhp-agentd 서비스를 이용한 노크

``` nohup /nhp-agent/nhp-agentd run 2>&1 & ``` 명령으로 nhp-agentd 서비스를 시작하면 다음과 같이 정상적으로 접근할 수 있습니다:

```shell
root@68a230812459:/workdir# nohup /nhp-agent/nhp-agentd run 2>&1 &
root@6e21724b68f1:/workdir# curl -i http://177.7.0.10
HTTP/1.1 200 OK
Content-Length: 26
Content-Type: application/json; charset=utf-8
Date: Tue, 08 Jul 2025 06:21:10 GMT

{"message":"Hello World!"}
```

nhp-agent가 시작되면 NHP-AC의 80 포트를 스캔할 수 있습니다.

```shell
root@ee88ec992447:/# nmap 177.7.0.10
Starting Nmap 7.93 ( https://nmap.org ) at 2025-07-03 07:37 UTC
Nmap scan report for nhp-ac.docker_nginx (177.7.0.10)
Host is up (0.000094s latency).
Not shown: 999 filtered tcp ports (no-response)
PORT   STATE SERVICE
80/tcp open  http
MAC Address: 12:B4:5C:EB:72:F4 (Unknown)

Nmap done: 1 IP address (1 host up) scanned in 4.96 seconds
```

### 4.4 시나리오 3: 시뮬레이션된 인가 서비스 로그인을 통한 검증

nhp-agentd 서비스를 중지하고, NHP-Agent 컨테이너 내의 nginx를 시작합니다.

```shell
root@6e21724b68f1:/workdir# ps -aux|grep nhp-agentd
root        38  0.3  0.2 1974072 20448 pts/0   Sl   02:55   0:00 /nhp-agent/nhp-agentd run
root        51  0.0  0.0   2844  1424 pts/0    S+   02:55   0:00 grep --color=auto nhp-agentd
root@6e21724b68f1:/workdir# kill 38
root@6e21724b68f1:/workdir# nginx
```

접속: <http://localhost/plugins/example?resid=demo&action=login>

- 페이지가 정상적으로 표시되어야 합니다.
- 노크 전 접속: <https://localhost/>는 타임아웃(504 Gateway Time-out)됩니다.
- 로그인 클릭(노크 후), 페이지가 정상적으로 이동하며 <https://localhost/>에 정상적으로 접근할 수 있습니다(참고: 개방 시간은 15초이며, 15초 후에는 접근이 금지됩니다).
- NHP-Agent 컨테이너 내에서 ```curl -i  http://177.7.0.10```를 실행하면 내용이 정상적으로 표시됩니다.
- 로그인 클릭(노크 후) 시 NHP-AC의 80 포트를 스캔할 수 있습니다.

```shell
root@ee88ec992447:/# nmap 177.7.0.10
Starting Nmap 7.93 ( https://nmap.org ) at 2025-07-03 07:37 UTC
Nmap scan report for nhp-ac.docker_nginx (177.7.0.10)
Host is up (0.000094s latency).
Not shown: 999 filtered tcp ports (no-response)
PORT   STATE SERVICE
80/tcp open  http
MAC Address: 12:B4:5C:EB:72:F4 (Unknown)

Nmap done: 1 IP address (1 host up) scanned in 4.96 seconds
```

### 4.5 ipset 규칙 적용 여부 검증

```shell
docker exec -it nhp-ac ipset list
```

nhp-agentd 또는 인가 플러그인을 통해 노크한 후, NHP-AC의 ipset에 다음과 같은 결과가 나타나면 규칙이 정상적으로 기록된 것이며, 이는 노크가 성공했음을 의미합니다:
***Name: defaultset Rules***

```shell
Name: defaultset
Type: hash:ip,port,ip
Revision: 5
Header: family inet hashsize 1024 maxelem 1000000 timeout 120 counters
Size in memory: 656
References: 7
Number of entries: 2
Members:
177.7.0.8,udp:80,177.7.0.10 timeout 8 packets 0 bytes 0
177.7.0.8,tcp:80,177.7.0.10 timeout 8 packets 90 bytes 14565

Name: defaultset_down
Type: hash:ip,port,ip
Revision: 5
Header: family inet hashsize 1024 maxelem 1000000 timeout 121 counters
Size in memory: 208
References: 2
Number of entries: 0
Members:

Name: tempset
Type: hash:net,port
Revision: 7
Header: family inet hashsize 1024 maxelem 1000000 timeout 5 counters
Size in memory: 456
References: 2
Number of entries: 0
Members:
```

## 5. 코드 수정 후 이미지 재빌드 및 디버깅

실제 디버깅 과정에서는 코드를 수정한 후 ```docker compose build [container_name]```(예: ``` docker compose build nhp-ac ```로 nhp-ac 이미지를 빌드)을 사용하여 해당 서비스를 재빌드하고 디버깅할 수 있습니다.

### 5.1 코드 수정

IDE(예: vscode)로 프로젝트를 열고 OpenNHP 소스 코드를 수정합니다.

### 5.2 서비스 재빌드 및 디버깅

수정된 서비스에 대해 다음과 같은 방식으로 재빌드 및 디버깅을 진행합니다.

### 5.2.1 nhp-server 재빌드 및 시작

```shell
cd ./docker
docker compose build nhp-server
docker stop nhp-server && docker rm nhp-server
docker compose up -d
```

### 5.2.2 nhp-ac 재빌드 및 시작

```shell
cd ./docker
docker compose build nhp-ac
docker stop nhp-ac && docker rm nhp-ac
docker compose up -d
```
