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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	psutilnet "github.com/shirou/gopsutil/net"
	"github.com/swarleynunez/superfog/core/bindings"
	"github.com/swarleynunez/superfog/core/docker"
	"github.com/swarleynunez/superfog/core/eth"
	"github.com/swarleynunez/superfog/core/onos"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"math"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	// Gas limit for each smart contract function
	/*DeployControllerGasLimit uint64 = 3600000
	RegisterNodeGasLimit     uint64 = 530000
	SendEventGasLimit        uint64 = 340000
	SendReplyGasLimit        uint64 = 230000
	VoteSolverGasLimit       uint64 = 180000
	SolveEventGasLimit       uint64 = 110000
	RecordContainerGasLimit  uint64 = 370000
	RemoveContainerGasLimit  uint64 = 150000*/

	// Reputable actions
	SendEventAction     = "sendEvent"
	SendReplyAction     = "sendReply"
	VoteSolverAction    = "voteSolver"
	SolveEventAction    = "solveEvent"
	RegisterAppAction   = "registerApp"
	RegisterCtrAction   = "registerCtr"
	ActivateCtrAction   = "activateCtr"
	UpdateCtrAction     = "updateCtr"
	UnregisterAppAction = "unregisterApp"
	UnregisterCtrAction = "unregisterCtr"
)

var (
	// Unexported and "readonly" global parameters
	_ethc   *ethclient.Client
	_ks     *keystore.KeyStore
	_from   accounts.Account
	_pmutex *sync.Mutex
	_cinst  *bindings.Controller
	_finst  *bindings.Faucet
	_docc   *client.Client
	_onosc  *onos.Client

	// Errors
	errUnknownTask = errors.New("unknown event task")
)

type networks map[string]dockertypes.NetworkStats

//////////
// Init //
//////////
func InitNode(ctx context.Context, deploying bool) {

	// Load .env configuration
	utils.LoadEnv()

	var (
		nodeDir = utils.GetEnv("ETH_NODE_DIR")
		// TODO: test --> addr    = utils.GetEnv("NODE_ADDR")
		pass = utils.GetEnv("NODE_PASS")
	)

	// Connect to the Ethereum local node
	_ethc = eth.Connect(utils.FormatPath(nodeDir, "geth.ipc"))

	// Load Ethereum keystore
	keypath := utils.FormatPath(nodeDir, "keystore")
	_ks = eth.LoadKeystore(keypath)

	// Load and unlock an Ethereum account
	// TODO: test --> _from = eth.LoadAccount(_ks, addr, pass)
	_from = eth.LoadAccount(_ks, pass)

	// Debug
	fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Loaded account: "+_from.Address.String(), "\n")

	_pmutex = &sync.Mutex{} // Mutex to synchronize access to network ports

	// Get smart contracts instances
	if !deploying {
		_cinst = controllerInstance()
		_finst = faucetInstance(getFaucetContract())

		// Debug
		fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Loaded controller address: ", utils.GetEnv("CONTROLLER_ADDR"), "\n")
	}

	// Connect to the Docker local node
	_docc = docker.Connect(ctx)

	// Connect to a cluster ONOS controller
	_onosc = onos.Connect()

	// Debug
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			for {
				// Create and configure a transaction
				tx := eth.SignedEtherTransaction(ctx, _ethc, _ks, _from, "12345678", common.HexToAddress("0x22dbCF83D13a84C53893903189Ee33d1115C0259"), 0)

				// Send transaction
				err := _ethc.SendTransaction(ctx, tx)

				if err != nil {
					continue
				} else {
					break
				}
			}
		}
	}()
}

func InitNodeState(ctx context.Context) {

	// Get distributed registry active containers
	ctrs := GetActiveContainers()
	for rcid := range ctrs {
		// Am I the host?
		if IsContainerHost(rcid, _from.Address) {
			cname := GetContainerName(rcid)

			// Does the container exist locally?
			c := SearchDockerContainers(ctx, "name", cname, true)
			if c == nil {
				// Decode container info
				var cinfo types.ContainerInfo
				utils.UnmarshalJSON(ctrs[rcid].Info, &cinfo)

				// TODO: check ONOS SDN state when nodes initialize its container state
				go func() {
					createDockerContainer(ctx, &cinfo, cname)
					startDockerContainer(ctx, cname)
				}()
			} else {
				// Is the container running?
				c = SearchDockerContainers(ctx, "name", cname, false)
				if c == nil {
					startDockerContainer(ctx, cname)
				}
			}
		}
	}
}

/////////////
// Getters //
/////////////
func GetFromAccount() common.Address {
	return _from.Address
}

func GetControllerInst() *bindings.Controller {
	return _cinst
}

func GetSpecs() *types.NodeSpecs {

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
		Arch:      hi.KernelArch,
		Cores:     uint64(cores),
		CpuMhz:    ci[0].Mhz,
		MemTotal:  vm.Total,
		DiskTotal: du.Total,
		OS:        hi.OS,
		IP:        getNodeIP(),
	}
}

func GetState() *types.State {

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
	sort.Slice(disks, func(i, j int{
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
		CpuUsage:  cp[0],
		MemUsage:  vm.Used,
		DiskUsage: du.Used,
		//Disks:          disks,
		//Processes:      proc,
		NetPacketsSent: nio[0].PacketsSent,
		NetPacketsRecv: nio[0].PacketsRecv,
		//NetBytesSent:   nio[0].BytesSent,
		//NetBytesRecv:   nio[0].BytesRecv,
		//NetInterfaces:  inets,
		//NetConnections: conns,
	}
}

func getContainerState(ctx context.Context, cname string) *types.State {

	// Get current stats
	cs, err := _docc.ContainerStatsOneShot(ctx, cname)
	utils.CheckError(err, utils.WarningMode)

	// Decode stats
	var stats dockertypes.StatsJSON
	err = json.NewDecoder(cs.Body).Decode(&stats)
	utils.CheckError(err, utils.WarningMode)

	// TODO. Container summary (only if the container is running)
	ctr := SearchDockerContainers(ctx, "name", cname, false)

	// Group all NICs
	ns := groupNetworkStats(stats.Networks)

	return &types.State{
		CpuUsage:       calculateCpuPercent(&stats.CPUStats, &stats.PreCPUStats),
		MemUsage:       stats.MemoryStats.Usage,
		DiskUsage:      uint64(ctr[0].SizeRw) + getVolumesSize(ctx, ctr[0].Mounts), // Get disk usage (rw size and volumes size)
		NetPacketsSent: ns.TxPackets,
		NetPacketsRecv: ns.RxPackets,
	}
}

//////////////
// Handling //
//////////////
// Tasks to execute when the sender and the solver are the same node
func RunTask(ctx context.Context, event *types.Event, eid uint64) {

	// Decode event type
	var etype types.EventType
	utils.UnmarshalJSON(event.EType, &etype)

	switch etype.RequiredTask {
	case types.NewContainerTask:
		if event.Rcid > 0 {
			// Get container linked to the event
			ctr := GetContainer(event.Rcid)

			// Decode container info
			var cinfo types.ContainerInfo
			utils.UnmarshalJSON(ctr.Info, &cinfo)

			// Run task
			NewContainer(ctx, &cinfo, ctr.Appid, event.Rcid, true)
		}
	case types.MigrateContainerTask:
		if event.Rcid > 0 {
			// TODO: run tasks to balance cluster nodes (resource usage)?
			RestartContainer(ctx, GetContainerName(event.Rcid))
		}
	default:
		utils.CheckError(errUnknownTask, utils.WarningMode)
		return
	}

	// Solve related event
	err := SolveEvent(ctx, eid)
	utils.CheckError(err, utils.WarningMode)
}

// Tasks to execute when the cluster selects a solver
func RunEventTask(ctx context.Context, event *types.Event, eid uint64) {

	// Decode event type
	var etype types.EventType
	utils.UnmarshalJSON(event.EType, &etype)

	switch etype.RequiredTask {
	case types.NewContainerTask, types.MigrateContainerTask:
		if event.Rcid > 0 {
			// Get container linked to the event
			ctr := GetContainer(event.Rcid)

			// Decode container info
			var cinfo types.ContainerInfo
			utils.UnmarshalJSON(ctr.Info, &cinfo)

			// Run event task
			NewContainer(ctx, &cinfo, ctr.Appid, event.Rcid, true)
		}
	default:
		utils.CheckError(errUnknownTask, utils.WarningMode)
		return
	}

	// Solve related event
	err := SolveEvent(ctx, eid)
	utils.CheckError(err, utils.WarningMode)
}

// Tasks to execute when the cluster solve an event
func RunEventEndingTask(ctx context.Context, event *types.Event) {

	// Decode event type
	var etype types.EventType
	utils.UnmarshalJSON(event.EType, &etype)

	switch etype.RequiredTask {
	case types.NewContainerTask:
		// Do nothing
	case types.MigrateContainerTask:
		if event.Rcid > 0 {
			// Get container linked to the event
			ctr := GetContainer(event.Rcid)

			// Run ending task
			StopContainer(ctx, ctr.Appid, event.Rcid, true)
		}
	default:
		utils.CheckError(errUnknownTask, utils.WarningMode)
		return
	}
}

///////////
// Tasks //
///////////
// onosaction: require ONOS SDN additional actions?
func NewContainer(ctx context.Context, cinfo *types.ContainerInfo, appid, rcid uint64, onosaction bool) {

	// Does the container exist locally?
	cname := GetContainerName(rcid)
	c := SearchDockerContainers(ctx, "name", cname, true)
	if c == nil {
		createDockerContainer(ctx, cinfo, cname)
		startDockerContainer(ctx, cname)
	} else {
		// Is the container running?
		c = SearchDockerContainers(ctx, "name", cname, false)
		if c == nil {
			startDockerContainer(ctx, cname)
		}
	}

	// TODO: integrate Docker HEALTHCHECK
	time.Sleep(7 * time.Second)

	// ONOS SDN plugin
	if onosaction {
		ONOSAddVSInstance(ctx, appid, rcid, getNodeIP())
		ONOSActivateVirtualService(appid)
	}
}

func StartContainer(ctx context.Context, cname string) {

	// Does the container exist locally?
	c := SearchDockerContainers(ctx, "name", cname, true)
	if c != nil {
		// Is the container running?
		c = SearchDockerContainers(ctx, "name", cname, false)
		if c == nil {
			startDockerContainer(ctx, cname)
		}
	}
}

func RestartContainer(ctx context.Context, cname string) {

	// Does the container exist locally?
	c := SearchDockerContainers(ctx, "name", cname, true)
	if c != nil {
		// Is the container running?
		c = SearchDockerContainers(ctx, "name", cname, false)
		if c == nil {
			startDockerContainer(ctx, cname)
		} else {
			restartDockerContainer(ctx, cname)
		}
	}
}

// Rename a container (temporarily, before remove the container)
/*func RenameContainer(ctx context.Context, cname string) (cid string) {

	// Does the container exist locally?
	c := SearchDockerContainers(ctx, "name", cname, true)
	if c != nil {
		cid = c[0].ID
		renameDockerContainer(ctx, cname, cid)
	}

	return
}*/

// onosaction: require ONOS SDN additional actions?
func StopContainer(ctx context.Context, appid, rcid uint64, onosaction bool) {

	// ONOS SDN plugin
	if onosaction {
		ONOSDeleteVSInstance(appid, rcid)
	}

	// Does the container exist locally?
	cname := GetContainerName(rcid)
	c := SearchDockerContainers(ctx, "name", cname, true)
	if c != nil {
		// Is the container running?
		c = SearchDockerContainers(ctx, "name", cname, false)
		if c != nil {
			stopDockerContainer(ctx, cname)
		}
	}
}

// onosaction: require ONOS SDN additional actions?
func RemoveContainer(ctx context.Context, appid, rcid uint64, onosaction bool) {

	// ONOS SDN plugin
	if onosaction {
		ONOSDeleteVSInstance(appid, rcid)
	}

	// Does the container exist locally?
	cname := GetContainerName(rcid)
	c := SearchDockerContainers(ctx, "name", cname, true)
	if c != nil {
		// Is the container running?
		c = SearchDockerContainers(ctx, "name", cname, false)
		if c != nil {
			stopDockerContainer(ctx, cname)
			removeDockerContainer(ctx, cname)
		} else {
			removeDockerContainer(ctx, cname)
		}
	}
}

func RemoveDCRApplication(ctx context.Context, appid uint64) error {

	err := UnregisterApplication(ctx, appid)
	if err == nil {
		ONOSDeleteVirtualService(appid)
	}

	return err
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

func getVolumesSize(ctx context.Context, mnts []dockertypes.MountPoint) (r uint64) {

	// Get docker disk usage info (docker system df -v)
	resp, err := _docc.DiskUsage(ctx)
	utils.CheckError(err, utils.WarningMode)

	// Search and compare volumes by name
	for i := range mnts {
		for j := range resp.Volumes {
			if mnts[i].Name == resp.Volumes[j].Name {
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

func getNodeIP() net.IP {

	conn, err := net.Dial("udp", "8.8.8.8:80")
	utils.CheckError(err, utils.WarningMode)
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP // Type assertion of an interface type
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
		// Successful connection
		conn.Close()
		return false
	} else {
		return true
	}
}
