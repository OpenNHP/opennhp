---
layout: page
title: DHP快速开始
nav_order: 3
permalink: /zh-cn/dhp_quick_start/
---

# DHP快速开始
{: .fs-9 }

一个本地搭建的 Docker 调试环境，模拟 nhp-server、nhp-db和nhp-agent。此环境可用于：
{: .fs-6 .fw-300 }

- 快速理解 opendhp 的运作方式
- 基本逻辑验证
{: .fs-6 .fw-300 }


---

## 1. 概述

OpenDHP的主要目的是加强数据主权，在保持数据可用性的同时确保数据不可见性，并在整个数据生命周期中维护数据隐私。

本快速入门指南帮助开发人员快速搭建OpenDHP Docker环境、构建源代码并测试OpenDHP的关键功能。该环境设计得轻量且易于使用，非常适合希望快速测试和调试OpenDHP的开发人员。

### 1.1 架构
![Architecture](/images/OpenDHP_Arch_CN.png)

#### 1.1.1 网络拓扑

| 容器名              | IP            | 说明                                                                                                       |
| ------------------  | ------------  | --------------------------------------------------------------------------------------------------------- |
| NHP-Agent           | 177.7.0.8     | nhp-agentd， 端口映射: 443→Host: 8443                                                                       |
| NHP-Server          | 177.7.0.9     | nhp-serverd， 开放端口 62206                                                                               |
| NHP-DB              | 177.7.0.12    | nhp-db， 用来发布数据                                                                                       |

### 1.2 测试场景
#### 1.2.1 场景描述
为提升涉险账户识别的全面性与准确性，银行在内部风控识别基础上，还可通过与其他银行、支付机构、公安或监管平台提供的涉险账户信息联合验证。为保障数据安全与用户隐私，各参与方通过机密计算技术协同判断某账户是否存在风险行为，避免直接明文暴露用户数据。

#### 1.2.2 场景架构
![Scenario Architecture](/images/OpenDHP_Scenario_CN.png)

## 2. 安装Docker环境
关于这部分，请参考[NHP快速开始](/zh-cn/quick_start/)的相应章节。

## 3. 运行和配置环境

以下启动命令，在启动过程会相应的构建 nhp-server、nhp-db和nhp-agent镜像。

### 3.1 启动所有服务

```shell
cd ./docker
docker compose -f docker-compose.dhp.yaml up -d
```

### 3.2 以DHP模式启动nhp-agent
由于 `nhp-agent` 默认不会自动启动，因此需要手动启动它。

```shell
docker exec -it nhp-agent /bin/bash
/nhp-agent/nhp-agentd dhp
```

### 3.3 启动nhp-db
由于 `nhp-db` 默认不会自动启动，因此需要手动启动它。

```shell
docker exec -it nhp-db /bin/bash
/nhp-db/nhp-db run
```

### 3.4 配置DHP相关的参数
由于代理和 TEE 密钥在首次启动时生成，因此需要在nhp-server中配置代理公钥以建立信任，在nhp-db中配置TEE公钥以建立信任。此外，还需要在nhp-server中配置受信任的执行环境，用于评估远程证明报告。

#### 3.4.1 配置agent公钥到nhp-server
DHP Web 控制台的每个请求都需要携带 Bearer Token。nhp-agent 首次启动时会在 `etc/console.token`（权限 0600）中生成该 Token。默认 Docker 环境中 agent 以 root 运行，宿主机上的非 root 用户无法直接读取挂载出来的文件，因此请通过容器读取。执行下面的命令前先导出：
```shell
export NHP_CONSOLE_TOKEN=$(docker exec nhp-agent cat /nhp-agent/etc/console.token)
```
控制台默认只监听回环地址（`etc/console.toml` 中的 `ListenAddr`）；Docker 环境在容器内监听所有地址，并且只发布到主机回环地址（`127.0.0.1:8443`）。在 Docker 网络内部，控制台监听 nhp-agent 容器的所有网卡，因此 `network_default` 上的其他容器也能访问到它，但同样需要 Token。8443 端口只发布到宿主机回环地址，同一端口上的 DHP demo 入口也受此影响：请在 Docker 宿主机上用浏览器打开 https://localhost:8443。如果需要从其他机器访问，请在明确知情的情况下修改 `docker-compose.dhp.yaml` 中的端口映射（例如改回 `"8443:443"`），控制台仍然受 Token 保护。

在此默认的 Docker 环境中，8443端口已为nhp-agent映射到主机，因此可以通过https://localhost:8443从主机访问nhp-agent的 HTTP 接口。你可以使用以下curl命令获取代理的公钥。

```shell
curl --insecure -H "Authorization: Bearer $NHP_CONSOLE_TOKEN" https://localhost:8443/api/v1/key/agent
{"publicKey":"f+HWVbhQ6ZR3e+INU7ZSGyn3XNls5TUdbZWlPmj/1v890WLDW7RcnnbJmqqufymK+Yb99dadX+PlhK4qFYxtOg=="}
```
接下来，配置agent公钥在nhp-server中。
```shell
docker exec -it nhp-server /bin/bash
vi /nhp-server/etc/agent.toml
# list the agent peers for the server under [[Agents]] table

# PubKeyBase64: public key for the agent in base64 format.
# ExpireTime (epoch timestamp in seconds): peer key validation will fail when it expires.
[[Agents]]
PubKeyBase64 = "f+HWVbhQ6ZR3e+INU7ZSGyn3XNls5TUdbZWlPmj/1v890WLDW7RcnnbJmqqufymK+Yb99dadX+PlhK4qFYxtOg=="
ExpireTime = 1924991999
```
配置完代理公钥后，你可以使用以下curl命令检查代理状态：
```shell
curl --insecure -H "Authorization: Bearer $NHP_CONSOLE_TOKEN" https://localhost:8443/api/v1/status/agent
{"attestationVerified":false,"running":true,"trustedByNHPDB":false,"trustedByNHPServer":true}
```
你将看到`trustedByNHPServer`为`true`，这表示该代理已被 NHP 服务器信任。
#### 3.4.2 配置TEE证明到nhp-server
你可以使用以下curl命令获取 TEE 远程证明报告。

**注意：**该远程证明报告是在非 TEE 环境下根据容器信息生成，非TEE环境仅用于测试目的。

```shell
curl --insecure -H "Authorization: Bearer $NHP_CONSOLE_TOKEN" https://localhost:8443/api/v1/attestation/tee
{"measure":"3460bc69b9d273ad15c91074d8fd41abc5d5ccac50730d2e0495d08558848e34","serial number":"3460bc69b9d273ad15c91074d8fd41abc5d5ccac50730d2e0495d08558848e34"}
```
接下来，你需要在nhp-server中配置远程证明信息。
```shell
docker exec -it nhp-server /bin/bash
vi /nhp-server/etc/config.toml
```

如果运行本演示栈的主机没有 Hygon CSV 卡，那么上面的远程证明报告并不是一份合法的 CSV 报告，nhp-server 上的 csv 校验器会拒绝它（你将看到 `attestationVerified:false`，以及服务端的一条 Error 日志）。要让 nhp-server 接受容器生成的证据，请在 `config.toml` 中加入下面这一行：

```toml
AttestationScheme = "test"
```

默认值（key 缺失，或任何不等于 `"test"` 的值）为 `"csv"`，这要求一份真实的 Hygon CSV 远程证明报告。`"test"` 接受自声明的证据，**没有任何加密保证** —— 服务端在启动时以及每次重载且解析后的 scheme 仍为 `"test"` 时，都会输出一条 `Critical: AttestationScheme="test": ...` 日志。请把这条日志视作 **"你正在缺少一道安全护栏的情况下运行"**。切勿在信任边界之外可达的主机上设置 `AttestationScheme = "test"`。

然后再配置 TEE 白名单：
```shell
vi /nhp-server/etc/tee.toml
# list trusted execution environments under [[TEEs]] table

# Measure: cryptographic hashes that ensure the integrity of software and data within the TEE.
# SerialNumber: unique serial number of the TEE.

[[TEEs]]
Measure = "19178a674248bbca705863bbf75ecaa049fcf3dfcc5ff59a80dcc5cbb60dae59"
SerialNumber = "TMEX300023050201"

[[TEEs]]
Measure = "3460bc69b9d273ad15c91074d8fd41abc5d5ccac50730d2e0495d08558848e34"
SerialNumber = "3460bc69b9d273ad15c91074d8fd41abc5d5ccac50730d2e0495d08558848e34"
```
配置代理公钥后，你可以使用以下curl命令检查代理状态：
```shell
curl --insecure -H "Authorization: Bearer $NHP_CONSOLE_TOKEN" https://localhost:8443/api/v1/status/agent
{"attestationVerified":true,"running":true,"trustedByNHPDB":false,"trustedByNHPServer":true}
```
你将看到`attestationVerified`为`true`，这表示TEE已被NHP服务器信任。
#### 3.4.3 配置TEE公钥到nhp-db
你可以使用以下curl命令获取TEE公钥。
```shell
curl --insecure -H "Authorization: Bearer $NHP_CONSOLE_TOKEN" https://localhost:8443/api/v1/key/tee
{"publicKey":"pup5OzTTZjddv+WBgbUBkvHuBgJoBg0DU+I2c7Qj4lHlrVM8N/Yl9F6DEnbGFBWB89xrN6VLhYAIM4Xv+mu4KA=="}
```
接下来，你需要在nhp-db中配置TEE公钥。
```shell
docker exec -it nhp-db /bin/bash
vi /nhp-db/etc/tee.toml
# Configuration for trusted execution environment.

# TEEPublicKeyBase64: base64 encoded public key of TEE (Trusted Execution Environment).
[[TEEs]]
TEEPublicKeyBase64 = "pup5OzTTZjddv+WBgbUBkvHuBgJoBg0DU+I2c7Qj4lHlrVM8N/Yl9F6DEnbGFBWB89xrN6VLhYAIM4Xv+mu4KA=="
ExpireTime = 1924991999
```
## 4. 测试银行涉险账户场景
### 4.1 发布数据资源
你可以使用以下命令发布数据资源：
```shell
docker exec -it nhp-db /bin/bash
cd /nhp-db
./nhp-db run --mode encrypt --data-source-type online --source ./demo/risk.involved.accounts.csv --output ./risk.involved.accounts.csv.demo.ztdo --smart-policy ./demo/smart.policy.json --metadata ./demo/metadata.json
```
如果出现信息`Successfully register or update data object which doId is <doId>.`，表示数据资源已成功发布。

### 4.2 请求数据资源
#### 4.2.1 编写和编译可信应用
为了实现与可信应用的简便且统一的通信，我们采用了模型上下文协议（Model Context Protocol，简称 MCP）。这意味着可信应用被实现为MCP服务器，而内置于NHP Agent中的客户端则作为MCP客户端与可信应用进行通信。由于MCP框架几乎支持所有编程语言，因此使用任何语言实现可信应用都十分简单和直接。

以下是一个用 Golang 编写的简单可信应用示例，该示例在本演示中使用。
```go
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
    s := server.NewMCPServer("trusted application", "1.0.0",
        server.WithToolCapabilities(true),
    )

    s.AddTool(
        mcp.NewTool("verify_account",
            mcp.WithDescription("Verify account to check whether there are any risk factors associated with the account"),
            mcp.WithString("path",
                mcp.Required(),
                mcp.Description("path to file which records the account details"),
            ),
            mcp.WithString("account_id",
                mcp.Description("account id"),
				mcp.Required(),
            ),
        ),
        verifyAccountHandler,
    )

    // Start STDIO server
    if err := server.ServeStdio(s); err != nil {
        fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
        os.Exit(1)
    }
}

func verifyAccountHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    path, err := req.RequireString("path")
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

    accountId, err := req.RequireString("account_id")
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

	is_risk, err := findRecord(path, accountId)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

    return mcp.NewToolResultText(fmt.Sprintf(`{"account_id":"%s","is_risk": %t}`, accountId, is_risk)), nil
}

func findRecord(path string, accountId string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ','

	for {
		record, err := r.Read()
		if err != nil {
			if err == csv.ErrFieldCount {
				continue
			}

			if err == io.EOF {
				return false, nil
			}

			return false, err
		}

		if record[0] == accountId {
			return true, nil
		}
	}
}
```
假设我们已将其编译为名为 `ta` 的二进制文件。

#### 4.2.2 注册可信应用
只有经过运维密钥签名的可信应用才会被接受。先生成一次 ed25519 签名密钥，把公钥写入 `docker/nhp-agent/etc/console.toml` 的 `TrustedTAPublicKeys` 并重启 nhp-agentd 进程：在 3.2 节运行它的终端里按 Ctrl+C 停止，再重新执行 `/nhp-agent/nhp-agentd dhp`。只执行 `docker restart nhp-agent` 不够，因为容器不会自动启动 nhp-agentd（重启容器后需要再执行一遍 3.2 节）；控制台的 `/api/v1/agent/restart` 也不会重新加载 `console.toml`。然后对 TA 签名：
```shell
openssl genpkey -algorithm ed25519 -out ta_signing.pem
openssl pkey -in ta_signing.pem -pubout -outform DER | tail -c 32 | base64   # -> TrustedTAPublicKeys
{ printf 'opennhp-dhp-ta-v1\0'; openssl dgst -sha256 -binary ta; } > ta.msg   # context string + NUL + SHA-256 of the TA
export TA_SIGNATURE=$(openssl pkeyutl -sign -inkey ta_signing.pem -rawin -in ta.msg | base64)
```
`ta_signing.pem` 不要放在 agent 主机上。

签名只覆盖 TA 文件本身；`name`、`description`、`language`、`entry` 是未签名的描述信息（agent 执行的是已存储、已校验的文件）。如需不再接受某个已签名的 TA，把它的 SHA-256（`openssl dgst -sha256 ta`）加入 `console.toml` 的 `RevokedTADigests` 并重启 nhp-agentd 进程。吊销只拦截注册；已经在运行的 TA 也要靠这次进程重启才会被卸载。

你可以使用以下命令来注册可信应用：
```shell
curl --insecure -H "Authorization: Bearer $NHP_CONSOLE_TOKEN" --request POST --url https://localhost:8443/api/v1/ta/register --header 'content-type: multipart/form-data' --form file=@ta --form 'name=ta' --form 'language=go' --form 'entry=ta' --form 'description=trusted application demo' --form "signature=$TA_SIGNATURE"
[
  {
    "method": "POST",
    "name": "/api/v1/ta/ceca4572-644b-4bde-a4b6-ac6048f8fba6/verify_account",
    "description": "Verify account to check whether there are any risk factors associated with the account",
    "params": [
      {
        "name": "doId",
        "description": "identifier of the data object",
        "type": "string"
      },
      {
        "name": "account_id",
        "description": "account id",
        "type": "string"
      }
    ]
  }
]
```
注册成功后，你可以访问由可信应用暴露的HTTP RESTful API。

#### 4.2.3 执行任务
你可以使用以下curl命令调用这些暴露的RESTful API来执行隐私保护计算：
```shell
curl --insecure -H "Authorization: Bearer $NHP_CONSOLE_TOKEN" --request POST --url https://localhost:8443/api/v1/ta/ceca4572-644b-4bde-a4b6-ac6048f8fba6/verify_account --header 'content-type: application/json' --data '{"doId": "47d2b67c-ef80-45fc-814d-effd23baf788", "account_id": "62230121012345678901"}'
```
执行后，你将收到响应 `{"account_id":"62230121012345678901","is_risk":false}`，表示该账户不是涉险账户。

由于NHP Agent、可信应用和数据均受可信执行环境（TEE）保护，数据资源消费者无法直接访问数据资源。所有对数据的操作必须通过TEE内部受控且经过验证的执行流程进行，确保数据始终保持机密性，并且按照关联的智能数据策略进行处理。整个DHP设计保证了消费者能够使用数据，却永远无法查看或提取数据的原始内容。
