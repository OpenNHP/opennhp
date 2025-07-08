---
title: Quick Start
layout: page
nav_order: 2
permalink: /quick_start/
---

# Quick Start
{: .fs-9 }

A locally built Docker debugging environment, simulating nhp-server, nhp-ac, traefik, web-app, etc. This environment can be used for:
- Quickly understanding how opennhp works
- Plugin debugging
- Basic logic validation
- Partial performance stress testing

{: .fs-6 .fw-300 }

[中文版](/zh-cn/quick_start/){: .label .fs-4 }

---

## Overview

![Workflow](https://opennhp.org/images/infrastructure.jpg)

### Container
|Container Name|	IP	|Description|
|:---|:---|:---|
|NHP-Agent|	177.7.0.8|	Runs nhp-agentd & nginx (both disabled by default). Port mapping: 443→AC:80, 80→NHP-Server:62206|
|NHP-Server	|177.7.0.9	|Runs nhp-serverd with exposed port 62206|
|NHP-AC	|177.7.0.10|	Runs nhp-acd & traefik. All ports blocked by default|
|Web App	|177.7.0.11	|Protected web application. Only allows NHP-AC access on port 8080|

### Protection Effectiveness
|State|	Expected Result|
|:---|:---|
|Scenario 1	|Default (Protected State)	Ping or direct access to NHP-AC Server's proxied Web-app fails|
|Scenario 2	|After "knocking" via NHP-Agent	Can successfully access the NHP-AC protected Web-app|
|Scenario 3	|After web identity authentication "knock"	Can successfully access the NHP-AC protected Web-app|


## Installing Docker Environment
### Docker Desktop for Mac
```shell
brew install --cask docker
```
Alternative: Download the .dmg package directly from Docker's official website:
https://www.docker.com/products/docker-desktop/

### Docker Desktop for Windows
- System Requirements:
  - Windows 10/11 (64-bit, Pro/Enterprise/Home editions)
  - WSL 2 enabled (recommended) or Hyper-V

- Installation Steps:
  - Download Docker Desktop from the official website
  - Run the installer and follow the setup wizard

Launch Docker Desktop after installation completes

## Building from Source Code
***Note: This environment is built from local source code***
### Building the opennhp-base Docker Image
```shell
cd ./docker
docker build --no-cache -t opennhp-base:latest -f Dockerfile.base ../..
```

## Running and Testing the Environment
The following startup command will compile nhp-server, nhp-ac, web-app, and nhp-agent images during the startup process. In actual debugging, you can use ```docker compose build [container_name]``` (such as``` docker compose build nhp-ac ```to compile nhp-ac) to compile the service separately

### Start All Services
```shell
cd ./docker
docker compose up -d
```

### Scenario 1: Default (Protection Status: No NHP Agent and Web Authentication Knock)
Enter the nhp agentd container for verification
```shell
cd ./docker
docker exec -it nhp-agent bash
```
By default, the following error occurs when using curl NHP-AC (under protection)
```shell
root@68a230812459:/workdir# curl -i  http://177.7.0.10
curl: (28) Failed to connect to 177.7.0.10 port 80: Connection timed out
```

### Scenario 2: Using nhp-agentd service to knock on the door
After starting the nhp agentd service with the command ```nohup /nhp-agent/nhp-agentd run 2>&1 &```, the access is normal as follows:

```shell
root@68a230812459:/workdir# nohup /nhp-agent/nhp-agentd run 2>&1 &
root@6e21724b68f1:/workdir# curl -i http://177.7.0.10
HTTP/1.1 200 OK
Content-Length: 26
Content-Type: application/json; charset=utf-8
Date: Tue, 08 Jul 2025 06:21:10 GMT

{"message":"Hello World!"}
```

### Scenario 3: Using simulated authorization service login to verify
Stop the nhp-agentd service and start nginx in the NHP-Agent container
```shell
root@6e21724b68f1:/workdir# ps -aux|grep nhp-agentd
root        38  0.3  0.2 1974072 20448 pts/0   Sl   02:55   0:00 /nhp-agent/nhp-agentd run
root        51  0.0  0.0   2844  1424 pts/0    S+   02:55   0:00 grep --color=auto nhp-agentd
root@6e21724b68f1:/workdir# kill 38
root@6e21724b68f1:/workdir# nginx
```

visit: http://localhost/plugins/example?resid=demo&action=login

-Expected page to display normally
-Visit before knocking on the door: https://localhost/ Timeout (504 Gateway Time out)
-Click login (after knocking on the door), the page will jump to normal and can be accessed normally https://localhost/ (Note: The opening time is 15 seconds, and access is prohibited after 15 seconds)
-In the NHP Agent container, use ``` curl - i http://177.7.0.10 ```Can display content normally

### Scan the NHP-AC port in the NHP-Agent container

Enter the NHP-Agent container and install nmap
```shell
root@ee88ec992447:/# docker exec -it nhp-agent bash
root@ee88ec992447:/# apt-get update && apt-get install -y nmap
```
#### Scan NHP-AC ports
When nhp-agentd stops, no ports can be scanned
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

When nhp-agentd starts, it can scan to port 80 of NHP-AC
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

### Verify if the ipset rules are effective
```shell
docker exec -it nhp-ac ipset list
```
After knocking on the door through nhp-agentd or authorized plugins, if the following result appears in NHP-AC's ipset, it indicates that the rule was successfully written, which means that the knocking was successful:
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
