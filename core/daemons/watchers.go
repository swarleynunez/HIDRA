package daemons

import (
	"context"
	"fmt"
	"github.com/swarleynunez/superfog/core/bindings"
	"github.com/swarleynunez/superfog/core/managers"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"time"
)

// DEL (debug: all cluster nodes)
func WatchNewEvent() {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerNewEvent)

	// Subscription to the event
	sub, err := cinst.WatchNewEvent(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicate event logs
	ecache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if an event log has already been received
			if !log.Raw.Removed && !ecache[log.Eid] {
				ecache[log.Eid] = true

				// Debug
				event := managers.GetEvent(log.Eid)
				if event.Rcid > 0 {
					fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "NewEvent (EID=", log.Eid, ", Sender=", event.Sender.String(), ", RCID=", event.Rcid, ")\n")

				} else {
					fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "NewEvent (EID=", log.Eid, ", Sender=", event.Sender.String(), ")\n")
				}

				// Send a event reply containing the current node state
				go managers.SendReply(log.Eid, managers.GetState())
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DEL (debug: all cluster nodes)
func WatchRequiredReplies() {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerRequiredReplies)

	// Subscription to the event
	sub, err := cinst.WatchRequiredReplies(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicate event logs
	ecache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if an event log has already been received
			if !log.Raw.Removed && !ecache[log.Eid] {
				ecache[log.Eid] = true

				// Debug
				fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "RequiredReplies (EID=", log.Eid, ")\n")

				// Select and vote an event solver
				solver := selectSolver(log.Eid)
				if !utils.EmptyEthAddress(solver.String()) {
					go managers.VoteSolver(log.Eid, solver)
				}
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DEL (debug: only solver nodes)
func WatchRequiredVotes(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerRequiredVotes)

	// Subscription to the event
	sub, err := cinst.WatchRequiredVotes(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicate event logs
	ecache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if an event log has already been received
			if !log.Raw.Removed && !ecache[log.Eid] {
				ecache[log.Eid] = true

				// Am I the voted solver?
				event := managers.GetEvent(log.Eid)
				from := managers.GetFromAccount()
				if event.Solver == from {
					// Debug
					fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "RequiredVotes (EID=", log.Eid, ")\n")

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
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DEL (debug only sender nodes)
func WatchEventSolved(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerEventSolved)

	// Subscription to the event
	sub, err := cinst.WatchEventSolved(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicate event logs
	ecache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if an event log has already been received
			if !log.Raw.Removed && !ecache[log.Eid] {
				ecache[log.Eid] = true

				// Am I the event sender and not the event solver?
				event := managers.GetEvent(log.Eid)
				from := managers.GetFromAccount()
				if event.Sender == from {
					// Debug
					fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "EventSolved (EID=", log.Eid, ")\n")

					if event.Solver != from {
						// Execute required ending task (depends on the event type)
						go managers.RunEventEndingTask(ctx, event)
					}
				}
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DCR (debug only owner nodes)
func WatchContainerRegistered(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerContainerRegistered)

	// Subscription to the event
	sub, err := cinst.WatchContainerRegistered(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicate event logs
	ecache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if an event log has already been received
			if !log.Raw.Removed && !ecache[log.Rcid] {
				ecache[log.Rcid] = true

				// Am I the container owner?
				ctr := managers.GetContainer(log.Rcid)
				owner := managers.GetApplication(ctr.Appid).Owner
				from := managers.GetFromAccount()
				if owner == from {
					// Debug
					fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "ContainerRegistered (RCID=", log.Rcid, ", APPID=", ctr.Appid, ")\n")

					// Am I the container host?
					if managers.IsContainerHost(log.Rcid, from) {
						// Decode container info
						var cinfo types.ContainerInfo
						utils.UnmarshalJSON(ctr.Info, &cinfo)

						// Autodeploy mode (anonymous function)
						go func() {
							managers.NewContainer(ctx, &cinfo, managers.GetContainerName(log.Rcid))
							managers.ActivateContainer(log.Rcid)
						}()
					} else {
						// Encapsulate event type
						etype := types.EventType{
							RequiredTask:     types.NewContainerTask,
							TroubledResource: types.AllResources,
						}

						go managers.SendEvent(&etype, log.Rcid, managers.GetState())
					}
				}
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DCR (debug only host nodes)
func WatchContainerUpdated(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerContainerUpdated)

	// Subscription to the event
	sub, err := cinst.WatchContainerUpdated(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicate event logs
	ecache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if an event log has already been received
			if !log.Raw.Removed && !ecache[log.Rcid] {
				ecache[log.Rcid] = true

				// Am I the container host?
				if managers.IsContainerHost(log.Rcid, managers.GetFromAccount()) {
					// Debug
					fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "ContainerUpdated (RCID=", log.Rcid, ")\n")

					// Get and decode container info
					ctr := managers.GetContainer(log.Rcid)
					var cinfo types.ContainerInfo
					utils.UnmarshalJSON(ctr.Info, &cinfo)

					// TODO: think about which containers fields could be updated by the owner and how
					go func() { // Anonymous function
						managers.RemoveContainer(ctx, managers.GetContainerName(log.Rcid))
						managers.NewContainer(ctx, &cinfo, managers.GetContainerName(log.Rcid))
					}()
				} else {
					// Clean container old instances (if exists)
					go managers.RemoveContainer(ctx, managers.GetContainerName(log.Rcid))
				}
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DCR (debug only host nodes)
func WatchContainerUnregistered(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerContainerUnregistered)

	// Subscription to the event
	sub, err := cinst.WatchContainerUnregistered(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicate event logs
	ecache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:
			// Check if an event log has already been received
			if !log.Raw.Removed && !ecache[log.Rcid] {
				ecache[log.Rcid] = true

				// Am I the container host?
				if managers.IsContainerHost(log.Rcid, managers.GetFromAccount()) {
					// Debug
					fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "ContainerUnregistered (RCID=", log.Rcid, ")\n")
				}

				// Clean all container instances (in execution or old instances)
				go managers.RemoveContainer(ctx, managers.GetContainerName(log.Rcid))
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}
