package types

// Node specifications
type Spec uint8

const (
	CpuSpec Spec = iota
	MemSpec
	DiskSpec
	PktSentSpec
	PktRecvSpec
)

type NodeSpecs struct {
	Arch       string  `json:"arch"`
	Cores      uint64  `json:"cores"`      // Logical cores number
	Mhz        float64 `json:"mhz,string"` // Physical cores frequency
	MemTotal   uint64  `json:"mem"`        // In bytes
	DiskTotal  uint64  `json:"disk"`       // In bytes
	FileSystem string  `json:"fs"`
	OS         string  `json:"os"`
	Hostname   string  `json:"hostname"`
	BootTime   uint64  `json:"boot"` // Unix time
}

type NodeState struct {
	CpuPercent float64 `json:"cpu,string"`
	MemUsage   uint64  `json:"mem,string"`  // Used in bytes
	DiskUsage  uint64  `json:"disk,string"` // Used in bytes
	//Disks                                                      []*disk.IOCountersStat
	//Processes                                                  []*process.Process
	NetPacketsSent uint64 `json:"pkt_sent,string"` // Counting all NICs
	NetPacketsRecv uint64 `json:"pkt_recv,string"` // Counting all NICs
	//NetBytesSent, NetBytesRecv uint64 // Counting all NICs
	//NetInterfaces                                              []*net.InterfaceStat // The IPs can change
	//NetConnections                                             []*net.ConnectionStat
}
