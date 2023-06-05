package daemons

import (
	"context"
	"fmt"
	"github.com/swarleynunez/hidra/core/bindings"
	"github.com/swarleynunez/hidra/core/managers"
	"github.com/swarleynunez/hidra/core/types"
	"github.com/swarleynunez/hidra/core/utils"
	"time"
)

// DEL (debug: all cluster nodes)
func WatchNewEvent(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Log channel
	logs := make(chan *bindings.ControllerNewEvent)

	// Subscription to the event
	sub, err := cinst.WatchNewEvent(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicated logs
	lcache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if a log has already been received
			if !log.Raw.Removed && !lcache[log.Eid] {
				lcache[log.Eid] = true

				// Debug
				event := managers.GetEvent(log.Eid)
				if event.Rcid > 0 {
					fmt.Print("[", time.Now().UnixNano(), "] ", "NewEvent (EID=", log.Eid, ", Sender=", event.Sender.String(), ", RCID=", event.Rcid, ")\n")
				} else {
					fmt.Print("[", time.Now().UnixNano(), "] ", "NewEvent (EID=", log.Eid, ", Sender=", event.Sender.String(), ")\n")
				}

				// Am I the event sender?
				from := managers.GetFromAccount()
				if event.Sender != from {
					// Send an event reply containing the current node state
					go func() {
						err = managers.SendReply(ctx, log.Eid, managers.GetState())
						utils.CheckError(err, utils.WarningMode)
					}()
				}
			}
		case err = <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DEL (debug: all cluster nodes)
func WatchRequiredReplies(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	logs := make(chan *bindings.ControllerRequiredReplies)

	// Subscription to the event
	sub, err := cinst.WatchRequiredReplies(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicated logs
	lcache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if a log has already been received
			if !log.Raw.Removed && !lcache[log.Eid] {
				lcache[log.Eid] = true

				// Debug
				fmt.Print("[", time.Now().UnixNano(), "] ", "RequiredReplies (EID=", log.Eid, ")\n")

				// Select and vote an event solver
				solver := selectSolver(log.Eid)
				if !utils.EmptyEthAddress(solver.String()) {
					go func() {
						err = managers.VoteSolver(ctx, log.Eid, solver)
						utils.CheckError(err, utils.WarningMode)
					}()
				}
			}
		case err = <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DEL (debug: all cluster nodes)
func WatchRequiredVotes(ctx context.Context, ccache map[uint64]bool) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	logs := make(chan *bindings.ControllerRequiredVotes)

	// Subscription to the event
	sub, err := cinst.WatchRequiredVotes(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicated logs
	lcache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if a log has already been received
			if !log.Raw.Removed && !lcache[log.Eid] {
				lcache[log.Eid] = true

				// Debug
				event := managers.GetEvent(log.Eid)
				fmt.Print("[", time.Now().UnixNano(), "] ", "RequiredVotes (EID=", log.Eid, ", Solver=", event.Solver.String(), ")\n")

				// Am I the voted solver?
				from := managers.GetFromAccount()
				if event.Solver == from {
					// Update container cache
					if event.Rcid > 0 {
						ccache[event.Rcid] = true
					}

					// Am I the event sender?
					if event.Sender != from {
						// Execute required event task (depends on the event type)
						go managers.RunEventTask(ctx, event, log.Eid)
					} else {
						// Execute required local task (depends on the event type)
						go managers.RunTask(ctx, event, log.Eid)
					}
				}
			}
		case err = <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DEL (debug: all cluster nodes)
func WatchEventSolved(ctx context.Context, ccache map[uint64]bool) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	logs := make(chan *bindings.ControllerEventSolved)

	// Subscription to the event
	sub, err := cinst.WatchEventSolved(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicated logs
	lcache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if a log has already been received
			if !log.Raw.Removed && !lcache[log.Eid] {
				lcache[log.Eid] = true

				// Debug
				fmt.Print("[", time.Now().UnixNano(), "] ", "EventSolved (EID=", log.Eid, ")\n")

				// Am I the event sender and not the event solver?
				event := managers.GetEvent(log.Eid)
				from := managers.GetFromAccount()
				if event.Sender == from {
					if event.Solver != from {
						// Execute required ending task (depends on the event type)
						go managers.RunEventEndingTask(ctx, event)
					}
				}

				// Update container cache
				if event.Sender == from || event.Solver == from {
					if event.Rcid > 0 {
						ccache[event.Rcid] = false
					}
				}
			}
		case err = <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DCR (debug: only owner nodes)
func WatchApplicationRegistered() {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	logs := make(chan *bindings.ControllerApplicationRegistered)

	// Subscription to the event
	sub, err := cinst.WatchApplicationRegistered(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicated logs
	lcache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if a log has already been received
			if !log.Raw.Removed && !lcache[log.Appid] {
				lcache[log.Appid] = true

				// Am I the application owner?
				app := managers.GetApplication(log.Appid)
				from := managers.GetFromAccount()
				if app.Owner == from {
					// Debug
					fmt.Print("[", time.Now().UnixNano(), "] ", "ApplicationRegistered (APPID=", log.Appid, ")\n")

					// Decode application info
					var ainfo types.ApplicationInfo
					utils.UnmarshalJSON(app.Info, &ainfo)

					// ONOS SDN plugin
					managers.ONOSAddVirtualService(log.Appid, ainfo.Description, ainfo.IP, ainfo.Protocol, ainfo.Port)
				}
			}
		case err = <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DCR (debug: only owner nodes)
func WatchContainerRegistered(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	logs := make(chan *bindings.ControllerContainerRegistered)

	// Subscription to the event
	sub, err := cinst.WatchContainerRegistered(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicated logs
	lcache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if a log has already been received
			if !log.Raw.Removed && !lcache[log.Rcid] {
				lcache[log.Rcid] = true

				// Am I the container owner?
				ctr := managers.GetContainer(log.Rcid)
				owner := managers.GetApplication(ctr.Appid).Owner
				from := managers.GetFromAccount()
				if owner == from {
					// Debug
					fmt.Print("[", time.Now().UnixNano(), "] ", "ContainerRegistered (RCID=", log.Rcid, ", APPID=", ctr.Appid, ")\n")

					// Am I the container host?
					if managers.IsContainerHost(log.Rcid, from) {
						// Decode container info
						var cinfo types.ContainerInfo
						utils.UnmarshalJSON(ctr.Info, &cinfo)

						// Autodeploy mode (anonymous function)
						go func() {
							// managers.NewContainer(ctx, &cinfo, ctr.Appid, log.Rcid, true)
							err = managers.ActivateContainer(ctx, log.Rcid)
							utils.CheckError(err, utils.WarningMode)
						}()
					} else {
						// Encapsulate event type
						etype := types.EventType{
							RequiredTask:     types.NewContainerTask,
							TroubledResource: types.AllResources,
						}

						go func() {
							err = managers.SendEvent(ctx, &etype, log.Rcid, managers.GetState())
							utils.CheckError(err, utils.WarningMode)
						}()
					}
				}
			}
		case err = <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DCR (debug: only host nodes)
func WatchContainerUpdated(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	logs := make(chan *bindings.ControllerContainerUpdated)

	// Subscription to the event
	sub, err := cinst.WatchContainerUpdated(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// TODO: Check if an event log has already been received (repeated rcids)
			if !log.Raw.Removed {
				// Am I the container host?
				ctr := managers.GetContainer(log.Rcid)
				if managers.IsContainerHost(log.Rcid, managers.GetFromAccount()) {
					// Debug
					fmt.Print("[", time.Now().UnixNano(), "] ", "ContainerUpdated (RCID=", log.Rcid, ")\n")

					// Decode container info
					var cinfo types.ContainerInfo
					utils.UnmarshalJSON(ctr.Info, &cinfo)

					// TODO: think about which containers fields could be updated by the owner and how
					go func() { // Anonymous function
						// TODO: manage container ports by cluster node
						// managers.RemoveContainer(ctx, ctr.Appid, log.Rcid, true)
						// managers.NewContainer(ctx, &cinfo, ctr.Appid, log.Rcid, true)
					}()
				} else {
					// Clean container old instances (if exists)
					// go managers.RemoveContainer(ctx, ctr.Appid, log.Rcid, false)
				}
			}
		case err = <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DCR (debug: only host nodes)
func WatchContainerUnregistered(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	logs := make(chan *bindings.ControllerContainerUnregistered)

	// Subscription to the event
	sub, err := cinst.WatchContainerUnregistered(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicated logs
	lcache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if a log has already been received
			if !log.Raw.Removed && !lcache[log.Rcid] {
				lcache[log.Rcid] = true

				// Am I the container host?
				_ = managers.GetContainer(log.Rcid)
				if managers.IsContainerHost(log.Rcid, managers.GetFromAccount()) {
					// Debug
					fmt.Print("[", time.Now().UnixNano(), "] ", "ContainerUnregistered (RCID=", log.Rcid, ")\n")

					// Check if it is unregistering an application and remove container
					// go managers.RemoveContainer(ctx, ctr.Appid, log.Rcid, !managers.IsApplicationUnregistered(ctr.Appid))
				} else {
					// Clean container old instances (if exists)
					// go managers.RemoveContainer(ctx, ctr.Appid, log.Rcid, false)
				}
			}
		case err = <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}
