---
layout: page
title: 源代码解读
parent: 中文版
nav_order: 6
permalink: /zh-cn/code/
---

# OpeNHP代码解读
{: .fs-9 }

---

## 1. 层级架构

1. 上层逻辑组件层负责UDP的连接建立、维护与断开
2. Device层负责：1.将上层的消息明文转为NHP报文并发送到连接；2.将从连接收到的NHP报文转化为消息明文并提供上层处理
3. 上层逻辑组件提供
![avatar](/images/provide.png)

## 2. 连接管理

1. 上层逻辑组件可以建立并维护多个连接UdpConn，根据实际需求创建所需对象成员。每一个UdpConn起一个线程进行收发包操作。
2. 每一个UdpConn需要建立一个Device层的ConnData，并向Device ConnData传递实际连接中的远端地址，报文收发通道，cookie等。
3. 每一个UdpConn允许进行多次双向的transaction或单向发包。（agent除外，原则上agent每次请求都创建一个新的连接）
4. 每一个transaction都建立一个自身的线程和通道用于维持交互操作，超时后自行销毁。Local transaction（本地创建的交互）由device统一管理，Remote transaction（远端创建的交互）由远端连接管理，transaction的回应在收发包时需要找出相应的transaction线程进行后续操作。

## 3. 对象命名

1. 上层逻辑组件在收发方向上可能具有多重身份，Device层中使用initiator和responder表示发起方和接收方。

## 4. 报文缓冲区的创建与销毁（回收）

1. 为了提高吞吐率，报文缓冲区不采用自动垃圾回收机制而采用waitpool分配回收机制。
2. 接收：device创建报文缓冲区接收网络数据，根据NHP包头对报文进行解析与校验。解析结果存储在ResponderSessionParams结构中（名称不好理解，可能会改变）。明文消息仍然会使用报文缓冲区。缓冲区的销毁分两种情况，单向通信的结构体在上层应用获取明文消息后销毁。transaction接收缓冲区在transaction结束后销毁。
3. 发送：device创建报文缓冲区，填充包头并对消息进行加密后存储在InitiatorSessionParams结构中并发送。transaction发送在未收到对端回应时会重试发送。缓冲区的销毁分两种情况，单向通信的结构体在发送后销毁。transaction发送的缓冲区在transaction结束后销毁。

**消息的加密与解密：**
连接中接收到的UDP原始数据会被device解析并放入device的MsgToPacketQueue队列中，等待后端处理。
发送消息到连接时，需构建initiatorsessionstarter结构传入消息信息与连接信息，放入device的MsgToPacketQueue队列中，device会将消息进行加密发出。

## 5. NHP-Device架构设计

1. Device负责NHP报文与消息的转换。Device初始化时需要指定类型和私钥。Device视自身类型只对相应的包进行处理。

2. 用于承载发送和接收报文的buffer比较大，所以由Device的内存Pool统一发放并回收（如果依赖于Go后台垃圾回收，高并发时会造成大量内存开销）。所以在开发时一定要注意buffer的分配**Device.AllocatePoolPacket\(\)** 和回收**Device.ReleasePoolPacket\(\)**。

   - 报文buffer回收点位于
     - 发送报文被发送后（本地transaction除外）
     - 接收报文解析完毕时（远程transaction除外）
     - 本地或远程transaction线程停止时

3. 上层逻辑调用接口**SendMsgToPacket**将消息转换成加密报文并发送到连接。

4. 上层逻辑调用接口**RecvPacketToMsg**将加密报文解析成消息后放入**DecryptedMsgQueue**队列并等待处理（通常情况）。

   - 特殊情况：如果请求发起方已指定接收通道，解析后的消息会被送到请求方指定的消息通道**ResponseMsgCh**，而不放进常规消息队列进行排队。

5. 交互（**transaction**）：一次请求需要等待一次回复的操作称为交互。一次由Device发起的交互请求为本地交互（**LocalTransaction**），一次由Device接收到的交互请求为远程交互（**RemoteTransaction**）。由于回应报文需要继承请求报文生成的**ChainKey**，所以所有的交互分发由Device进行管理。

6. 连接上下文（**ConnectionData**）：由上层逻辑传入的与连接相关的所有信息，Device在加密消息后将报文发送到连接。一个连接可以进行多个**transaction**。

7. 在建立发送请求时，需要创建**MsgAssembler**结构体。

   - Agent和AC必须填写消息类型**HeaderType**、对端**RemoteAddr**、对端公钥**PeerPk**和消息明文**Message**（如无特殊情况都采用消息压缩）。将填写好的**MsgAssembler**发给各自的**sendMessageRoutine\(\)** 即可进行新连接的建立或寻找已存在连接并进行转换后报文的发送。

   - Server必须填写消息类型**HeaderType**、连接上下文**ConnData**、对端公钥**PeerPk**和消息明文**Message**（如无特殊情况都采用消息压缩）。将填写好的**MsgAssembler**发给**Device.SendMsgToPacket\(\)** 即可进行转换后报文的发送。

   - 如果存在交互，可以直接使用上一条获得的 **\*PacketParserData**填入**MsgAssembler**结构体的**PrevParserData**字段，从而可以省略填写**RemoteAddr**、**ConnData**、**PeerPk**。

   - 如果请求期待回复数据，需要创建一个接收**PacketParserData**的通道，并对**MsgAssembler**结构体的**ResponseMsgCh**字段赋值。
