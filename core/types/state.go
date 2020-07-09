package types

// Node or container state at a specific time
type State struct {
	CpuPercent float64 `json:"cpu,string"`
	MemUsage   uint64  `json:"mem,string"`  // Used in bytes
	DiskUsage  uint64  `json:"disk,string"` // Used in bytes
	//Disks                                                      []*disk.IOCountersStat
	//Processes                                                  []*process.Process
	NetPacketsSent uint64 `json:"psent,string"` // Counting all NICs
	NetPacketsRecv uint64 `json:"precv,string"` // Counting all NICs
	//NetBytesSent, NetBytesRecv uint64 // Counting all NICs
	//NetInterfaces                                              []*net.InterfaceStat // The IPs can change
	//NetConnections                                             []*net.ConnectionStat
}
