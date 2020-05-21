package types

type NodeSpecs struct {
	Arch       string `json:"arch"`
	Cores      uint64 `json:"cores"` // Logical cores number
	Mhz        string `json:"mhz"`   // Physical cores frequency
	MemTotal   uint64 `json:"mem"`   // In bytes
	DiskTotal  uint64 `json:"disk"`  // In bytes
	FileSystem string `json:"fs"`
	OS         string `json:"os"`
	Hostname   string `json:"hostname"`
	BootTime   uint64 `json:"boot"` // Unix time
}

type NodeState struct {
	CpuPercent string `json:"cpu"`
	MemUsage   uint64 `json:"mem"`  // Used in bytes
	DiskUsage  uint64 `json:"disk"` // Used in bytes
	//Disks                                                      []*disk.IOCountersStat
	//Processes                                                  []*process.Process
	NetPacketsSent uint64 `json:"pkt_sent"`
	NetPacketsRecv uint64 `json:"pkt_recv"` // Counting all NICs
	//NetBytesSent, NetBytesRecv uint64 // Counting all NICs
	//NetInterfaces                                              []*net.InterfaceStat // The IPs can change
	//NetConnections                                             []*net.ConnectionStat
}
