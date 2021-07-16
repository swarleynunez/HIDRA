package types

import "math/big"

type ClusterState struct {
	NodeCount   uint64
	NextEventId uint64
	NextAppId   uint64
	NextCtrId   uint64
	DeployedAt  *big.Int // Unix time
}

type ClusterConfig struct {
	InitNodeRep int64
	NodesThld   uint8 // Percentage threshold to calculate required nodes (0-100)
	VotesThld   uint8 // Percentage threshold to calculate required votes (0-100)
}

// Node or container state at a specific time
type State struct {
	CpuUsage  float64 `json:"cpu,string"` // In percentage
	MemUsage  uint64  `json:"mem"`        // In bytes
	DiskUsage uint64  `json:"disk"`       // In bytes
	//Disks          []*disk.IOCountersStat
	//Processes      []*process.Process
	NetPacketsSent uint64 `json:"psent"` // Counting all NICs
	NetPacketsRecv uint64 `json:"precv"` // Counting all NICs
	//NetBytesSent   uint64 `json:"bsent,string"` // Counting all NICs
	//NetBytesRecv   uint64 `json:"brecv,string"` // Counting all NICs
	//NetInterfaces  []*net.InterfaceStat // The IPs can change
	//NetConnections []*net.ConnectionStat
}
