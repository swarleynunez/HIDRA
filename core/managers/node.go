package managers

import (
	"context"
	"encoding/json"
	"fmt"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/swarleynunez/superfog/core/bindings"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"math"
)

const (
	// Gas limit of each smart contract function
	DeployControllerGasLimit uint64 = 3500000
	RegisterNodeGasLimit     uint64 = 530000
	SendEventGasLimit        uint64 = 350000
	SendReplyGasLimit        uint64 = 230000
	VoteSolverGasLimit       uint64 = 180000
	SolveEventGasLimit       uint64 = 110000
	RecordContainerGasLimit  uint64 = 310000
	RemoveContainerGasLimit  uint64 = 130000

	// Reputable actions
	SendEventAction       = "sendEvent"
	SendReplyAction       = "sendReply"
	VoteSolverAction      = "voteSolver"
	SolveEventAction      = "solveEvent"
	RecordContainerAction = "recordContainer"
	RemoveContainerAction = "removeContainer"
)

// Unexported and "readonly" global parameters
var (
	_ethc  *ethclient.Client  // Ethereum client
	_dcli  *client.Client     // Docker client
	_ks    *keystore.KeyStore // Ethereum keystore
	_from  accounts.Account   // Selected Ethereum account
	_cinst *bindings.Controller
	_finst *bindings.Faucet
)

type networks map[string]dockertypes.NetworkStats

/////////////
// Setters //
/////////////
func InitNode(ethc *ethclient.Client, dcli *client.Client, ks *keystore.KeyStore, from accounts.Account) {

	_ethc = ethc
	_dcli = dcli
	_ks = ks
	_from = from

	// Deploy controller smart contract or get an instance
	_cinst = controllerInstance()

	// Faucet instance
	_finst = faucetInstance(getFaucetAddress())

	// Register node in the network
	registerNode()
}

/////////////
// Getters //
/////////////
func GetFromAccount() accounts.Account {
	return _from
}

func GetControllerInst() *bindings.Controller {
	return _cinst
}

func getContainerByName(ctx context.Context, cname string) *dockertypes.Container {

	filter := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: cname})
	ctr, err := _dcli.ContainerList(ctx, dockertypes.ContainerListOptions{Size: true, Filters: filter})
	utils.CheckError(err, utils.WarningMode)

	return &ctr[0]
}

///////////
// Specs //
///////////
func GetNodeSpecs() *types.NodeSpecs {

	hi, err := host.Info()
	utils.CheckError(err, utils.WarningMode)

	cores, err := cpu.Counts(true) // Counting physical and logical cores
	utils.CheckError(err, utils.WarningMode)

	ci, err := cpu.Info()
	utils.CheckError(err, utils.WarningMode)

	vm, err := mem.VirtualMemory()
	utils.CheckError(err, utils.WarningMode)

	du, err := disk.Usage("/") // File system root path
	utils.CheckError(err, utils.WarningMode)

	return &types.NodeSpecs{
		Arch:       hi.KernelArch,
		Cores:      uint64(cores),
		CpuMhz:     ci[0].Mhz,
		MemTotal:   vm.Total,
		DiskTotal:  du.Total,
		FileSystem: du.Fstype,
		Os:         hi.OS,
		Hostname:   hi.Hostname,
		BootTime:   hi.BootTime,
	}
}

func GetNodeState() *types.State {

	cp, err := cpu.Percent(0, false) // Total CPU usage (all cores)
	utils.CheckError(err, utils.WarningMode)

	vm, err := mem.VirtualMemory()
	utils.CheckError(err, utils.WarningMode)

	du, err := disk.Usage("/") // File system root path
	utils.CheckError(err, utils.WarningMode)

	// Entirely disk usage
	du.Used = du.Total - du.Free
	du.UsedPercent = (float64(du.Used) / float64(du.Total)) * 100.0

	/*dio, err := disk.IOCounters()
	utils.CheckError(err, utils.WarningMode)

	// Store disks information in a slice
	var disks []*disk.IOCountersStat
	for _, v := range dio {
		disks = append(disks, &v)
	}

	// Sort disks by name (disk0, disk1)
	sort.Slice(disks, func(i, j int) bool {
		return disks[i].Name < disks[j].Name
	})

	proc, err := process.Processes()
	utils.CheckError(err, utils.WarningMode)*/

	nio, err := net.IOCounters(false) // Get global net I/O stats (all NICs)
	utils.CheckError(err, utils.WarningMode)

	/*ni, err := net.Interfaces()
	utils.CheckError(err, utils.WarningMode)

	// Store each interface as pointer
	var inets []*net.InterfaceStat
	for _, v := range ni {
		inets = append(inets, &v)
	}

	nc, err := net.Connections("inet") // Only inet connections (tcp, udp)
	utils.CheckError(err, utils.WarningMode)

	// Store each connection as pointer
	var conns []*net.ConnectionStat
	for _, v := range nc {
		conns = append(conns, &v)
	}*/

	return &types.State{
		CpuPercent: cp[0],
		MemUsage:   vm.Used,
		DiskUsage:  du.Used,
		//Disks:          disks,
		//Processes:      proc,
		NetPacketsSent: nio[0].PacketsSent,
		//NetBytesSent:   nio[0].BytesSent,
		NetPacketsRecv: nio[0].PacketsRecv,
		//NetBytesRecv:   nio[0].BytesRecv,
		//NetInterfaces:  inets,
		//NetConnections: conns,
	}
}

// In order to record a container in the distributed registry
func getContainerInfo(ctx context.Context, cid string) *types.ContainerInfo {

	// Container info
	ctr, err := _dcli.ContainerInspect(ctx, cid)
	utils.CheckError(err, utils.WarningMode)

	// Container image info
	img, _, err := _dcli.ImageInspectWithRaw(ctx, ctr.Image)
	utils.CheckError(err, utils.WarningMode)

	return &types.ContainerInfo{
		ContainerConfig: types.ContainerConfig{
			ImageTag:    ctr.Config.Image,
			CPULimit:    uint64(ctr.HostConfig.NanoCPUs),
			MemLimit:    uint64(ctr.HostConfig.Memory),
			VolumeBinds: ctr.HostConfig.Binds,
			Ports:       ctr.HostConfig.PortBindings,
		},
		IPAddress: ctr.NetworkSettings.IPAddress, // TODO
		ImageArch: img.Architecture,
		ImageOs:   img.Os,
		ImageSize: uint64(img.VirtualSize),
	}
}

func getContainerState(ctx context.Context, cname string) *types.State {

	// Get current stats
	cs, err := _dcli.ContainerStatsOneShot(ctx, cname)
	utils.CheckError(err, utils.WarningMode)

	// Decode stats
	var stats dockertypes.StatsJSON
	err = json.NewDecoder(cs.Body).Decode(&stats)
	utils.CheckError(err, utils.WarningMode)

	// Container summary
	ctr := getContainerByName(ctx, cname)

	// Group all NICs
	ns := groupNetworkStats(stats.Networks)

	return &types.State{
		CpuPercent: calculateCpuPercent(&stats.CPUStats, &stats.PreCPUStats),
		MemUsage:   stats.MemoryStats.Usage,
		// Get disk usage (rw size and volumes size)
		DiskUsage:      uint64(ctr.SizeRw) + getVolumesSize(ctx, &ctr.Mounts),
		NetPacketsSent: ns.TxPackets,
		NetPacketsRecv: ns.RxPackets,
	}
}

//////////////
// Handling //
//////////////
func RunTask(ctx context.Context, task types.Task) {

	fmt.Print("DEBUG: Running task\n")
}

func RunEventTask(ctx context.Context, eid uint64, task types.Task) {

	fmt.Print("DEBUG: Running event task (EID=", eid, ")\n")
}

func RunEventEndingTask(ctx context.Context, eid uint64, task types.Task) {

	fmt.Print("DEBUG: Running event ending task (EID=", eid, ")\n")
}

///////////
// Tasks //
///////////
func NewContainer(ctx context.Context) (cid string) { // TODO. Testing

	// Check and format tag
	imgTag, err := utils.FormatImageTag("nginx")
	utils.CheckError(err, utils.WarningMode)

	ctype := types.ContainerType{
		Impact:      5,
		MainSpec:    types.CpuSpec,
		ServiceType: types.WebServerServ,
	}
	cc := types.ContainerConfig{
		ImageTag: imgTag,
		CPULimit: 0.5 * 1e9,
		MemLimit: 536870912,
		VolumeBinds: []string{ // Volumes created automatically
			"vol:/vol",
		},
		Ports: nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostPort: "8080",
				},
			},
		},
	}

	cid = createContainer(ctx, &cc)
	startContainer(ctx, cid, &ctype)

	return
}

func DeleteContainer(ctx context.Context, cname string) {

	stopContainer(ctx, cname)

	// TODO. Backup tasks
	// commit --> create an image from a container (snapshot preserving rw)
	// save, load --> compress and uncompress images (tar or stdin/stdout)
	// volumes --> manual backup or using --volumes-from (temporal container)

	removeContainer(ctx, cname)
}

/////////////
// Helpers //
/////////////
func calculateCpuPercent(cpu, precpu *dockertypes.CPUStats) (pct float64) {

	// Container and system cpu times variation
	ctrDelta := float64(cpu.CPUUsage.TotalUsage) - float64(precpu.CPUUsage.TotalUsage)
	sysDelta := float64(cpu.SystemUsage) - float64(precpu.SystemUsage)

	if ctrDelta > 0.0 && sysDelta > 0.0 {
		cores := float64(len(cpu.CPUUsage.PercpuUsage)) // Number of cores
		pct = (ctrDelta / sysDelta) * cores * 100.0
	}

	return
}

func getVolumesSize(ctx context.Context, mnts *[]dockertypes.MountPoint) (r uint64) {

	// Get docker disk usage info (docker system df -v)
	resp, err := _dcli.DiskUsage(ctx)
	utils.CheckError(err, utils.WarningMode)

	// Search and compare volumes by name
	for i := range *mnts {
		for j := range resp.Volumes {
			if (*mnts)[i].Name == resp.Volumes[j].Name {
				// Volume stats
				count := resp.Volumes[j].UsageData.RefCount // Number of containers using this volume
				size := resp.Volumes[j].UsageData.Size

				// Divide volume size among containers which use it
				if count > 0 && size > 0 {
					r += uint64(math.Ceil(float64(size) / float64(count))) // Round up
				}
			}
		}
	}

	return
}

func groupNetworkStats(net networks) (ns dockertypes.NetworkStats) {

	for i := range net {
		//ns.RxBytes += net[i].RxBytes
		ns.RxPackets += net[i].RxPackets
		//ns.RxErrors += net[i].RxErrors
		//ns.RxDropped += net[i].RxDropped
		//ns.TxBytes += net[i].TxBytes
		ns.TxPackets += net[i].TxPackets
		//ns.TxErrors += net[i].TxErrors
		//ns.TxDropped += net[i].TxDropped
	}

	return
}
