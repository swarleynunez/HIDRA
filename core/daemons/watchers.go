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

			// TODO. Debug
			event := managers.GetEvent(log.EventId)
			fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "DEBUG: NewEvent (EID=", log.EventId, ", Sender=", event.Sender.String(), ", DynType=", event.DynType, ")\n")

			// Send a event reply containing the current node state
			managers.SendReply(log.EventId, managers.GetNodeState())

			//////////
			// TEST //
			/*time.Sleep(5 * time.Second)
			replies, err := cinst.GetEventReplies(nil, log.EventId)
			utils.CheckError(err, utils.WarningMode)
			fmt.Println(replies)*/
			//////////

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
			fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "DEBUG: RequiredReplies (EID=", log.EventId, ")\n")

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
			fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "DEBUG: RequiredVotes (EID=", log.EventId, ", Solver=", log.Solver.String(), ")\n")

			// Node account
			from := managers.GetFromAccount()

			// Am I the voted solver?
			if log.Solver == from.Address {
				// Get related event header
				event := managers.GetEvent(log.EventId)

				// Decode dynamic event type
				var etype types.EventType
				utils.UnmarshalJSON(event.DynType, &etype)

				// Am I the event sender?
				if event.Sender == from.Address {
					// Get related event metadata
					rcid := etype.Metadata["rcid"].(float64)         // TODO
					cname := managers.GetContainerName(uint64(rcid)) // TODO

					managers.RunTask(ctx, types.RestartContainerTask, cname)
				} else {
					// Execute required task (from dynamic event type)
					managers.RunEventTask(ctx, etype)

					// TODO
					managers.SolveEvent(log.EventId)
				}

				// Solve related event
				//managers.SolveEvent(log.EventId)
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
			fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "DEBUG: EventSolved (EID=", log.EventId, ")\n")

			// Get related event header
			event := managers.GetEvent(log.EventId)

			// Node account
			from := managers.GetFromAccount()

			// Am I the event sender?
			if event.Sender == from.Address {
				// Decode dynamic event type
				var etype types.EventType
				utils.UnmarshalJSON(event.DynType, &etype)

				// Any completion tasks?
				managers.RunEventEndingTask(ctx, etype)
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
			fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "DEBUG: NewContainer (RCID=", log.RegistryCtrId, ", CID=", log.CtrId, ", Host=", log.Host.String(), ")\n")

			// Node account
			from := managers.GetFromAccount()

			// Am I the container host?
			if log.Host == from.Address {
				// Set registry container ID to the local container
				managers.SetContainerName(ctx, log.CtrId, managers.GetContainerName(log.RegistryCtrId))
			}

			// TODO. Refresh local registry
			/*ctrs := managers.GetContainerReg()
			for key := range ctrs {
				fmt.Println(key, " --> ", *ctrs[key])
			}*/
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
			fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "DEBUG: ContainerRemoved (RCID=", log.RegistryCtrId, ")\n")

			// TODO. Refresh local registry
			/*ctrs := managers.GetContainerReg()
			for key := range ctrs {
				fmt.Println(key, " --> ", *ctrs[key])
			}*/
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}
