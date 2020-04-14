package daemon

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
)

func InitState() (*types.HostSpecs, *types.HostState) {

	hi, _ := host.Info()

	cores, err := cpu.Counts(true) // Counting physical and logical cores
	utils.CheckError(err, utils.WarningMode)

	ci, err := cpu.Info()
	utils.CheckError(err, utils.WarningMode)

	vm, err := mem.VirtualMemory()
	utils.CheckError(err, utils.WarningMode)

	du, err := disk.Usage("/") // File system root path
	utils.CheckError(err, utils.WarningMode)

	specs := &types.HostSpecs{
		Arch:       hi.KernelArch,
		Cores:      cores,
		Mhz:        ci[0].Mhz,
		MemTotal:   vm.Total,
		DiskTotal:  du.Total,
		FileSystem: du.Fstype,
		OS:         hi.OS,
		Hostname:   hi.Hostname,
		BootTime:   hi.BootTime,
		Uptime:     hi.Uptime,
	}

	return specs, UpdateState()
}

func UpdateState() *types.HostState {

	cp, err := cpu.Percent(0, false) // Total CPU usage (all cores)
	utils.CheckError(err, utils.WarningMode)

	vm, err := mem.VirtualMemory()
	utils.CheckError(err, utils.WarningMode)

	du, err := disk.Usage("/") // File system root path
	utils.CheckError(err, utils.WarningMode)

	//dio, err := disk.IOCounters()
	//utils.CheckError(err, utils.WarningMode)
	//
	//// Store disks information in a slice
	//var disks []*disk.IOCountersStat
	//for _, v := range dio {
	//	disks = append(disks, &v)
	//}
	//
	//// Sort disks by name (disk0, disk1)
	//sort.Slice(disks, func(i, j int) bool {
	//	return disks[i].Name < disks[j].Name
	//})
	//
	//proc, err := process.Processes()
	//utils.CheckError(err, utils.WarningMode)

	nio, err := net.IOCounters(false) // Get global net I/O stats (all NICs)
	utils.CheckError(err, utils.WarningMode)

	//ni, err := net.Interfaces()
	//utils.CheckError(err, utils.WarningMode)
	//
	//// Store each interface as pointer
	//var inets []*net.InterfaceStat
	//for _, v := range ni {
	//	inets = append(inets, &v)
	//}
	//
	//nc, err := net.Connections("inet") // Only inet connections (tcp, udp)
	//utils.CheckError(err, utils.WarningMode)
	//
	//// Store each connection as pointer
	//var conns []*net.ConnectionStat
	//for _, v := range nc {
	//	conns = append(conns, &v)
	//}

	return &types.HostState{
		CpuPercent:  cp[0],
		MemUsage:    vm.Used,
		MemPercent:  vm.UsedPercent,
		DiskUsage:   du.Used,
		DiskPercent: du.UsedPercent,
		//Disks:          disks,
		//Processes:      proc,
		NetPacketsSent: nio[0].PacketsSent,
		NetBytesSent:   nio[0].BytesSent,
		NetPacketsRecv: nio[0].PacketsRecv,
		NetBytesRecv:   nio[0].BytesRecv,
		//NetInterfaces:  inets,
		//NetConnections: conns,
	}
}
