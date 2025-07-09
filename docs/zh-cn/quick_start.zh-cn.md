---
layout: page
title: 快速开始
parent: 中文版
nav_order: 2
permalink: /zh-cn/quick_start/
---

# 快速开始
{: .fs-9 }
一个本地搭建的 Docker 调试环境，模拟 nhp-server、nhp-ac、traefik、web-app 等。此环境可用于：
- 快速理解 opennhp 的运作方式
- 插件调试
- 基本逻辑验证
- 局部性能压力测试

{: .fs-6 .fw-300 }

[English](/quick_start/){: .label .fs-4 }

---

## 1. 概述

![Workflow](https://opennhp.org/images/infrastructure.jpg)

### 1.1 容器
| 容器名 | IP | 说明 |
|-------|-------|-------|
| NHP-Agent  | 177.7.0.8    |nhp-agentd & nginx（默认均不运行），443->AC:80, 80-> NHP-Server:62206|
| NHP-Server | 177.7.0.9    |nhp-serverd，开放端口 62206|
| NHP-AC     | 177.7.0.10   |nhp-acd & traefik，禁止任何端口访问|
| Web app    | 177.7.0.11   |被保护的 Web app，只允许 NHP-AC 访问 8080 端口|

### 1.2 防护效果
| 场景一   | 默认（防护状态）   | ping 或者访问 NHP-AC Server 代理的 Web-app 失败|
|-------         |-------         |-------|
| 场景二 | 通过 NHP-Agent 敲门后  | 能正常访问通过 NHP-AC 防护的 Web-app |
| 场景三 | 通过 web 身份认证敲门后 | 能正常访问通过 NHP-AC 防护的 Web-app|
 
## 2. 安装 Docker 环境
### 2.1 Docker Desktop for Mac
```shell
brew install --cask docker
```
或者从 Docker 官网下载 .dmg 文件手动安装：
https://www.docker.com/products/docker-desktop/
### 2.2 Docker Desktop for Windows
- 系统要求：
  - Windows 10/11（64 位，专业版/企业版/家庭版）
  - 启用 WSL 2（推荐）或 Hyper-V
- 安装步骤
  - 下载 Docker Desktop：官网下载
  - 运行安装程序，按提示完成安装。
  

安装完成后，启动 Docker Desktop。

## 3. 编译代码
***注意：该环境基于本地源代码编译***
### 3.1 编译 opennhp-base 镜像
***注意: 先进入到 docker 目录(cd ./docker)***
```shell
cd ./docker
docker build --no-cache -t opennhp-base:latest -f Dockerfile.base ../..
```

## 4. 运行与测试
以下启动命令，在启动过程会相应的编译 nhp-server、nhp-ac、web-app、nhp-agent 镜像，在实际调试过程中，可使用 ``` docker compose build [container_name] ```（如：```docker compose build nhp-ac ``` 编译 nhp-ac）对服务单独编译
### 4.1 启动所有服务
***注意: 先进入到 docker 目录(cd ./docker)***
```shell
cd ./docker
docker compose up -d
```

### 4.2 场景一: 默认（防护状态：没有 nhp-agentd 与 web 身份认证 敲门）
进入 nhp-agentd 容器进行验证
***注意: 先进入到 docker 目录(cd ./docker)***
```shell
cd ./docker
docker exec -it nhp-agent bash
```
默认情况下，通过 curl NHP-AC 会出现以下错误（保护中）
```shell
root@68a230812459:/workdir# curl -i  http://177.7.0.10
curl: (28) Failed to connect to 177.7.0.10 port 80: Connection timed out
```

### 4.3 场景二: 使用 nhp-agentd 服务来敲门
通过``` nohup /nhp-agent/nhp-agentd run 2>&1 & ``` 命令来启动 nhp-agentd 服务后，访问正常，如下：

```shell
root@68a230812459:/workdir# nohup /nhp-agent/nhp-agentd run 2>&1 &
root@6e21724b68f1:/workdir# curl -i http://177.7.0.10
HTTP/1.1 200 OK
Content-Length: 26
Content-Type: application/json; charset=utf-8
Date: Tue, 08 Jul 2025 06:21:10 GMT

{"message":"Hello World!"}
```

### 4.4 场景三: 使用模拟授权服务的登录来验证
停止 nhp-agentd 服务，并启动 NHP-Agent 容器中的 nginx 
```shell
root@6e21724b68f1:/workdir# ps -aux|grep nhp-agentd
root        38  0.3  0.2 1974072 20448 pts/0   Sl   02:55   0:00 /nhp-agent/nhp-agentd run
root        51  0.0  0.0   2844  1424 pts/0    S+   02:55   0:00 grep --color=auto nhp-agentd
root@6e21724b68f1:/workdir# kill 38
root@6e21724b68f1:/workdir# nginx
```
访问：http://localhost/plugins/example?resid=demo&action=login

- 预期页面正常显示
- 敲门前访问：https://localhost/ 超时（504 Gateway Time-out）
- 点击登录（敲门后），页面正常跳转，并能正常访问 https://localhost/ （注：开门时间为 15s，15s后禁止访问）
- 在 NHP-Agent 容器内，通过 ``` curl -i  http://177.7.0.10 ``` 能正常显示内容

### 4.5 在 NHP-Agent 容器中，扫描 NHP-AC 端口
进入 NHP-Agent 容器 并安装 nmap
```shell
root@ee88ec992447:/# docker exec -it nhp-agent bash
root@ee88ec992447:/# apt-get update && apt-get install -y nmap
```
#### 4.5.1 扫描 NHP-AC 端口
当 nhp-agent 停止，扫描不到任何端口
```shell
root@ee88ec992447:/# nmap 177.7.0.10
Starting Nmap 7.93 ( https://nmap.org ) at 2025-07-03 07:33 UTC
Nmap scan report for nhp-ac.docker_nginx (177.9.0.10)
Host is up (0.000044s latency).
All 1000 scanned ports on nhp-ac.docker_nginx (177.9.0.10) are in ignored states.
Not shown: 1000 filtered tcp ports (no-response)
MAC Address: 12:B4:5C:EB:72:F4 (Unknown)

Nmap done: 1 IP address (1 host up) scanned in 21.84 seconds
```
当 nhp-agent 启动，可以扫描到 NHP-AC 的 80 端口
```shell
root@ee88ec992447:/# nmap 177.7.0.10
Starting Nmap 7.93 ( https://nmap.org ) at 2025-07-03 07:37 UTC
Nmap scan report for nhp-ac.docker_nginx (177.9.0.10)
Host is up (0.000094s latency).
Not shown: 999 filtered tcp ports (no-response)
PORT   STATE SERVICE
80/tcp open  http
MAC Address: 12:B4:5C:EB:72:F4 (Unknown)

Nmap done: 1 IP address (1 host up) scanned in 4.96 seconds
```

### 4.6 验证 ipset 规则是否生效
```shell
docker exec -it nhp-ac ipset list
```
通过 nhp-agentd 或 授权插件敲门后，如果 NHP-AC 的 ipset 中出现以下结果，则表示规则写入成功，这意味着敲门成功：
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

## 5. 修改代码重新构建镜像并调试
在实际的调试中，修改完代码后，可以使用 ```docker compose build [container_name]``` (如：``` docker compose build nhp-ac ``` 构建 nhp-ac 镜像) 来重新构建相应的服务来进行调试

### 5.1 构建 nhp-server 服务并调试
```shell
cd ./docker
docker compose build nhp-server
docker stop nhp-server && docker rm nhp-server 
docker compose up -d
```

### 5.2 构建 nhp-ac 服务并调试
```shell
cd ./docker
docker compose build nhp-ac
docker stop nhp-ac && docker rm nhp-ac 
docker compose up -d
```