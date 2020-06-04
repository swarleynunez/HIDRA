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

// Unexported and "readonly" global parameters
var (
	_ethc  *ethclient.Client
	_ks    *keystore.KeyStore
	_from  accounts.Account
	_cinst *contracts.Controller
	_finst *contracts.Faucet
)

func Init(ethc *ethclient.Client, ks *keystore.KeyStore, from accounts.Account) {

	// Ethereum node
	_ethc = ethc
	_ks = ks

	// Selected Ethereum account
	_from = from

	// Deploy controller smart contract or get an instance
	_cinst = controllerInstance()

	// Faucet instance
	_finst = faucetInstance(getFaucetAddress())

	// Register node in the network
	registerNode()
}

func registerNode() {

	if !isNodeRegistered(_from.Address) {

		// Get node specs
		specs := getSpecs()

		// Create and configure a transactor
		auth := eth.GetTransactor(_ks, _from, _ethc, RegisterNodeGasLimit)

		// Send transaction
		_, err := _cinst.RegisterNode(auth, utils.MarshalJSON(specs))
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
		inst, err := contracts.NewController(common.HexToAddress(caddr), _ethc)
		utils.CheckError(err, utils.WarningMode)
		cinst = inst
	} else {
		//utils.CheckError(eth.ErrMalformedAddr, utils.PanicMode)

		// TODO
		// Create and configure a transactor
		auth := eth.GetTransactor(_ks, _from, _ethc, DeployControllerGasLimit)

		// Create smart contract
		caddr, _, inst, err := contracts.DeployController(auth, _ethc)
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

	finst, err := contracts.NewFaucet(faddr, _ethc)
	utils.CheckError(err, utils.WarningMode)

	return
}

func nodeInstance(naddr common.Address) (ninst *contracts.Node) {

	ninst, err := contracts.NewNode(naddr, _ethc)
	utils.CheckError(err, utils.WarningMode)

	return
}

////////////////
// Reputables //
////////////////
func sendEvent(etype *types.EventType, state *types.NodeState) {

	// Reputation limit for this function
	limit := getActionLimit(SendEventAction)

	if isNodeRegistered(_from.Address) && hasNodeReputation(limit) {

		// Create and configure a transactor
		auth := eth.GetTransactor(_ks, _from, _ethc, SendEventGasLimit)

		// Send transaction
		createdAt := uint64(time.Now().Unix())
		_, err := _cinst.SendEvent(auth, utils.MarshalJSON(etype), createdAt, utils.MarshalJSON(state))
		utils.CheckError(err, utils.WarningMode)
	}
}

func sendReply(eid uint64, state *types.NodeState) {

	// Reputation limit for this function
	limit := getActionLimit(SendReplyAction)

	if isNodeRegistered(_from.Address) &&
		hasNodeReputation(limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		!hasAlreadyReplied(eid, _from.Address) {

		// Create and configure a transactor
		auth := eth.GetTransactor(_ks, _from, _ethc, SendReplyGasLimit)

		// Send transaction
		createdAt := uint64(time.Now().Unix())
		_, err := _cinst.SendReply(auth, eid, utils.MarshalJSON(state), createdAt)
		utils.CheckError(err, utils.WarningMode)
	}
}

func voteSolver(eid uint64, addr common.Address) {

	// Reputation limit for this function
	limit := getActionLimit(VoteSolverAction)

	if isNodeRegistered(_from.Address) &&
		hasNodeReputation(limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		hasAlreadyReplied(eid, addr) &&
		!hasAlreadyVoted(eid) {

		// Create and configure a transactor
		auth := eth.GetTransactor(_ks, _from, _ethc, VoteSolverGasLimit)

		// Send transaction
		_, err := _cinst.VoteSolver(auth, eid, addr)
		utils.CheckError(err, utils.WarningMode)
	}
}

func solveEvent(eid uint64) {

	// Reputation limit for this function
	limit := getActionLimit(SolveEventAction)

	if isNodeRegistered(_from.Address) &&
		hasNodeReputation(limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		canSolveEvent(eid) {

		// Create and configure a transactor
		auth := eth.GetTransactor(_ks, _from, _ethc, SolveEventGasLimit)

		// Send transaction
		solvedAt := uint64(time.Now().Unix())
		_, err := _cinst.SolveEvent(auth, eid, solvedAt)
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
	sub, err := _cinst.WatchNewEvent(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: NewEvent (EID=", log.EventId, ")\n")

			// Send a event reply containing the current node state
			sendReply(log.EventId, getState())
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

func watchRequiredReplies() {

	// Event/log channel
	logs := make(chan *contracts.ControllerRequiredReplies)

	// Subscription to the event
	sub, err := _cinst.WatchRequiredReplies(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: RequiredReplies (EID=", log.EventId, ")\n")

			// Select and vote the best event solver
			solver := selectBestSolver(log.EventId)
			if !utils.EmptyEthAddress(solver.String()) {
				voteSolver(log.EventId, solver)
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

func watchRequiredVotes() {

	// Event/log channel
	logs := make(chan *contracts.ControllerRequiredVotes)

	// Subscription to the event
	sub, err := _cinst.WatchRequiredVotes(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: RequiredVotes (EID=", log.EventId, ", Solver=", log.Solver.String(), ")\n")

			// I am the voted solver?
			if log.Solver == _from.Address {
				// Execute required task (from dynamic event type)
				runEventTask(log.EventId, types.CreateTask)

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
	sub, err := _cinst.WatchEventSolved(nil, logs)
	utils.CheckError(err, utils.WarningMode)

	// Infinite loop
	for {
		select {
		case log := <-logs:

			// Debug
			fmt.Print("DEBUG: EventSolved (EID=", log.EventId, ", Sender=", log.Sender.String(), ")\n")

			// I am the event sender?
			if log.Sender == _from.Address {
				// Any completion tasks?
				runEventEndingTask(log.EventId, types.CreateTask)
			}
		case err := <-sub.Err():
			utils.CheckError(err, utils.WarningMode)
		}
	}
}

/////////////
// Getters //
/////////////
func getFaucetAddress() (faddr common.Address) {

	faddr, err := _cinst.Faucet(nil)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getNodeContract(addr common.Address) (naddr common.Address) {

	naddr, err := _cinst.Nodes(nil, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getActionLimit(action string) (rep int64) {

	rep, err := _finst.GetActionLimit(nil, action)
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
	var ns types.NodeSpecs
	utils.UnmarshalJSON(ss, &ns)

	return &ns
}

func getEvent(eid uint64) *types.Event {

	ce, err := _cinst.Events(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	// Convert contract event type
	e := types.Event(ce)

	return &e
}

/////////////
// Helpers //
/////////////
func isNodeRegistered(addr common.Address) (r bool) {

	r, err := _cinst.IsNodeRegistered(nil, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasNodeReputation(lrep int64) (r bool) {

	// Check if the node has enough reputation (greater or equal than a limit)
	r, err := _cinst.HasNodeReputation(nil, _from.Address, lrep)
	utils.CheckError(err, utils.WarningMode)

	return
}

func existEvent(eid uint64) (r bool) {

	r, err := _cinst.ExistEvent(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isEventSolved(eid uint64) (r bool) {

	r, err := _cinst.IsEventSolved(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasAlreadyReplied(eid uint64, addr common.Address) (r bool) {

	r, err := _cinst.HasAlreadyReplied(nil, eid, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasAlreadyVoted(eid uint64) (r bool) {

	r, err := _cinst.HasAlreadyVoted(nil, eid, _from.Address)
	utils.CheckError(err, utils.WarningMode)

	return
}

func canSolveEvent(eid uint64) (r bool) {

	r, err := _cinst.CanSolveEvent(nil, eid, _from.Address)
	utils.CheckError(err, utils.WarningMode)

	return
}
