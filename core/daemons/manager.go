package daemons

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/swarleynunez/superfog/core/contracts"
	"github.com/swarleynunez/superfog/core/eth"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"os"
	"time"
)

const (
	// Gas limit of each smart contract function
	DeployControllerGasLimit uint64 = 3000000
	RegisterNodeGasLimit     uint64 = 530000
	SendEventGasLimit        uint64 = 350000
	SendReplyGasLimit        uint64 = 230000
	VoteSolverGasLimit       uint64 = 180000
	SolveEventGasLimit       uint64 = 100000

	// Reputable actions
	SendEventAction  = "sendEvent"
	SendReplyAction  = "sendReply"
	VoteSolverAction = "voteSolver"
	SolveEventAction = "solveEvent"
)

// Common parameters to all functions
var (
	ethc  *ethclient.Client
	ks    *keystore.KeyStore
	cinst *contracts.Controller
	finst *contracts.Faucet
	from  accounts.Account
)

//////////////////
// Initializers //
//////////////////
func Init(_ethc *ethclient.Client, _ks *keystore.KeyStore, _from accounts.Account) {

	// Ethereum node
	ethc = _ethc
	ks = _ks

	// Selected Ethereum account
	from = _from

	// Deploy controller smart contract or get an instance
	cinst = controllerInstance()

	// Faucet instance
	finst = faucetInstance(getFaucetAddress())
}

func RegisterNode() {

	if !isNodeRegistered(from.Address) {

		// Get host specs and set the first state
		specs, _ := firstHostState()

		// Create and configure a transactor
		auth := eth.GetTransactor(ks, from, ethc, RegisterNodeGasLimit)

		// Send transaction
		_, err := cinst.RegisterNode(auth, utils.MarshalJSON(specs))
		utils.CheckError(err, utils.WarningMode)
	}
}

///////////////
// Instances //
///////////////
func controllerInstance() (cinst *contracts.Controller) {

	// Controller smart contract address
	caddr := os.Getenv("CONTROLLER_ADDR")

	if utils.ValidEthAddress(caddr) {
		// Get instance
		inst, err := contracts.NewController(common.HexToAddress(caddr), ethc)
		utils.CheckError(err, utils.WarningMode)
		cinst = inst
	} else {
		//utils.CheckError(eth.ErrMalformedAddr, utils.PanicMode)

		// TODO
		// Create and configure a transactor
		auth := eth.GetTransactor(ks, from, ethc, DeployControllerGasLimit)

		// Create smart contract
		caddr, _, inst, err := contracts.DeployController(auth, ethc)
		utils.CheckError(err, utils.WarningMode)
		cinst = inst

		if err == nil {
			// Save the controller address
			utils.SetEnvKey("CONTROLLER_ADDR", caddr.String())
		}
	}

	return
}

func faucetInstance(faddr common.Address) (finst *contracts.Faucet) {

	finst, err := contracts.NewFaucet(faddr, ethc)
	utils.CheckError(err, utils.WarningMode)

	return
}

func nodeInstance(naddr common.Address) (ninst *contracts.Node) {

	ninst, err := contracts.NewNode(naddr, ethc)
	utils.CheckError(err, utils.WarningMode)

	return
}

////////////////
// Reputables //
////////////////
func sendEvent(etype *types.EventType, state *types.NodeState) {

	// Reputation limit for this function
	limit := getActionLimit(SendEventAction)

	if isNodeRegistered(from.Address) && hasNodeReputation(limit) {

		// Create and configure a transactor
		auth := eth.GetTransactor(ks, from, ethc, SendEventGasLimit)

		// Send transaction
		createdAt := uint64(time.Now().Unix())
		_, err := cinst.SendEvent(auth, utils.MarshalJSON(etype), createdAt, utils.MarshalJSON(state))
		utils.CheckError(err, utils.WarningMode)
	}
}

func sendReply(eid uint64, state *types.NodeState) {

	// Reputation limit for this function
	limit := getActionLimit(SendReplyAction)

	if isNodeRegistered(from.Address) &&
		hasNodeReputation(limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		!hasAlreadyReplied(eid, from.Address) {

		// Create and configure a transactor
		auth := eth.GetTransactor(ks, from, ethc, SendReplyGasLimit)

		// Send transaction
		createdAt := uint64(time.Now().Unix())
		_, err := cinst.SendReply(auth, eid, utils.MarshalJSON(state), createdAt)
		utils.CheckError(err, utils.WarningMode)
	}
}

func voteSolver(eid uint64, addr common.Address) {

	// Reputation limit for this function
	limit := getActionLimit(VoteSolverAction)

	if isNodeRegistered(from.Address) &&
		hasNodeReputation(limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		hasAlreadyReplied(eid, addr) &&
		!hasAlreadyVoted(eid) {

		// Create and configure a transactor
		auth := eth.GetTransactor(ks, from, ethc, VoteSolverGasLimit)

		// Send transaction
		_, err := cinst.VoteSolver(auth, eid, addr)
		utils.CheckError(err, utils.WarningMode)
	}
}

func solveEvent(eid uint64) {

	// Reputation limit for this function
	limit := getActionLimit(SolveEventAction)

	if isNodeRegistered(from.Address) &&
		hasNodeReputation(limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		canSolveEvent(eid) {

		// Create and configure a transactor
		auth := eth.GetTransactor(ks, from, ethc, SolveEventGasLimit)

		// Send transaction
		solvedAt := uint64(time.Now().Unix())
		_, err := cinst.SolveEvent(auth, eid, solvedAt)
		utils.CheckError(err, utils.WarningMode)
	}
}

//////////////
// Watchers //
//////////////
func watchNewEvent() {

	// Event/log channel
	logs := make(chan *contracts.ControllerNewEvent)

	// Subscription to the event
	sub, err := cinst.WatchNewEvent(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Println("------------------------")
			fmt.Print("DEBUG: NewEvent (EID=", log.EventId, ")\n")

			// Send a event reply containing the current host state
			sendReply(log.EventId, getHostState())
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

func watchRequiredReplies() {

	// Event/log channel
	logs := make(chan *contracts.ControllerRequiredReplies)

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
			solver := selectBestSolver(log.EventId)
			voteSolver(log.EventId, solver)
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

func watchRequiredVotes() {

	// Event/log channel
	logs := make(chan *contracts.ControllerRequiredVotes)

	// Subscription to the event
	sub, err := cinst.WatchRequiredVotes(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: RequiredVotes (EID=", log.EventId, ", Solver=", log.Solver.String(), ")\n")

			// I am the voted solver?
			if log.Solver == from.Address {
				// Execute required task (from dynamic event type)
				runEventTask(log.EventId)

				// Solve related event
				solveEvent(log.EventId)
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

func watchEventSolved() {

	// Event/log channel
	logs := make(chan *contracts.ControllerEventSolved)

	// Subscription to the event
	sub, err := cinst.WatchEventSolved(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: EventSolved (EID=", log.EventId, ", Sender=", log.Sender.String(), ")\n")

			// I am the event sender?
			if log.Sender == from.Address {
				// Any completion tasks?
				runEndingTasks(log.EventId)
			}
			fmt.Println("------------------------")
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

/////////////
// Getters //
/////////////
func getFaucetAddress() (faddr common.Address) {

	faddr, err := cinst.Faucet(nil)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getNodeContract(addr common.Address) (naddr common.Address) {

	naddr, err := cinst.Nodes(nil, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getActionLimit(action string) (rep int64) {

	rep, err := finst.GetActionLimit(nil, action)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getNodeReputation(addr common.Address) (rep int64) {

	ninst := nodeInstance(getNodeContract(addr))
	rep, err := ninst.GetReputation(nil)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getNodeSpecs(addr common.Address) *types.NodeSpecs {

	ninst := nodeInstance(getNodeContract(addr))
	ss, err := ninst.NodeSpecs(nil)
	utils.CheckError(err, utils.WarningMode)

	// Decoding
	var specs types.NodeSpecs
	utils.UnmarshalJSON(ss, &specs)

	return &specs
}

func getEvent(eid uint64) *types.Event {

	ce, err := cinst.Events(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	// Convert contract event type
	e := types.Event(ce)

	return &e
}

/////////////
// Helpers //
/////////////
func isNodeRegistered(addr common.Address) (r bool) {

	r, err := cinst.IsNodeRegistered(nil, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasNodeReputation(lrep int64) (r bool) {

	// Check if the node has enough reputation (greater or equal than a limit)
	r, err := cinst.HasNodeReputation(nil, from.Address, lrep)
	utils.CheckError(err, utils.WarningMode)

	return
}

func existEvent(eid uint64) (r bool) {

	r, err := cinst.ExistEvent(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isEventSolved(eid uint64) (r bool) {

	r, err := cinst.IsEventSolved(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasAlreadyReplied(eid uint64, addr common.Address) (r bool) {

	r, err := cinst.HasAlreadyReplied(nil, eid, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasAlreadyVoted(eid uint64) (r bool) {

	r, err := cinst.HasAlreadyVoted(nil, eid, from.Address)
	utils.CheckError(err, utils.WarningMode)

	return
}

func canSolveEvent(eid uint64) (r bool) {

	r, err := cinst.CanSolveEvent(nil, eid, from.Address)
	utils.CheckError(err, utils.WarningMode)

	return
}
