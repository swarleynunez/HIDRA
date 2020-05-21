package daemons

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"strconv"
)

func checkStateRules() {

}

func selectBestSolver(eid uint64) (addr common.Address) {

	// Get related event header
	header := getEvent(eid)

	// Decode dynamic event type
	var etype types.EventType
	utils.UnmarshalJSON(header.DynType, &etype)

	// Get event replies
	replies, err := cinst.GetEventReplies(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	// Temporal variable for different value types
	var best interface{}

	for _, v := range replies {

		// Get reply node specs
		specs := getNodeSpecs(v.Sender)

		// Decode reply node state
		var state types.NodeState
		utils.UnmarshalJSON(v.NodeState, &state)

		// Metrics by spec
		switch etype.Spec {
		case "cpu":
			cpuPercent, err := strconv.ParseFloat(state.CpuPercent, 64)
			utils.CheckError(err, utils.WarningMode)

			mhz, err := strconv.ParseFloat(specs.Mhz, 64)
			utils.CheckError(err, utils.WarningMode)

			// Metric (ratio)
			met := cpuPercent / mhz

			if best == nil {
				best, addr = met, v.Sender
				continue
			}

			if met < best.(float64) {
				best, addr = met, v.Sender
			}
		case "mem":
			// Metric (free memory)
			met := specs.MemTotal - state.MemUsage

			if best == nil {
				best, addr = met, v.Sender
				continue
			}

			if met > best.(uint64) {
				best, addr = met, v.Sender
			}
		case "disk":
			// Metric (free storage space)
			met := specs.DiskTotal - state.DiskUsage

			if best == nil {
				best, addr = met, v.Sender
				continue
			}

			if met > best.(uint64) {
				best, addr = met, v.Sender
			}
		case "pkt_sent":
			// Metric (sent packets)
			met := state.NetPacketsSent

			if best == nil {
				best, addr = met, v.Sender
				continue
			}

			if met < best.(uint64) {
				best, addr = met, v.Sender
			}
		case "pkt_recv":
			// Metric (received packets)
			met := state.NetPacketsRecv

			if best == nil {
				best, addr = met, v.Sender
				continue
			}

			if met < best.(uint64) {
				best, addr = met, v.Sender
			}
		}
	}

	return
}

func runEventTask(eid uint64) {

	fmt.Println("DEBUG: Running event task... DONE!")
}

func runEndingTasks(eid uint64) {

	fmt.Println("DEBUG: Running ending tasks... DONE!")
}
