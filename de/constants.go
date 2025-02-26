package de

const (
	MaxConcurrentConnection      = 256
	DefaultConnectionTimeoutMs   = 30 * 1000 // 30 seconds to delete idle connection
	PacketQueueSizePerConnection = 8         // nhp agent does not need large transactions
	DoType_Default               = "ZTDO"    //DHP协议默认的数据格式“零信任数据对象ZTDO”。
	DoType_Other                 = "OTHER"   //DHP协议的其他数据格式,暂时不用”。
)
