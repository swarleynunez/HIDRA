package daemons

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/hidra/core/managers"
	"github.com/swarleynunez/hidra/core/types"
	"github.com/swarleynunez/hidra/core/utils"
	"slices"
	"strconv"
	"time"
)

const (
	//blueInfoFormat = "\033[1;34m[%d] %s (Limit: %v, Usage: %v)\033[0m\n"
	blueInfoFormat = "[%d] %s (Limit: %v, Usage: %v)\n"
)

var (
	errUnknownSpec         = errors.New("unknown specification")
	errBoundNotImplemented = errors.New("spec bound type not implemented")
	errUnknownAction       = errors.New("unknown rule action")
	errNoContainersFound   = errors.New("no containers found")
	errReputationDraw      = errors.New("reputation draw")
)

// MonitorV1 //
func runRuleAction(ctx context.Context, rule *types.Rule, ccache map[uint64]bool, usage interface{}) {

	switch rule.Action {
	case types.SendEventAction:
		rcid, err := selectContainer(ctx, ccache)
		if err == nil {
			ccache[rcid] = true

			// Encapsulate event type
			etype := types.EventType{
				RequiredTask: types.MigrateContainerTask,
				Resource:     rule.Resource,
			}

			// Debug
			// fmt.Print("[", time.Now().UnixMilli(), "] ", "Sending an event...\n")

			go func() {
				err = managers.SendEvent(ctx, &etype, rcid)
				if err != nil {
					ccache[rcid] = false
					utils.CheckError(err, utils.WarningMode)
				}
			}()
		}
		fallthrough
	case types.ProceedAction:
		// Execute specific and local stuff
		if rule.Action == types.ProceedAction { // Due to the fallthrough

		}
		fallthrough
	case types.LogAction:
		// Save log into a file, send log to a remote server...
		fallthrough
	case types.WarnAction:
		fmt.Printf(blueInfoFormat, time.Now().UnixMilli(), rule.Msg, rule.Limit, usage)
	case types.IgnoreAction:
		// Do nothing
	default:
		utils.CheckError(errUnknownAction, utils.WarningMode)
		return
	}
}

// Select an event solver according to spec metrics
/*func selectSolver(eid uint64) (addr common.Address) {

	event := managers.GetEvent(eid)
	replies := managers.GetEventReplies(eid)

	// Decode event type
	var etype types.EventType
	utils.UnmarshalJSON(event.EType, &etype)

	// Current best value (variable for different value types)
	var best interface{}

	for _, v := range replies {
		// Preventing the sender from being the solver
		if v.Replier == event.Sender {
			continue
		}

		// Decode replier state
		var state types.State
		utils.UnmarshalJSON(v.NodeState, &state)

		// Get and decode replier specs
		var specs types.NodeSpecs
		utils.UnmarshalJSON(managers.GetNodeSpecs(v.Replier), &specs)

		var met interface{}
		var comp types.RuleComparator

		// Select metric and comparator
		switch etype.Resource {
		case types.AllResources, types.CpuResource: // TODO: implement algorithm to check all resources at the same time
			met = state.CpuUsage / specs.CpuMhz // Ratio
			comp = types.LessComp
		case types.MemResource:
			met = specs.MemTotal - state.MemUsage // Free memory
			comp = types.GreaterComp
		case types.DiskResource:
			met = specs.DiskTotal - state.DiskUsage // Free storage space
			comp = types.GreaterComp
		case types.PktSentResource:
			met = state.NetPacketsSent // Sent packets
			comp = types.LessComp
		case types.PktRecvResource:
			met = state.NetPacketsRecv // Received packets
			comp = types.LessComp
		default:
			utils.CheckError(errUnknownSpec, utils.WarningMode)
			continue
		}

		// Set the best till now
		if best == nil {
			best, addr = met, v.Replier
			continue
		}

		if ok, err := utils.CompareValues(met, comp, best); ok {
			best, addr = met, v.Replier
		} else {
			utils.CheckError(err, utils.WarningMode)
		}
	}

	return
}*/

// Select a container according to its config and spec usage
func selectContainer(ctx context.Context, ccache map[uint64]bool) (uint64, error) {

	// Get DCR active containers
	for rcid := range managers.GetActiveContainers() {
		// Check if am I the host and if a previous event has already been sent for the container
		if managers.IsContainerHost(rcid, managers.GetFromAccount()) && !ccache[rcid] {
			/*cname := managers.GetContainerName(rcid)
			c := managers.SearchDockerContainers(ctx, "name", cname, true)
			if c != nil {
				// TODO: implement container selector (next container?)
				return rcid, nil
			}*/
			return rcid, nil
		}
	}

	return 0, errNoContainersFound
}

// Functions to select the spec metric type depending on the rule limit type
func selectCpuMetric(mt types.RuleMetricType, state *types.State) (usage interface{}) {

	switch mt {
	case types.PercentMetric:
		usage = state.CpuUsage // Usage %
	}

	return
}

func selectMemMetric(mt types.RuleMetricType, state *types.State) (usage interface{}) {

	specs := managers.GetSpecs()

	switch mt {
	case types.UnitsMetric:
		usage = state.MemUsage // Bytes
	case types.PercentMetric:
		usage = (float64(state.MemUsage) / float64(specs.MemTotal)) * 100.0
	}

	return
}

func selectDiskMetric(mt types.RuleMetricType, state *types.State) (usage interface{}) {

	specs := managers.GetSpecs()

	switch mt {
	case types.UnitsMetric:
		usage = state.DiskUsage // Bytes
	case types.PercentMetric:
		usage = (float64(state.DiskUsage) / float64(specs.DiskTotal)) * 100.0
	}

	return
}

func selectPktSentMetric(mt types.RuleMetricType, state *types.State) (usage interface{}) {

	switch mt {
	case types.UnitsMetric:
		usage = state.NetPacketsSent // Packet count
	}

	return
}

func selectPktRecvMetric(mt types.RuleMetricType, state *types.State) (usage interface{}) {

	switch mt {
	case types.UnitsMetric:
		usage = state.NetPacketsRecv // Packet count
	}

	return
}

// MonitorV2 //
func updateNodeReputations(nodeStore types.NodeStore, lossProbTh, latTh uint64) {

	// For each peer
	for nodeAddr, nodeInfo := range nodeStore {
		if nodeInfo.CurrentEpoch.TotalPackets == 0 {
			continue
		}

		// FILTER_1: availability
		filter1 := checkAvailabilityFilter(nodeInfo.CurrentEpoch, lossProbTh)

		// FILTER_2: latency
		filter2 := checkLatencyFilter(nodeInfo.CurrentEpoch, latTh)

		// Calculate reputation value
		var repValue uint8
		if filter1 && filter2 {
			repValue = 1
		}

		aggregateReputationValue(nodeStore, nodeAddr, repValue)

		// Reset current epoch
		nodeStore[nodeAddr].CurrentEpoch = types.EpochInfo{}
	}
}

func checkAvailabilityFilter(currentEpoch types.EpochInfo, lossProbTh uint64) (success bool) {

	if float64(currentEpoch.OKPackets)/float64(currentEpoch.TotalPackets) >= float64(100-lossProbTh)/100 {
		success = true
	}

	return
}

func checkLatencyFilter(currentEpoch types.EpochInfo, latTh uint64) (success bool) {

	count := len(currentEpoch.Latencies)
	if count > 0 {
		var total uint64
		for _, v := range currentEpoch.Latencies {
			total += v
		}
		if float64(total)/float64(count) <= float64(latTh) {
			success = true
		}
	}

	return
}

func aggregateReputationValue(nodeStore types.NodeStore, nodeAddr common.Address, repValue uint8) {

	// Save reputation value as historical value
	nodeStore[nodeAddr].Reputation.Values = append(nodeStore[nodeAddr].Reputation.Values, repValue)

	// TODO. Update reputation score
	rvs := nodeStore[nodeAddr].Reputation.Values
	var total uint64
	for _, v := range rvs {
		total += uint64(v)
	}
	nodeStore[nodeAddr].Reputation.Score = float64(total) / float64(len(rvs))
}

func selectSolver(eid uint64) common.Address {

	fmt.Println("\nREPLIES:")

	// Get reputation scores per node
	replies := managers.GetEventReplies(eid)
	scores := make(map[common.Address][]float64)
	for _, reply := range replies {

		fmt.Println(reply.Replier, reply.RepScores)

		for _, rs := range reply.RepScores {
			// Check reputation score
			score, err := strconv.ParseFloat(rs.Score, 64)
			if err != nil || score < 0 || score > 1 {
				utils.CheckError(err, utils.WarningMode)
				continue
			}

			// Store reputation score
			scores[rs.Node] = append(scores[rs.Node], score)
		}
	}

	fmt.Println("\nSCORES PER NODE:")

	// Aggregate reputation scores per node prioritizing the best
	maxrss := managers.GetClusterConfig().MaxRepScores
	totals := make(map[common.Address]float64)
	for naddr, nrss := range scores {
		// Sort node scores in descending order
		slices.Sort(nrss)
		slices.Reverse(nrss)

		fmt.Println(naddr, nrss)

		var count uint64
		for _, score := range nrss {
			// Limiting the number of reputation scores to be counted
			if count == maxrss {
				break
			}
			count++

			// Add reputation score to the total
			totals[naddr] += score
		}
	}

	fmt.Println("\nTOTAL SCORE PER NODE:")
	for k, v := range totals {
		fmt.Println(k, v)
	}

	// Get and decode container info
	rcid := managers.GetEvent(eid).Rcid
	var cinfo types.ContainerInfo
	if rcid > 0 {
		utils.UnmarshalJSON(managers.GetContainer(rcid).Info, &cinfo)
	}

	// Get the address of the most reputed node
	var (
		bestScore float64 = -1
		bestAddr  common.Address
	)
	for addr, total := range totals {
		// FILTER_3: resources
		if rcid > 0 && !managers.CanExecuteContainer(addr, cinfo.CpuLimit, cinfo.MemLimit) {
			continue
		}

		if total > bestScore {
			bestScore, bestAddr = total, addr
		} else if total == bestScore {
			// TODO: manage draws
			utils.CheckError(errReputationDraw, utils.WarningMode)
		}
	}

	return bestAddr
}
