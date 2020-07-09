package daemons

import (
	"context"
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

func StartMonitor(ctx context.Context) {

	// Goroutines to receive events
	go WatchNewEvent()
	go WatchRequiredReplies()
	go WatchRequiredVotes(ctx)
	go WatchEventSolved(ctx)
	go WatchNewContainer(ctx)
	go WatchContainerRemoved()

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

		/////////////
		// Testing //
		/////////////
		/*_ = managers.NewContainer(ctx)
		time.Sleep(5 * time.Second)

		fmt.Println(managers.GetRegActiveContainers())
		c := managers.GetRegContainer(19)
		fmt.Println(*c)

		managers.DeleteContainer(ctx, "registry_ctr_19")

		fmt.Println(managers.GetRegActiveContainers())
		c = managers.GetRegContainer(19)
		fmt.Println(*c)
		break*/

		// Check all state rules
		checkStateRules(ctx, cycles, mInter, cTime)
	}

	/////////////
	// Testing //
	/////////////
	/*cid := managers.NewContainer(ctx)
	fmt.Println(cid)
	cname := managers.SetContainerName(ctx, cid, 1)
	time.Sleep(5 * time.Second)

	state := managers.GetContainerState(ctx, cname)
	fmt.Println("STATE:", *state)

	time.Sleep(1 * time.Second)
	managers.DeleteContainer(ctx, cname)*/
}
