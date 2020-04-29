package types

type NodeSpecs struct {
	Arch                     string
	Cores                    int     // Logical cores number
	Mhz                      float64 // Physical cores frequency
	MemTotal, DiskTotal      uint64  // In bytes
	FileSystem, OS, Hostname string
	BootTime                 uint64 // Unix time
}

type NodeState struct {
	CpuPercent  float64
	MemUsage    uint64 // In bytes
	MemPercent  float64
	DiskUsage   uint64 // In bytes
	DiskPercent float64
	//Disks                                                      []*disk.IOCountersStat
	//Processes                                                  []*process.Process
	NetPacketsSent, NetBytesSent, NetPacketsRecv, NetBytesRecv uint64 // Counting all NICs
	//NetInterfaces                                              []*net.InterfaceStat // The IPs can change
	//NetConnections                                             []*net.ConnectionStat
}
