package de

const (
	MaxConcurrentConnection      = 256
	DefaultConnectionTimeoutMs   = 30 * 1000 // 30 seconds to delete idle connection
	PacketQueueSizePerConnection = 64        // nhp agent does not need large transactions
	DoType_Default               = "ZTDO"    //The DHP protocol enforces encryption by default, and its core data unit is the Zero Trust Data Object (ZTDO)。
	DoType_Other                 = "OTHER"
)
