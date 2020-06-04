package daemons

import (
	"context"
	"encoding/json"
	"fmt"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"io"
	"io/ioutil"
)

// Hysteresis cycles
type cycleCounter struct {
	measures uint64
	triggers uint64
}

// Hysteresis cycles by rule name
type cycles map[string]cycleCounter

func getSpecs() *types.NodeSpecs {

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
		Mhz:        ci[0].Mhz,
		MemTotal:   vm.Total,
		DiskTotal:  du.Total,
		FileSystem: du.Fstype,
		OS:         hi.OS,
		Hostname:   hi.Hostname,
		BootTime:   hi.BootTime,
	}
}

func getState() *types.NodeState {

	cp, err := cpu.Percent(0, false) // Total CPU usage (all cores)
	utils.CheckError(err, utils.WarningMode)

	vm, err := mem.VirtualMemory()
	utils.CheckError(err, utils.WarningMode)

	du, err := disk.Usage("/") // File system root path
	utils.CheckError(err, utils.WarningMode)

	// Entirely disk usage
	du.Used = du.Total - du.Free
	du.UsedPercent = (float64(du.Used) / float64(du.Total)) * 100.0

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

	return &types.NodeState{
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

func StartMonitor() {

	/*// Goroutines to receive events
	go watchNewEvent()
	go watchRequiredReplies()
	go watchRequiredVotes()
	go watchEventSolved()

	// Get and parse monitor time interval
	mInter, err := strconv.ParseUint(os.Getenv("MONITOR_INTERVAL"), 10, 64)
	utils.CheckError(err, utils.WarningMode)

	// Get and parse cycle time
	cTime, err := strconv.ParseUint(os.Getenv("CYCLE_TIME"), 10, 64)
	utils.CheckError(err, utils.WarningMode)

	// Node rule cycles
	cycles := cycles{}

	// Main infinite loop
	for {
		time.Sleep(time.Duration(mInter) * time.Millisecond)

		// Check all state rules
		checkStateRules(cycles, mInter, cTime)
	}*/

	test()
}

func test() {

	// Connect to the Docker node
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	utils.CheckError(err, utils.WarningMode)

	// Variables
	ctx := context.Background()
	image := "nginx"

	// Pull a docker image
	out, err := cli.ImagePull(ctx, image, dockertypes.ImagePullOptions{})
	utils.CheckError(err, utils.WarningMode)
	_, err = io.Copy(ioutil.Discard, out)
	utils.CheckError(err, utils.WarningMode)

	// List images
	images, err := cli.ImageList(ctx, dockertypes.ImageListOptions{All: true})
	utils.CheckError(err, utils.WarningMode)
	for i := range images {
		fmt.Println(images[i].RepoTags)
	}

	// Delete all containers
	containers, err := cli.ContainerList(ctx, dockertypes.ContainerListOptions{All: true})
	utils.CheckError(err, utils.WarningMode)
	for i := range containers {
		err = cli.ContainerRemove(ctx, containers[i].ID, dockertypes.ContainerRemoveOptions{Force: true})
		utils.CheckError(err, utils.WarningMode)
	}

	// Set container options
	config := &container.Config{Image: image}
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostPort: "8080",
				},
			},
		},
		AutoRemove: true,
	}

	// Create container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	utils.CheckError(err, utils.WarningMode)
	fmt.Println(resp.ID)

	// Run container
	err = cli.ContainerStart(ctx, resp.ID, dockertypes.ContainerStartOptions{})
	utils.CheckError(err, utils.WarningMode)

	//
	containers, err = cli.ContainerList(ctx, dockertypes.ContainerListOptions{All: true})
	utils.CheckError(err, utils.WarningMode)
	for i := range containers {
		fmt.Print("INFO: ", containers[i], "\n")

		stats, err := cli.ContainerStatsOneShot(ctx, containers[i].ID)
		utils.CheckError(err, utils.WarningMode)

		var containerStats map[string]interface{}
		err = json.NewDecoder(stats.Body).Decode(&containerStats)
		utils.CheckError(err, utils.WarningMode)
		fmt.Println(containerStats)
	}
}
