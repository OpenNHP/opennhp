# NHP-Device架构设计

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
