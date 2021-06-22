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

// DEL
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
			if !log.Raw.Removed && !ecache[log.EventId] {
				ecache[log.EventId] = true

				// Debug
				event := managers.GetEvent(log.EventId)
				fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "NewEvent (EID=", log.EventId, ", Sender=", event.Sender.String(), ", DynType=", event.DynType, ")\n")

				// Send a event reply containing the current node state
				go managers.SendReply(log.EventId, managers.GetNodeState())
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

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
			if !log.Raw.Removed && !ecache[log.EventId] {
				ecache[log.EventId] = true

				// Debug
				fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "RequiredReplies (EID=", log.EventId, ")\n")

				// Select and vote the best event solver
				solver := selectBestSolver(log.EventId, cinst)
				if !utils.EmptyEthAddress(solver.String()) {
					go managers.VoteSolver(log.EventId, solver)
				}
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

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
			if !log.Raw.Removed && !ecache[log.EventId] {
				ecache[log.EventId] = true

				// Debug
				fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "RequiredVotes (EID=", log.EventId, ", Solver=", log.Solver.String(), ")\n")

				// Node account
				from := managers.GetFromAccount()

				// Am I the voted solver?
				if log.Solver == from.Address {
					// Get related event header
					event := managers.GetEvent(log.EventId)

					// TODO. Am I the event sender?
					if event.Sender != from.Address {
						// Decode dynamic event type
						var etype types.EventType
						utils.UnmarshalJSON(event.DynType, &etype)

						// Execute required task (from dynamic event type)
						go managers.RunEventTask(ctx, etype, log.EventId)
					} else {
						// Solve related event
						go managers.SolveEvent(log.EventId)
					}
				}
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

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
			if !log.Raw.Removed && !ecache[log.EventId] {
				ecache[log.EventId] = true

				// Debug
				fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "EventSolved (EID=", log.EventId, ")\n")

				// Get related event header
				event := managers.GetEvent(log.EventId)

				// Node account
				from := managers.GetFromAccount()

				// TODO. Am I the event sender and the event solver?
				if event.Sender == from.Address && event.Solver != from.Address {
					// Decode dynamic event type
					var etype types.EventType
					utils.UnmarshalJSON(event.DynType, &etype)

					// Any completion tasks?
					go managers.RunEventEndingTask(ctx, etype)
				}
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

// DCR
func WatchNewContainer(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerNewContainer)

	// Subscription to the event
	sub, err := cinst.WatchNewContainer(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicate event logs
	ecache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Check if an event log has already been received
			if !log.Raw.Removed && !ecache[log.RegistryCtrId] {
				ecache[log.RegistryCtrId] = true

				// Debug
				fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "NewContainer (RCID=", log.RegistryCtrId, ", CID=", log.CtrId, ", Host=", log.Host.String(), ")\n")

				// Node account
				from := managers.GetFromAccount()

				// Am I the container host?
				if log.Host == from.Address {
					// Set registry container ID to the local container
					go managers.SetContainerName(ctx, log.CtrId, managers.GetContainerName(log.RegistryCtrId))
				}
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

func WatchContainerRemoved() {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerContainerRemoved)

	// Subscription to the event
	sub, err := cinst.WatchContainerRemoved(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Cache to avoid duplicate event logs
	ecache := map[uint64]bool{}

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Check if an event log has already been received
			if !log.Raw.Removed && !ecache[log.RegistryCtrId] {
				ecache[log.RegistryCtrId] = true

				// Debug
				fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "ContainerRemoved (RCID=", log.RegistryCtrId, ")\n")
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}
