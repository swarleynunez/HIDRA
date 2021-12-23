package daemons

import (
	"context"
	"github.com/swarleynunez/superfog/core/utils"
	"strconv"
	"time"
)

// Rule cycle counter (rcc)
type cycle struct {
	measures uint64
	triggers uint64
}

func Run(ctx context.Context) {

	// Rule cycle counter (rcc) per rule
	rccs := map[string]cycle{}

	// Get and parse monitor time interval
	minter, err := strconv.ParseUint(utils.GetEnv("MONITOR_INTERVAL"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	// Get and parse cycle time
	ctime, err := strconv.ParseUint(utils.GetEnv("CYCLE_TIME"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	// Cache to avoid duplicated events per container
	ccache := map[uint64]bool{}

	// Watchers to receive events
	go WatchNewEvent(ctx)
	go WatchRequiredReplies(ctx)
	go WatchRequiredVotes(ctx, ccache)
	go WatchEventSolved(ctx, ccache)
	go WatchApplicationRegistered()
	go WatchContainerRegistered(ctx)
	go WatchContainerUpdated(ctx)
	go WatchContainerUnregistered(ctx)

	// TODO: check node/Docker running ports (also check registered ports in DCR)
	// Recover node state from distributed registry
	//managers.InitNodeState(ctx)

	// Main loop
	for {
		time.Sleep(time.Duration(minter) * time.Millisecond)

		// Check all state rules
		checkStateRules(ctx, rccs, minter, ctime, ccache)
	}
}
