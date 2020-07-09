package daemons

import (
	"context"
	"fmt"
	"github.com/swarleynunez/superfog/core/bindings"
	"github.com/swarleynunez/superfog/core/managers"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
)

func WatchNewEvent() {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerNewEvent)

	// Subscription to the event
	sub, err := cinst.WatchNewEvent(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: NewEvent (EID=", log.EventId, ")\n")

			// Send a event reply containing the current node state
			managers.SendReply(log.EventId, managers.GetNodeState())
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

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: RequiredReplies (EID=", log.EventId, ")\n")

			// Select and vote the best event solver
			solver := selectBestSolver(log.EventId, cinst)
			if !utils.EmptyEthAddress(solver.String()) {
				managers.VoteSolver(log.EventId, solver)
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

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: RequiredVotes (EID=", log.EventId, ", Solver=", log.Solver.String(), ")\n")

			// Node account
			from := managers.GetFromAccount()

			// I am the voted solver?
			if log.Solver == from.Address {
				// Execute required task (from dynamic event type)
				managers.RunEventTask(ctx, log.EventId, types.CreateTask)

				// Solve related event
				managers.SolveEvent(log.EventId)
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

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: EventSolved (EID=", log.EventId, ", Sender=", log.Sender.String(), ")\n")

			// Node account
			from := managers.GetFromAccount()

			// I am the event sender?
			if log.Sender == from.Address {
				// Any completion tasks?
				managers.RunEventEndingTask(ctx, log.EventId, types.CreateTask)
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

func WatchNewContainer(ctx context.Context) {

	// Controller smart contract instance
	cinst := managers.GetControllerInst()

	// Event/log channel
	logs := make(chan *bindings.ControllerNewContainer)

	// Subscription to the event
	sub, err := cinst.WatchNewContainer(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: NewContainer (RCID=", log.RegistryCtrId, ", CID=", log.CtrId, ", Host=", log.Host.String(), ")\n")

			// Node account
			from := managers.GetFromAccount()

			// I am the container host?
			if log.Host == from.Address {
				// Set registry container ID to the local container
				_ = managers.SetContainerName(ctx, log.CtrId, log.RegistryCtrId)
			}

			// TODO. Refresh local registry
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

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: ContainerRemoved (RCID=", log.RegistryCtrId, ")\n")

			// TODO. Refresh local registry
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}
