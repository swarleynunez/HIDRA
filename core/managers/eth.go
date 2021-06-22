package managers

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/superfog/core/bindings"
	"github.com/swarleynunez/superfog/core/eth"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"os"
	"time"
)

///////////////
// Instances //
///////////////
func controllerInstance() (cinst *bindings.Controller) {

	// Controller smart contract address
	caddr := os.Getenv("CONTROLLER_ADDR")

	if utils.ValidEthAddress(caddr) {
		// Get instance
		inst, err := bindings.NewController(common.HexToAddress(caddr), _ethc)
		utils.CheckError(err, utils.FatalMode)
		cinst = inst

		// Debug
		fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Loaded controller address: ", caddr, "\n")
	} else {
		utils.CheckError(eth.ErrMalformedAddr, utils.FatalMode)
	}

	return
}

func faucetInstance(faddr common.Address) (finst *bindings.Faucet) {

	finst, err := bindings.NewFaucet(faddr, _ethc)
	utils.CheckError(err, utils.FatalMode)

	return
}

func nodeInstance(naddr common.Address) (ninst *bindings.Node) {

	ninst, err := bindings.NewNode(naddr, _ethc)
	utils.CheckError(err, utils.WarningMode)

	return
}

/////////////
// Setters //
/////////////
func DeployController() {

	// Create and configure a transactor
	_nmutex.Lock()
	auth := eth.GetTransactor(_ks, _from, _nonce, _ethc, DeployControllerGasLimit)
	_nonce++ // Next nonce for the next transaction
	_nmutex.Unlock()

	// Create smart contract
	caddr, _, _, err := bindings.DeployController(auth, _ethc)
	utils.CheckError(err, utils.FatalMode)

	// Save the controller address
	utils.SetEnvKey("CONTROLLER_ADDR", caddr.String())

	// Debug
	fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Controller address: ", caddr.String(), "\n")
}

func RegisterNode() {

	if !IsNodeRegistered(_from.Address) {

		specs := GetNodeSpecs()

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, _ethc, RegisterNodeGasLimit)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.RegisterNode(auth, utils.MarshalJSON(specs))
		utils.CheckError(err, utils.FatalMode)

		// Debug
		fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Node registered\n")
	} else {
		// Debug
		fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Node is already registered\n")
	}
}

////////////////
// Reputables //
////////////////
func SendEvent(etype *types.EventType, state *types.State) {

	// Reputation limit for this function
	limit := getActionLimit(SendEventAction)

	if hasNodeReputation(limit) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, _ethc, SendEventGasLimit)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		createdAt := uint64(time.Now().Unix())
		_, err := _cinst.SendEvent(auth, utils.MarshalJSON(etype), createdAt, utils.MarshalJSON(state))
		utils.CheckError(err, utils.WarningMode)
	}
}

func SendReply(eid uint64, state *types.State) {

	// Reputation limit for this function
	limit := getActionLimit(SendReplyAction)

	if hasNodeReputation(limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		!hasAlreadyReplied(eid, _from.Address) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, _ethc, SendReplyGasLimit)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		createdAt := uint64(time.Now().Unix())
		_, err := _cinst.SendReply(auth, eid, utils.MarshalJSON(state), createdAt)
		utils.CheckError(err, utils.WarningMode)
	}
}

func VoteSolver(eid uint64, addr common.Address) {

	// Reputation limit for this function
	limit := getActionLimit(VoteSolverAction)

	if hasNodeReputation(limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		hasAlreadyReplied(eid, addr) &&
		!hasAlreadyVoted(eid) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, _ethc, VoteSolverGasLimit)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.VoteSolver(auth, eid, addr)
		utils.CheckError(err, utils.WarningMode)
	}
}

func SolveEvent(eid uint64) {

	// Reputation limit for this function
	limit := getActionLimit(SolveEventAction)

	if hasNodeReputation(limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		canSolveEvent(eid) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, _ethc, SolveEventGasLimit)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		solvedAt := uint64(time.Now().Unix())
		_, err := _cinst.SolveEvent(auth, eid, solvedAt)
		utils.CheckError(err, utils.WarningMode)
	}
}

// About distributed registry
func recordContainerOnReg(cinfo *types.ContainerInfo, startedAt uint64, cid string) {

	// Reputation limit for this function
	limit := getActionLimit(RecordContainerAction)

	if hasNodeReputation(limit) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, _ethc, RecordContainerGasLimit)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.RecordContainer(auth, utils.MarshalJSON(cinfo), startedAt, cid)
		utils.CheckError(err, utils.WarningMode)
	}
}

// About distributed registry
func removeContainerFromReg(rcid uint64, finishedAt uint64) {

	// Reputation limit for this function
	limit := getActionLimit(RemoveContainerAction)

	if hasNodeReputation(limit) &&
		existRegContainer(rcid) &&
		isRegContainerHost(rcid) &&
		isRegContainerActive(rcid) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, _ethc, RemoveContainerGasLimit)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.RemoveContainer(auth, rcid, finishedAt)
		utils.CheckError(err, utils.WarningMode)
	}
}

/////////////
// Getters //
/////////////
func getFaucetAddress() (faddr common.Address) {

	faddr, err := _cinst.Faucet(nil)
	utils.CheckError(err, utils.FatalMode)

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

// Get node specs from its smart contract
func GetSpecs(addr common.Address) *types.NodeSpecs {

	ninst := nodeInstance(getNodeContract(addr))
	ss, err := ninst.NodeSpecs(nil)
	utils.CheckError(err, utils.WarningMode)

	// Decoding
	var ns types.NodeSpecs
	utils.UnmarshalJSON(ss, &ns)

	return &ns
}

// Get network events from controller smart contract
func GetEvent(eid uint64) *types.Event {

	ce, err := _cinst.Events(nil, eid)
	utils.CheckError(err, utils.WarningMode)

	// Convert contract event type
	e := types.Event(ce)

	return &e
}

// About distributed registry
func getRegActiveContainers() (ctrs []uint64) {

	ctrs, err := _cinst.GetActiveContainers(nil)
	utils.CheckError(err, utils.WarningMode)

	return
}

// About distributed registry
func getRegContainer(rcid uint64) *types.Container {

	ctr, err := _cinst.Containers(nil, rcid)
	utils.CheckError(err, utils.WarningMode)

	// Convert contract container type
	c := types.Container(ctr)

	return &c
}

// About distributed registry
func GetContainerReg() map[uint64]*types.Container {

	ac := getRegActiveContainers()

	ctrs := make(map[uint64]*types.Container)
	for i := range ac {
		ctrs[ac[i]] = getRegContainer(ac[i])
	}

	return ctrs
}

/////////////
// Helpers //
/////////////
func IsNodeRegistered(addr common.Address) (r bool) {

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

// About distributed registry
func existRegContainer(rcid uint64) (r bool) {

	r, err := _cinst.ExistContainer(nil, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}

// About distributed registry
func isRegContainerHost(rcid uint64) (r bool) {

	r, err := _cinst.IsContainerHost(nil, _from.Address, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}

// About distributed registry
func isRegContainerActive(rcid uint64) (r bool) {

	r, err := _cinst.IsContainerActive(nil, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}
