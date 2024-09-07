---
layout: page
title: 代码解读
parent: Chinese
nav_order: 4
permalink: /zh-cn/code/
---

# OpeNHP代码解读

## 1. 层级架构

1.上层逻辑组件层负责UDP的连接建立、维护与断开
2.Device层负责：1.将上层的消息明文转为NHP报文并发送到连接；2.将从连接收到的NHP报文转化为消息明文并提供上层处理
3.上层逻辑组件提供
![avatar](./images/provide.png)

## 2. 连接管理

1.上层逻辑组件可以建立并维护多个连接UdpConn，根据实际需求创建所需对象成员。每一个UdpConn起一个线程进行收发包操作。
2.每一个UdpConn需要建立一个Device层的ConnData，并向Device ConnData传递实际连接中的远端地址，报文收发通道，cookie等。
3.每一个UdpConn允许进行多次双向的transaction或单向发包。（agent除外，原则上agent每次请求都创建一个新的连接）
4.每一个transaction都建立一个自身的线程和通道用于维持交互操作，超时后自行销毁。Local transaction（本地创建的交互）由device统一管理，Remote transaction（远端创建的交互）由远端连接管理，transaction的回应在收发包时需要找出相应的transaction线程进行后续操作。

## 3. 对象命名

1.上层逻辑组件在收发方向上可能具有多重身份，Device层中使用initiator和responder表示发起方和接收方。

## 4. 报文缓冲区的创建与销毁（回收）

1. 为了提高吞吐率，报文缓冲区不采用自动垃圾回收机制而采用waitpool分配回收机制。
2. 接收：device创建报文缓冲区接收网络数据，根据NHP包头对报文进行解析与校验。解析结果存储在ResponderSessionParams结构中（名称不好理解，可能会改变）。明文消息仍然会使用报文缓冲区。缓冲区的销毁分两种情况，单向通信的结构体在上层应用获取明文消息后销毁。transaction接收缓冲区在transaction结束后销毁。
3. 发送：device创建报文缓冲区，填充包头并对消息进行加密后存储在InitiatorSessionParams结构中并发送。transaction发送在未收到对端回应时会重试发送。缓冲区的销毁分两种情况，单向通信的结构体在发送后销毁。transaction发送的缓冲区在transaction结束后销毁。

**消息的加密与解密：**
连接中接收到的UDP原始数据会被device解析并放入device的MsgToPacketQueue队列中，等待后端处理。
发送消息到连接时，需构建initiatorsessionstarter结构传入消息信息与连接信息，放入device的MsgToPacketQueue队列中，device会将消息进行加密发出。
