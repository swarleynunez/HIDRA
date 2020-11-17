package daemons

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/superfog/core/bindings"
	"github.com/swarleynunez/superfog/core/managers"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"github.com/swarleynunez/superfog/inputs"
	"time"
)

const (
	yellowWarnFormat = "\033[1;33m[%s] %s (Bound: %v, Now: %v)\033[0m\n"
)

var (
	errUnknownSpec         = errors.New("unknown specification")
	errBoundNotImplemented = errors.New("spec bound type not implemented")
	errUnknownAction       = errors.New("unknown rule action")
	errNoContainersFound   = errors.New("no containers found")
)

func checkStateRules(ctx context.Context, cycles cycles, mInter, cTime uint64) {

	state := managers.GetNodeState()

	for _, v := range inputs.Rules {

		// Current spec value (variable for different value types)
		var now interface{}

		switch v.Spec {
		case types.CpuSpec:
			now = selectCpuMetric(v.MetricType, state)
		case types.MemSpec:
			now = selectMemMetric(v.MetricType, state)
		case types.DiskSpec:
			now = selectDiskMetric(v.MetricType, state)
		case types.PktSentSpec:
			now = selectPktSentMetric(v.MetricType, state)
		case types.PktRecvSpec:
			now = selectPktRecvMetric(v.MetricType, state)
		default:
			utils.CheckError(errUnknownSpec, utils.WarningMode)
			continue // Drop rule check
		}

		if now == nil {
			utils.CheckError(errBoundNotImplemented, utils.WarningMode)
			continue // Drop rule check
		}

		// Get rule cycle counter (rcc)
		cc := cycles[v.NameId]

		// Count a measure if the rcc has already started
		if cc.measures > 0 {
			cc.measures++
		}

		// Rule checking
		if ok, err := utils.CompareValues(now, v.Comparator, v.Bound); ok {

			// Start a rcc
			if cc.measures == 0 {
				cc.measures++
			}

			// Count a trigger
			cc.triggers++

			// TODO. Rcc checking
			if cc.measures == cTime/mInter && cc.measures == cc.triggers {
				runRuleAction(ctx, &v, state, now)
			}
		} else {
			utils.CheckError(err, utils.WarningMode)
		}

		// Reset rcc
		if cc.measures == cTime/mInter {
			cc = cycleCounter{}
		}

		// Update rcc for the next state
		cycles[v.NameId] = cc
	}
}

func runRuleAction(ctx context.Context, rule *types.Rule, state *types.State, now interface{}) {

	switch rule.Action {
	case types.SendEventAction:
		// TODO
		rcid, err := selectWorseContainer(ctx)
		if err == nil {
			etype := types.EventType{
				Spec: rule.Spec,
				Task: types.MigrateContainerTask, // TODO
				Metadata: map[string]interface{}{
					"rcid": rcid,
				},
			}
			managers.SendEvent(&etype, state)
		}

		fallthrough
	case types.ProceedAction:
		if rule.Action == types.ProceedAction { // Due to the fallthrough
			//managers.RunTask(ctx, types.CreateTask)	// TODO
		}
		fallthrough
	case types.LogAction:
		// Save log into a file, send log to a remote server...
		fallthrough
	case types.WarnAction:
		fmt.Printf(yellowWarnFormat, time.Now().Format("15:04:05.000000"), rule.Msg, rule.Bound, now)
	case types.IgnoreAction:
		// Do nothing
	default:
		utils.CheckError(errUnknownAction, utils.WarningMode)
		return
	}
}

// Select the best event solver according to spec metrics
func selectBestSolver(eid uint64, cinst *bindings.Controller) (addr common.Address) {

	// Get related event header
	event := managers.GetEvent(eid)

	// Decode dynamic event type
	var etype types.EventType
	utils.UnmarshalJSON(event.DynType, &etype)

	// Get event replies
	replies, err := cinst.GetEventReplies(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	// Current best value (variable for different value types)
	var best interface{}

	for _, v := range replies {

		// Get replier specs
		ns := managers.GetSpecs(v.Sender)

		// Decode reply node state
		var state types.State
		utils.UnmarshalJSON(v.NodeState, &state)

		var met interface{}
		var comp types.Comparator

		// Select metric and comparator
		switch etype.Spec {
		case types.CpuSpec:
			met = state.CpuPercent / ns.CpuMhz // Ratio
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

// TODO. Select the worse container according to its config and spec usage
func selectWorseContainer(ctx context.Context) (uint64, error) {

	// Get distributed registry active containers
	ctrs := managers.GetContainerReg()
	for key := range ctrs {
		// Node account
		from := managers.GetFromAccount()

		// Am I the host?
		if ctrs[key].Host == from.Address {
			// Is the container running?
			ctr := managers.SearchDockerContainers(ctx, "name", managers.GetContainerName(key), false)
			if ctr != nil {
				// Decode container info
				var cinfo types.ContainerInfo
				utils.UnmarshalJSON(ctrs[key].Info, &cinfo)

				// TODO
				//if cinfo.MainSpec == spec {
				return key, nil
				//}
			}
		}
	}

	return 0, errNoContainersFound
}

// Functions to select the spec metric type depending on the rule bound type
func selectCpuMetric(mt types.MetricType, state *types.State) (now interface{}) {

	switch mt {
	case types.PercentMetric:
		now = state.CpuPercent // Usage %
	}

	return
}

func selectMemMetric(mt types.MetricType, state *types.State) (now interface{}) {

	specs := managers.GetNodeSpecs()

	switch mt {
	case types.UnitsMetric:
		now = state.MemUsage // Bytes
	case types.PercentMetric:
		now = (float64(state.MemUsage) / float64(specs.MemTotal)) * 100.0
	}

	return
}

func selectDiskMetric(mt types.MetricType, state *types.State) (now interface{}) {

	specs := managers.GetNodeSpecs()

	switch mt {
	case types.UnitsMetric:
		now = state.DiskUsage // Bytes
	case types.PercentMetric:
		now = (float64(state.DiskUsage) / float64(specs.DiskTotal)) * 100.0
	}

	return
}

func selectPktSentMetric(mt types.MetricType, state *types.State) (now interface{}) {

	switch mt {
	case types.UnitsMetric:
		now = state.NetPacketsSent // Packet count
	}

	return
}

func selectPktRecvMetric(mt types.MetricType, state *types.State) (now interface{}) {

	switch mt {
	case types.UnitsMetric:
		now = state.NetPacketsRecv // Packet count
	}

	return
}
