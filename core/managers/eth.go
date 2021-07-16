package managers

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/superfog/core/bindings"
	"github.com/swarleynunez/superfog/core/eth"
	"github.com/swarleynunez/superfog/core/types"
	"github.com/swarleynunez/superfog/core/utils"
	"os"
)

// To simulate reputable action execution
type RepAction struct {
	name  string
	count int
}

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
func DeployController() (caddr common.Address) {

	// Create and configure a transactor
	_nmutex.Lock()
	auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
	_nonce++ // Next nonce for the next transaction
	_nmutex.Unlock()

	// Create smart contract
	caddr, _, _, err := bindings.DeployController(auth, _ethc)
	utils.CheckError(err, utils.FatalMode)

	return
}

func RegisterNode() {

	// Create and configure a transactor
	_nmutex.Lock()
	auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
	_nonce++ // Next nonce for the next transaction
	_nmutex.Unlock()

	// Send transaction
	_, err := _cinst.RegisterNode(auth, utils.MarshalJSON(GetSpecs()))
	utils.CheckError(err, utils.FatalMode)
}

/////////////////////////
// Reputable functions //
/////////////////////////
func SendEvent(etype *types.EventType, rcid uint64, nstate *types.State) {

	limit := getActionLimit(SendEventAction)

	// Checking zone
	if hasNodeReputation(_from.Address, limit) {
		// Has the event a linked container?
		if rcid > 0 {
			appid := GetContainer(rcid).Appid
			if !existContainer(rcid) ||
				(!isApplicationOwner(appid, _from.Address) && !IsContainerHost(rcid, _from.Address)) ||
				isContainerAutodeployed(rcid) ||
				isContainerUnregistered(rcid) {
				return
			}
		}

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.SendEvent(auth, utils.MarshalJSON(etype), rcid, utils.MarshalJSON(nstate))
		utils.CheckError(err, utils.WarningMode)
	}
}

func SendReply(eid uint64, nstate *types.State) {

	limit := getActionLimit(SendReplyAction)

	// Checking zone
	if hasNodeReputation(_from.Address, limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		!hasAlreadyReplied(eid, _from.Address) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.SendReply(auth, eid, utils.MarshalJSON(nstate))
		utils.CheckError(err, utils.WarningMode)
	}
}

func VoteSolver(eid uint64, candAddr common.Address) {

	limit := getActionLimit(VoteSolverAction)
	thld := getClusterConfig().NodesThld
	count := uint64(len(GetEventReplies(eid)))

	// Checking zone
	if hasNodeReputation(_from.Address, limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		hasRequiredCount(thld, count) &&
		hasAlreadyReplied(eid, candAddr) &&
		!hasAlreadyVoted(eid, _from.Address) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.VoteSolver(auth, eid, candAddr)
		utils.CheckError(err, utils.WarningMode)
	}
}

func SolveEvent(eid uint64) {

	limit := getActionLimit(SolveEventAction)

	// Checking zone
	if hasNodeReputation(_from.Address, limit) &&
		existEvent(eid) &&
		!isEventSolved(eid) &&
		CanSolveEvent(eid, _from.Address) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.SolveEvent(auth, eid)
		utils.CheckError(err, utils.WarningMode)
	}
}

func RegisterApplication(ainfo *types.ApplicationInfo, cinfos []types.ContainerInfo, autodeploy bool) {

	// To simulate reputation
	actions := []RepAction{{RegisterAppAction, 1}, {RegisterCtrAction, len(cinfos)}}

	// Checking zone
	if hasEstimatedReputation(_from.Address, actions) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Encode info of each container
		var ci []string
		for i := range cinfos {
			ci = append(ci, utils.MarshalJSON(cinfos[i]))
		}

		// Send transaction
		_, err := _cinst.RegisterApplication(auth, utils.MarshalJSON(ainfo), ci, autodeploy)
		utils.CheckError(err, utils.WarningMode)
	}
}

func RegisterContainer(appid uint64, cinfo *types.ContainerInfo, autodeploy bool) {

	limit := getActionLimit(RegisterCtrAction)

	// Checking zone
	if hasNodeReputation(_from.Address, limit) &&
		existApplication(appid) &&
		isApplicationOwner(appid, _from.Address) &&
		!isApplicationUnregistered(appid) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.RegisterContainer(auth, appid, utils.MarshalJSON(cinfo), autodeploy)
		utils.CheckError(err, utils.WarningMode)
	}
}

func ActivateContainer(rcid uint64) {

	limit := getActionLimit(ActivateCtrAction)
	appid := GetContainer(rcid).Appid

	// Checking zone
	if hasNodeReputation(_from.Address, limit) &&
		existContainer(rcid) &&
		isApplicationOwner(appid, _from.Address) &&
		IsContainerHost(rcid, _from.Address) &&
		!isContainerUnregistered(rcid) &&
		!isContainerActive(rcid) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.ActivateContainer(auth, rcid)
		utils.CheckError(err, utils.WarningMode)
	}
}

func UpdateContainerInfo(rcid uint64, cinfo *types.ContainerInfo) {

	limit := getActionLimit(UpdateCtrAction)
	appid := GetContainer(rcid).Appid

	// Checking zone
	if hasNodeReputation(_from.Address, limit) &&
		existContainer(rcid) &&
		isApplicationOwner(appid, _from.Address) &&
		!isContainerUnregistered(rcid) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.UpdateContainerInfo(auth, rcid, utils.MarshalJSON(cinfo))
		utils.CheckError(err, utils.WarningMode)
	}
}

func UnregisterApplication(appid uint64) {

	// To simulate reputation
	count := len(getApplicationContainers(appid))
	actions := []RepAction{{UnregisterAppAction, 1}, {UnregisterCtrAction, count}}

	// Checking zone
	if hasEstimatedReputation(_from.Address, actions) &&
		existApplication(appid) &&
		isApplicationOwner(appid, _from.Address) &&
		!isApplicationUnregistered(appid) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.UnregisterApplication(auth, appid)
		utils.CheckError(err, utils.WarningMode)
	}
}

func UnregisterContainer(rcid uint64) {

	limit := getActionLimit(UnregisterCtrAction)
	appid := GetContainer(rcid).Appid

	// Checking zone
	if hasNodeReputation(_from.Address, limit) &&
		existContainer(rcid) &&
		isApplicationOwner(appid, _from.Address) &&
		!isContainerUnregistered(rcid) {

		// Create and configure a transactor
		_nmutex.Lock()
		auth := eth.GetTransactor(_ks, _from, _nonce, 5000000)
		_nonce++ // Next nonce for the next transaction
		_nmutex.Unlock()

		// Send transaction
		_, err := _cinst.UnregisterContainer(auth, rcid)
		utils.CheckError(err, utils.WarningMode)
	}
}

/////////////
// Getters //
/////////////
func getFaucetContract() (faddr common.Address) {

	faddr, err := _cinst.Faucet(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.FatalMode)

	return
}

func getNodeContract(addr common.Address) (naddr common.Address) {

	naddr, err := _cinst.Nodes(&bind.CallOpts{From: _from.Address}, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getClusterState() *types.ClusterState {

	state, err := _cinst.State(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	// Convert binding struct to native struct
	s := types.ClusterState(state)

	return &s
}

func getClusterConfig() *types.ClusterConfig {

	config, err := _cinst.Config(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	// Convert binding struct to native struct
	c := types.ClusterConfig(config)

	return &c
}

func getActionLimit(action string) (limit int64) {

	limit, err := _finst.GetActionLimit(&bind.CallOpts{From: _from.Address}, action)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getActionVariation(action string) (avar int64) {

	avar, err := _finst.GetActionVariation(&bind.CallOpts{From: _from.Address}, action)
	utils.CheckError(err, utils.WarningMode)

	return
}

func GetNodeSpecs(addr common.Address) (specs string) {

	ninst := nodeInstance(getNodeContract(addr))
	specs, err := ninst.GetSpecs(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	return
}

func getNodeReputation(addr common.Address) (rep int64) {

	ninst := nodeInstance(getNodeContract(addr))
	rep, err := ninst.GetReputation(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	return
}

func GetEvent(eid uint64) *types.Event {

	ce, err := _cinst.Events(&bind.CallOpts{From: _from.Address}, eid)
	utils.CheckError(err, utils.WarningMode)

	// Convert binding struct to native struct
	e := types.Event(ce)

	return &e
}

func GetEventReplies(eid uint64) (r []bindings.DELEventReply) {

	r, err := _cinst.GetEventReplies(&bind.CallOpts{From: _from.Address}, eid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func GetApplication(appid uint64) *types.Application {

	app, err := _cinst.Apps(&bind.CallOpts{From: _from.Address}, appid)
	utils.CheckError(err, utils.WarningMode)

	// Convert binding struct to native struct
	a := types.Application(app)

	return &a
}

func GetContainer(rcid uint64) *types.Container {

	ctr, err := _cinst.Ctrs(&bind.CallOpts{From: _from.Address}, rcid)
	utils.CheckError(err, utils.WarningMode)

	// Convert binding struct to native struct
	c := types.Container(ctr)

	return &c
}

func GetContainerInstances(rcid uint64) (insts []bindings.DCRContainerInstance) {

	insts, err := _cinst.GetContainerInstances(&bind.CallOpts{From: _from.Address}, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getActiveApplications() map[uint64]*types.Application {

	aa, err := _cinst.GetActiveApplications(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	apps := make(map[uint64]*types.Application)
	for i := range aa {
		apps[aa[i]] = GetApplication(aa[i])
	}

	return apps
}

func getApplicationContainers(appid uint64) map[uint64]*types.Container {

	ac, err := _cinst.GetApplicationContainers(&bind.CallOpts{From: _from.Address}, appid)
	utils.CheckError(err, utils.WarningMode)

	ctrs := make(map[uint64]*types.Container)
	for i := range ac {
		ctrs[ac[i]] = GetContainer(ac[i])
	}

	return ctrs
}

func GetActiveContainers() map[uint64]*types.Container {

	ac, err := _cinst.GetActiveContainers(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	ctrs := make(map[uint64]*types.Container)
	for i := range ac {
		ctrs[ac[i]] = GetContainer(ac[i])
	}

	return ctrs
}

/////////////
// Helpers //
/////////////
func IsNodeRegistered(addr common.Address) (r bool) {

	r, err := _cinst.IsNodeRegistered(&bind.CallOpts{From: _from.Address}, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasNodeReputation(addr common.Address, lrep int64) (r bool) {

	// Check if the node has enough reputation (greater or equal than a limit)
	r, err := _cinst.HasNodeReputation(&bind.CallOpts{From: _from.Address}, addr, lrep)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasEstimatedReputation(addr common.Address, actions []RepAction) (r bool) {

	// Initial node reputation
	rep := getNodeReputation(addr)

	for i := range actions {
		limit := getActionLimit(actions[i].name)

		// For each execution
		for j := 0; j < actions[i].count; j++ {
			if rep < limit {
				return false
			}

			// Simulate next reputation
			rep += getActionVariation(actions[i].name)
		}
	}

	return true
}

func hasRequiredCount(thld uint8, count uint64) (r bool) {

	r, err := _cinst.HasRequiredCount(&bind.CallOpts{From: _from.Address}, thld, count)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isApplicationOwner(appid uint64, addr common.Address) (r bool) {

	r, err := _cinst.IsApplicationOwner(&bind.CallOpts{From: _from.Address}, appid, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func IsContainerHost(rcid uint64, addr common.Address) (r bool) {

	r, err := _cinst.IsContainerHost(&bind.CallOpts{From: _from.Address}, rcid, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func existEvent(eid uint64) (r bool) {

	r, err := _cinst.ExistEvent(&bind.CallOpts{From: _from.Address}, eid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasAlreadyReplied(eid uint64, addr common.Address) (r bool) {

	r, err := _cinst.HasAlreadyReplied(&bind.CallOpts{From: _from.Address}, eid, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func hasAlreadyVoted(eid uint64, addr common.Address) (r bool) {

	r, err := _cinst.HasAlreadyVoted(&bind.CallOpts{From: _from.Address}, eid, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isEventSolved(eid uint64) (r bool) {

	r, err := _cinst.IsEventSolved(&bind.CallOpts{From: _from.Address}, eid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func CanSolveEvent(eid uint64, addr common.Address) (r bool) {

	r, err := _cinst.CanSolveEvent(&bind.CallOpts{From: _from.Address}, eid, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func existApplication(appid uint64) (r bool) {

	r, err := _cinst.ExistApplication(&bind.CallOpts{From: _from.Address}, appid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isApplicationUnregistered(appid uint64) (r bool) {

	r, err := _cinst.IsApplicationUnregistered(&bind.CallOpts{From: _from.Address}, appid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func existContainer(rcid uint64) (r bool) {

	r, err := _cinst.ExistContainer(&bind.CallOpts{From: _from.Address}, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isContainerAutodeployed(rcid uint64) (r bool) {

	r, err := _cinst.IsContainerAutodeployed(&bind.CallOpts{From: _from.Address}, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isContainerActive(rcid uint64) (r bool) {

	r, err := _cinst.IsContainerActive(&bind.CallOpts{From: _from.Address}, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isContainerUnregistered(rcid uint64) (r bool) {

	r, err := _cinst.IsContainerUnregistered(&bind.CallOpts{From: _from.Address}, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}
