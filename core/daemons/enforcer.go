package daemons

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/superfog/core/managers"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"github.com/swarleynunez/superfog/inputs"
	"time"
)

const (
	blueInfoFormat = "\033[1;34m[%d] %s (Limit: %v, Usage: %v)\033[0m\n"
)

var (
	errUnknownSpec         = errors.New("unknown specification")
	errBoundNotImplemented = errors.New("spec bound type not implemented")
	errUnknownAction       = errors.New("unknown rule action")
	errNoContainersFound   = errors.New("no containers found")
)

func checkStateRules(ctx context.Context, rccs map[string]cycle, minter, ctime uint64, ccache map[uint64]bool) {

	state := managers.GetState()

	for _, rule := range inputs.Rules {
		// Current spec value (variable for different value types)
		var usage interface{}

		switch rule.Resource {
		case types.CpuResource:
			usage = selectCpuMetric(rule.MetricType, state)
		case types.MemResource:
			usage = selectMemMetric(rule.MetricType, state)
		case types.DiskResource:
			usage = selectDiskMetric(rule.MetricType, state)
		case types.PktSentResource:
			usage = selectPktSentMetric(rule.MetricType, state)
		case types.PktRecvResource:
			usage = selectPktRecvMetric(rule.MetricType, state)
		default:
			utils.CheckError(errUnknownSpec, utils.WarningMode)
			continue // Drop rule check
		}

		if usage == nil {
			utils.CheckError(errBoundNotImplemented, utils.WarningMode)
			continue // Drop rule check
		}

		// Get the rcc
		rcc := rccs[rule.NameId]

		// Count a measure if the rcc has already started
		if rcc.measures > 0 {
			rcc.measures++
		}

		// Rule checking
		if ok, err := utils.CompareValues(usage, rule.Comparator, rule.Limit); ok {

			// Start the rcc
			if rcc.measures == 0 {
				rcc.measures++
			}

			// Count a trigger
			rcc.triggers++

			if rcc.measures == ctime/minter && rcc.measures == rcc.triggers {
				runRuleAction(ctx, &rule, ccache, state, usage)
			}
		} else {
			utils.CheckError(err, utils.WarningMode)
		}

		// Reset the rcc
		if rcc.measures == ctime/minter {
			rcc = cycle{}
		}

		// Update the rcc for the next state checking
		rccs[rule.NameId] = rcc
	}
}

func runRuleAction(ctx context.Context, rule *types.Rule, ccache map[uint64]bool, state *types.State, usage interface{}) {

	switch rule.Action {
	case types.SendEventAction:
		rcid, err := selectContainer(ctx, ccache)
		if err == nil {
			ccache[rcid] = true

			// Encapsulate event type
			etype := types.EventType{
				RequiredTask:     types.MigrateContainerTask,
				TroubledResource: rule.Resource,
			}

			// Debug
			//fmt.Print("[", time.Now().UnixMilli(), "] ", "Sending an event...\n")

			go func() {
				err = managers.SendEvent(ctx, &etype, rcid, state)
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
func selectSolver(eid uint64) (addr common.Address) {

	event := managers.GetEvent(eid)
	replies := managers.GetEventReplies(eid)

	// Decode event type
	var etype types.EventType
	utils.UnmarshalJSON(event.EType, &etype)

	// Current best value (variable for different value types)
	var best interface{}

	for _, v := range replies {
		// Decode replier state
		var state types.State
		utils.UnmarshalJSON(v.NodeState, &state)

		// Get and decode replier specs
		var specs types.NodeSpecs
		utils.UnmarshalJSON(managers.GetNodeSpecs(v.Replier), &specs)

		var met interface{}
		var comp types.RuleComparator

		// Select metric and comparator
		switch etype.TroubledResource {
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

	// TODO. Testing
	addr1 := common.HexToAddress("0x24056A909B4Ed25ac47fbe6421b45cA0DeF1da8C")
	addr2 := common.HexToAddress("0xb066c34E2C26E6E03042Ae4AA11Dfb9A28cd7C52")
	//addr3 := common.HexToAddress("0xa852f9A4f20651e4D6645d5200B5CAef06AFf4fB")

	if event.Sender == addr1 {
		addr = addr2
	} else if event.Sender == addr2 {
		addr = addr1
	}

	return
}

// Select a container according to its config and spec usage
func selectContainer(ctx context.Context, ccache map[uint64]bool) (uint64, error) {

	// Get distributed registry active containers
	ctrs := managers.GetActiveContainers()
	for rcid := range ctrs {
		// Check if am I the host and if a previous event has already been sent for the container
		if managers.IsContainerHost(rcid, managers.GetFromAccount()) && !ccache[rcid] {
			cname := managers.GetContainerName(rcid)
			c := managers.SearchDockerContainers(ctx, "name", cname, true)
			if c != nil {
				// TODO: implement container selector (next container?)
				return rcid, nil
			}
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
