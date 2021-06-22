package daemons

import (
	"context"
	"github.com/swarleynunez/superfog/core/managers"
	"github.com/swarleynunez/superfog/core/utils"
	"os"
	"strconv"
	"time"
)

// Hysteresis cycles
type cycleCounter struct {
	measures uint64
	triggers uint64
}

// Hysteresis cycles by rule name
type cycles map[string]cycleCounter

func Run(ctx context.Context) {

	// Node rule cycles
	cycles := cycles{}

	// Get and parse monitor time interval
	mInter, err := strconv.ParseUint(os.Getenv("MONITOR_INTERVAL"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	// Get and parse cycle time
	cTime, err := strconv.ParseUint(os.Getenv("CYCLE_TIME"), 10, 64)
	utils.CheckError(err, utils.FatalMode)

	// Cache to avoid sending duplicate events
	ecache := map[uint64]bool{}

	// Watchers to receive events
	go WatchNewEvent()
	go WatchRequiredReplies()
	go WatchRequiredVotes(ctx)
	go WatchEventSolved(ctx)
	go WatchNewContainer(ctx)
	go WatchContainerRemoved()

	// Recover host container state from distributed registry
	managers.InitContainerState(ctx)

	// Main loop
	for {
		time.Sleep(time.Duration(mInter) * time.Millisecond)

		// Check all state rules
		checkStateRules(ctx, cycles, mInter, cTime, ecache)
	}
}
