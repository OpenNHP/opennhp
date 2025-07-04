---
title: Quick Start
layout: page
nav_order: 2
permalink: /quick_start/
---

# Quick Start
{: .fs-9 }

A locally built Docker debugging environment that simulates the setup of nhp-server, nhp-ac, traefik, app, etc. This environment can be used for:
- plugins debugging
- basic logic verification
- local performance stress testing

{: .fs-6 .fw-300 }

[中文版](/zh-cn/quick_start/){: .label .fs-4 }

---

## Workflow

![Workflow](https://opennhp.org/images/infrastructure.jpg)

## Build Base Image

```shell
cd ./docker
docker build --no-cache -t opennhp-base:latest -f Dockerfile.base ../..
```

### Configure Local HTTPS

- Generate local HTTPS certificates
Enter to ./docker/certs and execute the following command:
```
openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
  -keyout server.key -out server.crt -subj "/CN=opennhp.cn" \
  -addext "subjectAltName=DNS:opennhp.cn,IP:127.0.0.1"
```

- Add /etc/hosts configuration

```
127.0.0.1       loginlocal.opennhp.org
127.0.0.1       applocal.opennhp.org
```


## Start
***Note: Enter the docker directory (cd ./docker) first***
```shell
docker compose up -d
```

## Testing

### Case 1: Use NHP agent service to knock on the door

#### Build nhp-agent image
***Note: Enter the docker directory (cd ./docker) first***
```shell
docker build --no-cache -t opennhp-agent:latest -f Dockerfile.agent ..
```

#### Create nhp-agent service container
```shell
docker run -d \
  --name nhp-agent \
  --network=host \
  -v ./nhp-agent/etc:/nhp-agent/etc \
  -v ./nhp-agent/logs:/nhp-agent/logs \
  opennhp-agent:latest
```
#### Stop/Start nhp-agent service
```shell
docker stop/start nhp-agent
```
#### Check the effect by starting/stopping the NHP agent service

- When nhp-agent starts, https://applocal.opennhp.org/ should be SUCCESSFUL.
- When nhp-agent stops, https://applocal.opennhp.org/ should be TIMEOUT.

### Case 2: Use authorization service to log in to knock on the door

https://loginlocal.opennhp.org/plugins/example?resid=demo&action=login

- The page should display normally
- After clicking login, it should redirect automatically

### Scan the nhp-ac container port in the nhp-enter container

#### Enter nhp-enter container && Install nmap tool
```shell
# docker exec -it nhp-enter bash
root@ee88ec992447:/# apt-get update && apt-get install nmap
```
#### Scan the nhp-ac container port
When nhp-agent stops，Unable to scan any ports
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
When nhp-agent starts，Can scan up to port 80
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

### Verify ipset Rules
```shell
docker exec -it nhp-ac ipset list
```
If the following results appear, it indicates successful writing, meaning the knock was successful:

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

## Stress Testing

```shell
ab -k -n 10000 -c 100 'https://loginlocal.opennhp.org/plugins/example?resid=demo&action=valid&username=user&password=password'
```