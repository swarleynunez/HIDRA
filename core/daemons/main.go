package daemons

import (
	"context"
	"fmt"
	"github.com/swarleynunez/hidra/core/managers"
	"github.com/swarleynunez/hidra/core/types"
	"github.com/swarleynunez/hidra/core/utils"
	"strconv"
	"time"
)

func Run(ctx context.Context, iface string) {

	// MonitorV1 //
	/*// Rule cycle counter (rcc) per rule
	rccs := map[string]types.CycleCounter{}

	// Cache to avoid duplicated events per container
	ctrCache := map[uint64]bool{}

	minter, err := strconv.ParseUint(utils.GetEnv("MONITOR_INTERVAL"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	ctime, err := strconv.ParseUint(utils.GetEnv("CYCLE_TIME"), 10, 64)
	utils.CheckError(err, utils.FatalMode)*/

	// MonitorV2 //
	mmp, err := strconv.ParseUint(utils.GetEnv("MAX_MONITORED_PKTS"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	lossProb, err := strconv.ParseUint(utils.GetEnv("PKT_LOSS_PROB"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	maxLatency, err := strconv.ParseUint(utils.GetEnv("PKT_MAX_LATENCY"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	epTime, err := strconv.ParseUint(utils.GetEnv("EPOCH_TIME"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	lossProbTh, err := strconv.ParseUint(utils.GetEnv("LOSS_PROB_THRESHOLD"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	latTh, err := strconv.ParseUint(utils.GetEnv("LATENCY_THRESHOLD"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	// Data structures
	nodeStore := types.NodeStore{}

	// Experiments
	latencies := make(map[uint64]types.EventTimes)
	pktCounter := types.PacketCounter{Max: mmp}

	// Watchers to receive blockchain events
	go WatchNewEvent(ctx, latencies, nodeStore)
	go WatchRequiredReplies(ctx)
	go WatchRequiredVotes(ctx)
	go WatchEventSolved(ctx, latencies)
	go WatchApplicationRegistered()
	go WatchContainerRegistered(ctx)
	//go WatchContainerUpdated(ctx)
	//go WatchContainerUnregistered(ctx)

	// TODO: check node/Docker running ports (also check registered ports in DCR)
	// Recover node state from DCR
	//managers.InitNodeState(ctx)

	// Get node network info
	nodeIP, nodePort := managers.GetNodeIPFromAddress(managers.GetFromAccount())

	// Debug
	fmt.Print("--> Node network info: ", nodeIP+":"+nodePort, "\n")
	fmt.Print("--> Packet simulator config:\n")
	fmt.Print("		Packet loss probability: ", lossProb, "%\n")
	fmt.Print("		Loss probability threshold: ", lossProbTh, "%\n")
	fmt.Print("		Packet maximum latency: ", maxLatency, "ms\n")
	fmt.Print("		Latency threshold: ", latTh, "ms\n\n")

	// Main loop V1
	/*go printEventLatencies(args)
	for {
		time.Sleep(time.Duration(minter) * time.Millisecond)

		// Check all state rules
		checkStateRules(ctx, rccs, minter, ctime, ctrCache)
	}*/

	// Main loop V2
	go monitorNetwork(iface, nodePort, lossProb, maxLatency, nodeStore, &pktCounter)
	for {
		time.Sleep(time.Duration(epTime) * time.Second)

		// In each epoch
		go updateNodeReputations(nodeStore, lossProbTh, latTh)
	}
}

/*func printEventLatencies(args []string) {

	count, err := strconv.Atoi(args[1])
	utils.CheckError(err, utils.FatalMode)

	for {
		b, err := os.ReadFile(args[0])
		utils.CheckError(err, utils.FatalMode)

		if strings.Count(string(b), "EventSolved") == count {
			var total float64
			for _, v := range latencies {
				latency := float64(v.end-v.start) / 1e9
				fmt.Println(latency)
				total += latency
			}
			fmt.Printf("LATENCY %.2f", total/float64(len(latencies)))
			return
		}

		time.Sleep(1000 * time.Millisecond)
	}
}*/
