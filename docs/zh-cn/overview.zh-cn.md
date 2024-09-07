---
layout: page
title: OpenNHP简介
parent: 中文版
nav_order: 1
description: "OpenNHP: 零信任网络隐身协议"
permalink: /zh-cn/overview/
---

# OpenNHP：零信任网络隐身协议

## 第一章：导读

**NHP**是由中国计算机学会CCF发布的零信任网络隐身协议，对比零信任SDP（软件定义边界）中的**单包授权协议SPA**，具有更好的隐身性、更强的性能和可靠性、更灵活的扩展性、以及良好的信创兼容性。

**OpenNHP**是基于NHP标准的开发的开源软件项目。在上手OpenNHP项目前，建议阅读以下文章：

关于OpenNHP的代码结构与技术详解，请阅读：

- [《OpenNHP代码解读文档》](./code.zh-cn.md)

## 第二章：OpenNHP的兼容性

OpenNHP具备良好的兼容性，尤其是对信创生态的支持，以下是OpenNHP兼容的密码算法和软硬件。

### 2.1 密码算法

| 国产密码算法 |  *SM2、SM3、SM4*  |  
|---|---|
| **国际密码算法**   |  ***Curve25519、AES、SHA256***  |

### 2.2 操作系统

| 操作系统 |  兼容性  |
|---|:---:|
| Windows   |  ✅ |
| 苹果MacOS   |  ✅ |
| 统信UOS   |  ✅ |
| 麒麟KylinOS   |  ✅ |
| 中电科普华OS   |  ✅ |  
| 苹果iOS   |  ✅ |
| 安卓Android   |  ✅ |

### 2.3 CPU指令集

| CPU类型 |  兼容性  |
|---|:---:|
| x86   |  ✅ |
| ARM华为鲲鹏   |  ✅ |
|  LoongArch龙芯   |  ✅ |
|  SW申威   |  ✅ |

