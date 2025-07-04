---
layout: page
title: 快速开始
parent: 中文版
nav_order: 2
permalink: /zh-cn/quick_start/
---

# 快速开始
{: .fs-9 }
一个本地搭建的 Docker 调试环境，模拟 nhp-server、nhp-ac、traefik、app 等。此环境可用于：
- 快速理解 opennhp 的运作方式
- 插件调试
- 基本逻辑验证
- 局部性能压力测试

{: .fs-6 .fw-300 }

[English](/quick_start/){: .label .fs-4 }

---

## 工作流程

![Workflow](https://opennhp.org/images/infrastructure.jpg)

## 编译 opennhp-base 镜像

```shell
cd ./docker
docker build --no-cache -t opennhp-base:latest -f Dockerfile.base ../..
```

### 配置本机的 ssl 证书

- 生成本机 ssl 证书
进入到 ./docker/certs 目录，执行以下的命令:
```
openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
  -keyout server.key -out server.crt -subj "/CN=opennhp.cn" \
  -addext "subjectAltName=DNS:opennhp.cn,IP:127.0.0.1"
```

- 添加 /etc/hosts 配置

```
127.0.0.1       loginlocal.opennhp.org
127.0.0.1       applocal.opennhp.org
```


## 开始
***注意: 先进入到 docker 目录(cd ./docker) ***
```shell
docker compose up -d
```

## 测试
### 场景一: 使用 nhp-agent 服务来敲门

#### 编译 nhp-agent 镜像
***注意: 先进入到 docker 目录(cd ./docker) ***
```shell
docker build --no-cache -t opennhp-agent:latest -f Dockerfile.agent ..
```

#### 新建 nhp-agent 服务
```shell
docker run -d \
  --name nhp-agent \
  -v ./nhp-agent/etc:/nhp-agent/etc \
  -v ./nhp-agent/logs:/nhp-agent/logs \
  opennhp-agent:latest
```
#### 停止/启动 nhp-agent 服务
```shell
docker stop/start nhp-agent
```
#### 停止/启动服务来验证效果

- 当 nhp-agent 启动时, https://applocal.opennhp.org/ 是允许访问的.
- 当 nhp-agent 停止时, https://applocal.opennhp.org/ 是不允许访问的.

### 场景二: 使用模拟授权服务的登录来验证
https://loginlocal.opennhp.org/plugins/example?resid=demo&action=login

- 预期页面正常显示
- 点击登录后，能正常跳转，并能访问正常

### 在 nhp-enter 容器中，扫描 nhp-ac 端口

#### 进入 nhp-enter 容器 并安装 nmap
```shell
# docker exec -it nhp-enter bash
root@ee88ec992447:/# apt-get update && apt-get install nmap
```
#### 扫描 nhp-ac 端口
当 nhp-agent 停止，扫描不到任何端口
```shell
# nmap 177.9.0.10
Starting Nmap 7.93 ( https://nmap.org ) at 2025-07-03 07:33 UTC
Nmap scan report for nhp-ac.docker_nginx (177.9.0.10)
Host is up (0.000044s latency).
All 1000 scanned ports on nhp-ac.docker_nginx (177.9.0.10) are in ignored states.
Not shown: 1000 filtered tcp ports (no-response)
MAC Address: 12:B4:5C:EB:72:F4 (Unknown)

Nmap done: 1 IP address (1 host up) scanned in 21.84 seconds
```
当 nhp-agent 启动，可以扫描到 80 端口
```shell
root@ee88ec992447:/# nmap 177.9.0.10
Starting Nmap 7.93 ( https://nmap.org ) at 2025-07-03 07:37 UTC
Nmap scan report for nhp-ac.docker_nginx (177.9.0.10)
Host is up (0.000094s latency).
Not shown: 999 filtered tcp ports (no-response)
PORT   STATE SERVICE
80/tcp open  http
MAC Address: 12:B4:5C:EB:72:F4 (Unknown)

Nmap done: 1 IP address (1 host up) scanned in 4.96 seconds
```

### 验证 ipset 规则是否生效
```shell
docker exec -it nhp-ac ipset list
```
如果出现以下结果，则表示规则写入成功，这意味着敲门成功：

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
177.9.0.13,udp:80,177.9.0.10 timeout 14 packets 0 bytes 0
177.9.0.13,tcp:80,177.9.0.10 timeout 14 packets 138 bytes 28068
192.168.65.1,tcp:80,177.9.0.10 timeout 14 packets 0 bytes 0
192.168.65.1,udp:80,177.9.0.10 timeout 14 packets 0 bytes 0

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

## 压力测试

```shell
ab -k -n 10000 -c 100 'https://loginlocal.opennhp.org/plugins/example?resid=demo&action=valid&username=user&password=password'
```