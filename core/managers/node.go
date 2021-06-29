package managers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	psutilnet "github.com/shirou/gopsutil/net"
	"github.com/swarleynunez/superfog/core/bindings"
	"github.com/swarleynunez/superfog/core/docker"
	"github.com/swarleynunez/superfog/core/eth"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"github.com/swarleynunez/superfog/inputs"
	"math"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	// Gas limit of each smart contract function
	DeployControllerGasLimit uint64 = 3600000
	RegisterNodeGasLimit     uint64 = 530000
	SendEventGasLimit        uint64 = 340000
	SendReplyGasLimit        uint64 = 230000
	VoteSolverGasLimit       uint64 = 180000
	SolveEventGasLimit       uint64 = 110000
	RecordContainerGasLimit  uint64 = 370000
	RemoveContainerGasLimit  uint64 = 150000

	// Reputable actions
	SendEventAction       = "sendEvent"
	SendReplyAction       = "sendReply"
	VoteSolverAction      = "voteSolver"
	SolveEventAction      = "solveEvent"
	RecordContainerAction = "recordContainer"
	RemoveContainerAction = "removeContainer"

	// Others
	blueMsgFormat = "\033[1;34m[%s] %s \033[0m\n"
)

var (
	// Unexported and "readonly" global parameters
	_ethc   *ethclient.Client
	_dcli   *client.Client
	_ks     *keystore.KeyStore
	_from   accounts.Account
	_nonce  uint64
	_nmutex *sync.Mutex
	_pmutex *sync.Mutex
	_cinst  *bindings.Controller
	_finst  *bindings.Faucet

	// Errors
	errUnknownTask = errors.New("unknown task")
)

type networks map[string]dockertypes.NetworkStats

//////////
// Init //
//////////
func InitNode(ctx context.Context, deploying bool) {

	// Load .env configuration
	utils.LoadEnv()

	var (
		nodeDir = os.Getenv("ETH_NODE_DIR")
		addr    = os.Getenv("NODE_ADDR")
		pass    = os.Getenv("NODE_PASS")
	)

	// Connect to the Ethereum node
	_ethc = eth.Connect(utils.FormatPath(nodeDir, "geth.ipc"))

	// Connect to the Docker node
	_dcli = docker.Connect(ctx)

	// Load Ethereum keystore
	keypath := utils.FormatPath(nodeDir, "keystore")
	_ks = eth.LoadKeystore(keypath)

	// Load and unlock an Ethereum account
	_from = eth.LoadAccount(_ks, addr, pass)

	// Get loaded Ethereum account nonce
	nonce, err := _ethc.PendingNonceAt(context.Background(), _from.Address)
	utils.CheckError(err, utils.FatalMode)
	_nonce = nonce

	_nmutex = &sync.Mutex{} // Mutex to synchronize access to account nonce
	_pmutex = &sync.Mutex{} // Mutex to synchronize access to network ports

	// Get smart contracts instances
	if !deploying {
		_cinst = controllerInstance()
		_finst = faucetInstance(getFaucetAddress())
	}
}

func InitContainerState(ctx context.Context) {

	// TODO. Testing
	rand.Seed(time.Now().Unix())
	cinfo := inputs.Containers[rand.Intn(len(inputs.Containers))]
	NewContainer(ctx, &cinfo)

	//UpdateContainerState(ctx)
}

// TODO
func UpdateContainerState(ctx context.Context) {

	// Get distributed registry active containers
	ctrs := GetContainerReg()
	for key := range ctrs {
		// Am I the host?
		if ctrs[key].Host == _from.Address {
			cname := GetContainerName(key)
			newRegCtr := false

			// Does the container exist locally?
			ctr := SearchDockerContainers(ctx, "name", cname, true)
			if ctr == nil {
				// Decode container info
				var cinfo types.ContainerInfo
				utils.UnmarshalJSON(ctrs[key].Info, &cinfo)

				// Avoid container renaming searching by cid
				ctr := SearchDockerContainers(ctx, "id", cinfo.Id, true)
				if ctr == nil {
					RemoveContainer(cname)
					NewContainer(ctx, &cinfo)
					newRegCtr = true
				} else {
					SetContainerName(ctx, cinfo.Id, cname)
				}
			}

			// Is the container running?
			if !newRegCtr {
				ctr = SearchDockerContainers(ctx, "name", cname, false)
				if ctr == nil {
					startDockerContainer(ctx, cname)
				}
			}
		}
	}
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
	for i, v := range dio {
		disks = append(disks, &v)
	}

	// Sort disks by name (disk0, disk1)
	sort.Slice(disks, func(i, j int) bool {
		return disks[i].Name < disks[j].Name
	})

	proc, err := process.Processes()
	utils.CheckError(err, utils.WarningMode)*/

	nio, err := psutilnet.IOCounters(false) // Get global net I/O stats (all NICs)
	utils.CheckError(err, utils.WarningMode)

	/*ni, err := net.Interfaces()
	utils.CheckError(err, utils.WarningMode)

	// Store each interface as pointer
	var inets []*net.InterfaceStat
	for i, v := range ni {
		inets = append(inets, &v)
	}

	nc, err := net.Connections("inet") // Only inet connections (tcp, udp)
	utils.CheckError(err, utils.WarningMode)

	// Store each connection as pointer
	var conns []*net.ConnectionStat
	for i, v := range nc {
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
func getContainerInfo(ctx context.Context, ctrNameId string) *types.ContainerInfo { // ctrNameId = id or name

	// Container info
	ctr, err := _dcli.ContainerInspect(ctx, ctrNameId)
	utils.CheckError(err, utils.WarningMode)

	// Container image info
	img, _, err := _dcli.ImageInspectWithRaw(ctx, ctr.Image)
	utils.CheckError(err, utils.WarningMode)

	return &types.ContainerInfo{
		Id:        ctr.ID,
		ImageTag:  ctr.Config.Image,
		ImageArch: img.Architecture,
		ImageOs:   img.Os,
		ImageSize: uint64(img.VirtualSize),
		ContainerSetup: types.ContainerSetup{
			ContainerConfig: types.ContainerConfig{
				CPULimit: uint64(ctr.HostConfig.NanoCPUs),
				MemLimit: uint64(ctr.HostConfig.Memory),
				Volumes:  ctr.HostConfig.Binds,
				Ports:    ctr.HostConfig.PortBindings,
			},
		},
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

	// Container summary (only if the container is running)
	ctr := SearchDockerContainers(ctx, "name", cname, false)

	// Group all NICs
	ns := groupNetworkStats(stats.Networks)

	return &types.State{
		CpuPercent: calculateCpuPercent(&stats.CPUStats, &stats.PreCPUStats),
		MemUsage:   stats.MemoryStats.Usage,
		// Get disk usage (rw size and volumes size)
		DiskUsage:      uint64((*ctr)[0].SizeRw) + getVolumesSize(ctx, &(*ctr)[0].Mounts),
		NetPacketsSent: ns.TxPackets,
		NetPacketsRecv: ns.RxPackets,
	}
}

//////////////
// Handling //
//////////////
/*func RunTask(ctx context.Context, task types.Task, cname string) {

	fmt.Printf(blueMsgFormat, time.Now().Format("15:04:05.000000"), "Running local task (RCID="+strconv.FormatUint(getRegContainerId(cname), 10)+")")

	switch task {
	case types.NewContainerTask:
	case types.RestartContainerTask:
		// Run task
		RestartContainer(ctx, cname)
	case types.StopContainerTask:
	case types.MigrateContainerTask:
	case types.DeleteContainerTask:
	default:
		utils.CheckError(errUnknownTask, utils.WarningMode)
		return
	}
}*/

func RunEventTask(ctx context.Context, etype types.EventType, eventId uint64) {

	// TODO. Ask for the event sender
	switch etype.Task {
	case types.NewContainerTask:
	case types.RestartContainerTask:
	case types.MigrateContainerTask:
		// Get related event metadata
		rcid := etype.Metadata["rcid"].(float64) // TODO

		ctr := getRegContainer(uint64(rcid))

		// Decode container info
		var cinfo types.ContainerInfo
		utils.UnmarshalJSON(ctr.Info, &cinfo)

		// Run task
		fmt.Printf(blueMsgFormat, time.Now().Format("15:04:05.000000"), "Executing event task (RCID="+strconv.FormatUint(uint64(rcid), 10)+")")
		NewContainer(ctx, &cinfo)
		fmt.Printf(blueMsgFormat, time.Now().Format("15:04:05.000000"), "Event task executed (RCID="+strconv.FormatUint(uint64(rcid), 10)+")")
	case types.DeleteContainerTask:
	default:
		utils.CheckError(errUnknownTask, utils.WarningMode)
		return
	}

	// Solve related event
	SolveEvent(eventId)
}

func RunEventEndingTask(ctx context.Context, etype types.EventType) {

	switch etype.Task {
	case types.NewContainerTask:
	case types.RestartContainerTask:
	case types.MigrateContainerTask:
		// Get related event metadata
		rcid := etype.Metadata["rcid"].(float64) // TODO
		cname := GetContainerName(uint64(rcid))  // TODO

		// Run task
		fmt.Printf(blueMsgFormat, time.Now().Format("15:04:05.000000"), "Executing ending task (RCID="+strconv.FormatUint(uint64(rcid), 10)+")")
		DeleteContainer(ctx, cname)
		fmt.Printf(blueMsgFormat, time.Now().Format("15:04:05.000000"), "Ending task executed (RCID="+strconv.FormatUint(uint64(rcid), 10)+")")
	case types.DeleteContainerTask:
	default:
		utils.CheckError(errUnknownTask, utils.WarningMode)
		return
	}
}

///////////
// Tasks //
///////////
func NewContainer(ctx context.Context, cinfo *types.ContainerInfo) {

	// Local actions
	cid := createDockerContainer(ctx, cinfo, "") // Empty cname
	startDockerContainer(ctx, cid)

	// Distributed actions
	RecordContainer(ctx, cid, &cinfo.ContainerType)
}

func RestartContainer(ctx context.Context, cname string) {

	restartDockerContainer(ctx, cname)
}

func DeleteContainer(ctx context.Context, cname string) {

	// Distributed actions
	RemoveContainer(cname)

	// Local actions
	// TODO. Timeout to stop a container
	//stopDockerContainer(ctx, cname)
	removeDockerContainer(ctx, cname)
}

func BackupContainer() {

	// TODO. Backup tasks. Improve flow
	// commit --> create an image from a container (snapshot preserving rw)
	// save, load --> compress and uncompress images (tar or stdin/stdout)
	// volumes --> manual backup or using --volumes-from (temporal container)
}

// About distributed registry
func RecordContainer(ctx context.Context, cid string, ctype *types.ContainerType) {

	cinfo := getContainerInfo(ctx, cid)
	cinfo.ContainerType = *ctype // Set container type
	stime := getContainerStartTime(ctx, cid)
	recordContainerOnReg(cinfo, stime, cid)
}

// About distributed registry
func RemoveContainer(cname string) {

	rcid := getRegContainerId(cname)
	ftime := uint64(time.Now().Unix())
	removeContainerFromReg(rcid, ftime)
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

func checkNodePorts(ctx context.Context, ports nat.PortMap) nat.PortMap {

	// To avoid repeated ports
	var usedPorts []string

	_pmutex.Lock()
	for i := range ports {
		for j := range ports[i] {
			// Container configured port (string format)
			strp := ports[i][j].HostPort

			for {
				// Is the port already used?
				var found bool
				for p := range usedPorts {
					if usedPorts[p] == strp {
						found = true
						break
					}
				}

				if isNodePortAvailable(i.Proto(), "localhost", strp) && !isPortAllocatedByDocker(ctx, strp) && !found {
					ports[i][j].HostPort = strp
					usedPorts = append(usedPorts, strp)
					break
				} else {
					// Set the next port
					nump, err := strconv.ParseUint(strp, 10, 64)
					utils.CheckError(err, utils.WarningMode)
					nump++
					strp = strconv.FormatUint(nump, 10)
				}
			}
		}
	}
	_pmutex.Unlock()

	return ports
}

func isNodePortAvailable(network, host, port string) bool {

	conn, err := net.Dial(network, net.JoinHostPort(host, port))

	if err == nil && conn != nil {
		conn.Close()
		return false
	} else {
		return true
	}
}
