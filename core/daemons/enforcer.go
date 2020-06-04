package daemons

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/superfog/core/policies"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
)

const (
	yellowFormat = "\033[1;33m[+] %s (Bound: %v, Now: %v)\033[0m\n"
)

var (
	errUnknownSpec         = errors.New("unknown specification")
	errBoundNotImplemented = errors.New("spec bound type not implemented")
	errUnknownAction       = errors.New("unknown rule action")
)

func checkStateRules(cycles cycles, mInter, cTime uint64) {

	state := getState()

	for _, r := range policies.Rules {

		// Current spec value (variable for different value types)
		var now interface{}

		switch r.Spec {
		case types.CpuSpec:
			now = selectCpuMetric(r.MetricType, state)
		case types.MemSpec:
			now = selectMemMetric(r.MetricType, state)
		case types.DiskSpec:
			now = selectDiskMetric(r.MetricType, state)
		case types.PktSentSpec:
			now = selectPktSentMetric(r.MetricType, state)
		case types.PktRecvSpec:
			now = selectPktRecvMetric(r.MetricType, state)
		default:
			utils.CheckError(errUnknownSpec, utils.WarningMode)
			continue // Drop rule check
		}

		if now == nil {
			utils.CheckError(errBoundNotImplemented, utils.WarningMode)
			continue // Drop rule check
		}

		// Get rule cycle counter (rcc)
		cc := cycles[r.NameId]

		// Count a measure if the rcc has already started
		if cc.measures > 0 {
			cc.measures++
		}

		// Rule checking
		if ok, err := utils.CompareValues(now, r.Comparator, r.Bound); ok {

			// Start a rcc
			if cc.measures == 0 {
				cc.measures++
			}

			// Count a trigger
			cc.triggers++

			// TODO. Rcc checking
			if cc.measures == cTime/mInter && cc.measures == cc.triggers {
				runRuleAction(&r, state, now)
			}
		} else {
			utils.CheckError(err, utils.WarningMode)
		}

		// Reset rcc
		if cc.measures == cTime/mInter {
			cc = cycleCounter{}
		}

		// Update rcc for the next state
		cycles[r.NameId] = cc
	}
}

func runRuleAction(rule *types.Rule, state *types.NodeState, now interface{}) {

	switch rule.Action {
	case types.SendEventAction:
		etype := types.EventType{
			Spec: rule.Spec,
			Task: types.CreateTask,
			Metadata: map[string]interface{}{
				"docker_image": 12345678,
			},
		}
		sendEvent(&etype, state)
	case types.ProceedAction:
		runTask(types.CreateTask)
		fallthrough
	case types.LogAction:
		// Save log into a file, send log to a remote server...
		fallthrough
	case types.WarnAction:
		fmt.Printf(yellowFormat, rule.Msg, rule.Bound, now)
	case types.IgnoreAction:
		// Do nothing
	default:
		utils.CheckError(errUnknownAction, utils.WarningMode)
		return
	}
}

// Functions to run node management tasks
func runTask(task types.Task) {

	fmt.Print("DEBUG: Running task\n")
}

func runEventTask(eid uint64, task types.Task) {

	fmt.Print("DEBUG: Running event task (EID=", eid, ")\n")
}

func runEventEndingTask(eid uint64, task types.Task) {

	fmt.Print("DEBUG: Running event ending task (EID=", eid, ")\n")
}

// Select the best event solver according to spec metrics
func selectBestSolver(eid uint64) (addr common.Address) {

	// Get related event header
	event := getEvent(eid)

	// Decode dynamic event type
	var etype types.EventType
	utils.UnmarshalJSON(event.DynType, &etype)

	// Get event replies
	replies, err := _cinst.GetEventReplies(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	// Current best value (variable for different value types)
	var best interface{}

	for _, v := range replies {

		// Get reply node specs
		ns := getNodeSpecs(v.Sender)

		// Decode reply node state
		var state types.NodeState
		utils.UnmarshalJSON(v.NodeState, &state)

		var met interface{}
		var comp types.Comparator

		// Select metric and comparator
		switch etype.Spec {
		case types.CpuSpec:
			met = state.CpuPercent / ns.Mhz // Ratio
			comp = types.LessComp
		case types.MemSpec:
			met = ns.MemTotal - state.MemUsage // Free memory
			comp = types.GreaterComp
		case types.DiskSpec:
			met = ns.DiskTotal - state.DiskUsage // Free storage space
			comp = types.GreaterComp
		case types.PktSentSpec:
			met = state.NetPacketsSent // Sent packets
			comp = types.LessComp
		case types.PktRecvSpec:
			met = state.NetPacketsRecv // Received packets
			comp = types.LessComp
		default:
			utils.CheckError(errUnknownSpec, utils.WarningMode)
			continue
		}

		// Set the best till now
		if best == nil {
			best, addr = met, v.Sender
			continue
		}

		if ok, err := utils.CompareValues(met, comp, best); ok {
			best, addr = met, v.Sender
		} else {
			utils.CheckError(err, utils.WarningMode)
		}
	}

	return
}

// Functions to select the spec metric type depending on the rule bound type
func selectCpuMetric(mt types.MetricType, state *types.NodeState) (now interface{}) {

	switch mt {
	case types.PercentType:
		now = state.CpuPercent // Usage %
	}

	return
}

func selectMemMetric(mt types.MetricType, state *types.NodeState) (now interface{}) {

	// Get node specs
	specs := getSpecs()

	switch mt {
	case types.UnitsType:
		now = state.MemUsage // Bytes
	case types.PercentType:
		now = (float64(state.MemUsage) / float64(specs.MemTotal)) * 100.0
	}

	return
}

func selectDiskMetric(mt types.MetricType, state *types.NodeState) (now interface{}) {

	// Get node specs
	specs := getSpecs()

	switch mt {
	case types.UnitsType:
		now = state.DiskUsage // Bytes
	case types.PercentType:
		now = (float64(state.DiskUsage) / float64(specs.DiskTotal)) * 100.0
	}

	return
}

func selectPktSentMetric(mt types.MetricType, state *types.NodeState) (now interface{}) {

	switch mt {
	case types.UnitsType:
		now = state.NetPacketsSent // Packet count
	}

	return
}

func selectPktRecvMetric(mt types.MetricType, state *types.NodeState) (now interface{}) {

	switch mt {
	case types.UnitsType:
		now = state.NetPacketsRecv // Packet count
	}

	return
}
