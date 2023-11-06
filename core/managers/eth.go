package managers

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/swarleynunez/hidra/core/bindings"
	"github.com/swarleynunez/hidra/core/eth"
	"github.com/swarleynunez/hidra/core/types"
	"github.com/swarleynunez/hidra/core/utils"
)

/*// To simulate reputable action execution
type RepAction struct {
	name  string
	count int
}*/

// Instances //
func controllerInstance() (cinst *bindings.Controller) {

	// Controller smart contract address
	caddr := utils.GetEnv("CONTROLLER_ADDR")

	if utils.ValidEthAddress(caddr) {
		inst, err := bindings.NewController(common.HexToAddress(caddr), _ethc)
		utils.CheckError(err, utils.FatalMode)
		cinst = inst
	} else {
		utils.CheckError(eth.ErrMalformedAddr, utils.FatalMode)
	}

	return
}

/*func faucetInstance(faddr common.Address) (finst *bindings.Faucet) {

	finst, err := bindings.NewFaucet(faddr, _ethc)
	utils.CheckError(err, utils.FatalMode)

	return
}*/

func nodeInstance(naddr common.Address) (ninst *bindings.Node) {

	ninst, err := bindings.NewNode(naddr, _ethc)
	utils.CheckError(err, utils.WarningMode)

	return
}

// Setters //
func DeployController(ctx context.Context) common.Address {

	for {
		// Create and configure a transactor
		auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000000)

		// Create smart contract
		caddr, _, _, err := bindings.DeployController(auth, _ethc)

		if err == nil {
			return caddr
		} else {
			utils.CheckError(err, utils.FatalMode)
		}
	}
}

func RegisterNode(ctx context.Context, port uint16) {

	// Txn data encoding
	ns := GetSpecs()
	ns.Port = port
	specs := utils.MarshalJSON(ns)

	for {
		// Create and configure a transactor
		auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000001)

		// Send transaction
		_, err := _cinst.RegisterNode(auth, specs)

		if err == nil {
			return
		} else {
			utils.CheckError(err, utils.FatalMode)
		}
	}
}

// Reputable functions //
func SendEvent(ctx context.Context, etype *types.EventType, rcid uint64) error {

	// Checking zone
	if rcid > 0 { // Has the event a linked container?
		if !existContainer(rcid) ||
			(!isApplicationOwner(GetContainer(rcid).Appid, _from.Address) && !IsContainerHost(rcid, _from.Address)) ||
			isContainerAutodeployed(rcid) ||
			isContainerUnregistered(rcid) {
			return errors.New(SendEventAction + ": transaction not sent")
		}
	}

	// Txn data encoding
	et := utils.MarshalJSON(etype)

	for {
		// Create and configure a transactor
		auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000002)

		// Send transaction
		_, err := _cinst.SendEvent(auth, et, rcid)

		if err != nil {
			//utils.CheckError(err, utils.WarningMode)
			continue
		}
		return err
	}
}

func SendReply(ctx context.Context, eid uint64, repScores []bindings.DELReputationScore) error {

	// Checking zone
	if existEvent(eid) &&
		!isEventSolved(eid) &&
		!GetEvent(eid).HasRequiredReplies &&
		!HasRequiredCount(GetClusterConfig().NodesTh, uint64(GetEventReplyCount(eid))) &&
		!hasAlreadyReplied(eid, _from.Address) {
		reputedNodes := make(map[common.Address]bool)
		for _, v := range repScores {
			if v.Node == GetFromAccount() ||
				!IsNodeRegistered(v.Node) ||
				reputedNodes[v.Node] {
				return errors.New(SendEventAction + ": transaction not sent")
			}
			reputedNodes[v.Node] = true
		}

		for {
			// Create and configure a transactor
			auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000003)

			// Send transaction
			_, err := _cinst.SendReply(auth, eid, repScores)

			if err != nil {
				//utils.CheckError(err, utils.WarningMode)
				continue
			}
			return err
		}
	} else {
		return errors.New(SendReplyAction + ": transaction not sent")
	}
}

func VoteSolver(ctx context.Context, eid uint64, candAddr common.Address) error {

	// Checking zone
	if existEvent(eid) &&
		!isEventSolved(eid) &&
		GetEvent(eid).HasRequiredReplies &&
		!GetEvent(eid).HasRequiredVotes &&
		IsNodeRegistered(candAddr) &&
		!hasAlreadyVoted(eid, _from.Address) {

		for {
			// Create and configure a transactor
			auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000004)

			// Send transaction
			_, err := _cinst.VoteSolver(auth, eid, candAddr)

			if err != nil {
				//utils.CheckError(err, utils.WarningMode)
				continue
			}
			return err
		}
	} else {
		return errors.New(VoteSolverAction + ": transaction not sent")
	}
}

func SolveEvent(ctx context.Context, eid uint64) error {

	// Checking zone
	if existEvent(eid) &&
		!isEventSolved(eid) &&
		canSolveEvent(eid, _from.Address) {

		for {
			// Create and configure a transactor
			auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000005)

			// Send transaction
			_, err := _cinst.SolveEvent(auth, eid)

			if err != nil {
				//utils.CheckError(err, utils.WarningMode)
				continue
			}
			return err
		}
	} else {
		return errors.New(SolveEventAction + ": transaction not sent")
	}
}

func RegisterApplication(ctx context.Context, ainfo *types.ApplicationInfo, cinfos []types.ContainerInfo, autodeploy bool) error {

	// Txn data encoding
	ai := utils.MarshalJSON(ainfo)
	var ci []string
	for _, cinfo := range cinfos {
		ci = append(ci, utils.MarshalJSON(cinfo))
	}

	for {
		// Create and configure a transactor
		auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000006)

		// Send transaction
		_, err := _cinst.RegisterApplication(auth, ai, ci, autodeploy)

		if err != nil {
			//utils.CheckError(err, utils.WarningMode)
			continue
		}
		return err
	}
}

/*func RegisterContainer(ctx context.Context, appid uint64, cinfo *types.ContainerInfo, autodeploy bool) error {

	// Checking zone
	if existApplication(appid) &&
		isApplicationOwner(appid, _from.Address) &&
		!IsApplicationUnregistered(appid) {

		// Txn data encoding
		ci := utils.MarshalJSON(cinfo)

		for {
			// Create and configure a transactor
			auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000007)

			// Send transaction
			_, err := _cinst.RegisterContainer(auth, appid, ci, autodeploy)

			if err != nil {
				utils.CheckError(err, utils.WarningMode)
				continue
			}
			return err
		}
	} else {
		return errors.New(RegisterCtrAction + ": transaction not sent")
	}
}

func ActivateContainer(ctx context.Context, rcid uint64) error {

	appid := GetContainer(rcid).Appid

	// Checking zone
	if existContainer(rcid) &&
		isApplicationOwner(appid, _from.Address) &&
		IsContainerHost(rcid, _from.Address) &&
		!isContainerUnregistered(rcid) &&
		!isContainerActive(rcid) {

		for {
			// Create and configure a transactor
			auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000008)

			// Send transaction
			_, err := _cinst.ActivateContainer(auth, rcid)

			if err != nil {
				utils.CheckError(err, utils.WarningMode)
				continue
			}
			return err
		}
	} else {
		return errors.New(ActivateCtrAction + ": transaction not sent")
	}
}

func UpdateContainerInfo(ctx context.Context, rcid uint64, cinfo *types.ContainerInfo) error {

	limit := getActionLimit(UpdateCtrAction)
	appid := GetContainer(rcid).Appid

	// Checking zone
	if hasNodeReputation(_from.Address, limit) &&
		existContainer(rcid) &&
		isApplicationOwner(appid, _from.Address) &&
		!isContainerInCurrentEvent(rcid) &&
		!isContainerUnregistered(rcid) {

		// Txn data encoding
		ci := utils.MarshalJSON(cinfo)

		for {
			// Create and configure a transactor
			auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000009)

			// Send transaction
			_, err := _cinst.UpdateContainerInfo(auth, rcid, ci)

			if err != nil {
				utils.CheckError(err, utils.WarningMode)
				continue
			}
			return err
		}
	} else {
		return errors.New(UpdateCtrAction + ": transaction not sent")
	}
}

func UnregisterApplication(ctx context.Context, appid uint64) error {

	ctrs := GetApplicationContainers(appid)
	actions := []RepAction{{UnregisterAppAction, 1}, {UnregisterCtrAction, len(ctrs)}}

	// Checking zone
	if hasEstimatedReputation(_from.Address, actions) &&
		existApplication(appid) &&
		isApplicationOwner(appid, _from.Address) &&
		!IsApplicationUnregistered(appid) {
		// Has the application a container that is currently being managed?
		for _, ctr := range ctrs {
			if isContainerInCurrentEvent(ctr) {
				return errors.New(UnregisterAppAction + ": transaction not sent")
			}
		}

		for {
			// Create and configure a transactor
			auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000010)

			// Send transaction
			_, err := _cinst.UnregisterApplication(auth, appid)

			if err != nil {
				utils.CheckError(err, utils.WarningMode)
				continue
			}
			return err
		}
	} else {
		return errors.New(UnregisterAppAction + ": transaction not sent")
	}
}

func UnregisterContainer(ctx context.Context, rcid uint64) error {

	limit := getActionLimit(UnregisterCtrAction)
	appid := GetContainer(rcid).Appid

	// Checking zone
	if hasNodeReputation(_from.Address, limit) &&
		existContainer(rcid) &&
		isApplicationOwner(appid, _from.Address) &&
		!isContainerInCurrentEvent(rcid) &&
		!isContainerUnregistered(rcid) {

		for {
			// Create and configure a transactor
			auth := eth.Transactor(ctx, _ethc, _ks, _from, 8000011)

			// Send transaction
			_, err := _cinst.UnregisterContainer(auth, rcid)

			if err != nil {
				utils.CheckError(err, utils.WarningMode)
				continue
			}
			return err
		}
	} else {
		return errors.New(UnregisterCtrAction + ": transaction not sent")
	}
}*/

// Getters //
/*func getFaucetContract() (faddr common.Address) {

	faddr, err := _cinst.Faucet(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.FatalMode)

	return
}*/

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

func GetClusterConfig() *types.ClusterConfig {

	config, err := _cinst.Config(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	// Convert binding struct to native struct
	c := types.ClusterConfig(config)

	return &c
}

/*func getActionLimit(action string) (limit int64) {

	limit, err := _finst.GetActionLimit(&bind.CallOpts{From: _from.Address}, action)
	utils.CheckError(err, utils.WarningMode)

	return
}

func getActionVariation(action string) (avar int64) {

	avar, err := _finst.GetActionVariation(&bind.CallOpts{From: _from.Address}, action)
	utils.CheckError(err, utils.WarningMode)

	return
}*/

func GetAllNodeSpecs() map[common.Address]string {

	rn, err := _cinst.GetRegisteredNodes(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	nodes := make(map[common.Address]string)
	for _, addr := range rn {
		nodes[addr] = GetNodeSpecs(addr)
	}

	return nodes
}

func GetNodeSpecs(addr common.Address) (specs string) {

	ninst := nodeInstance(getNodeContract(addr))
	specs, err := ninst.GetSpecs(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	return
}

/*func getNodeReputation(addr common.Address) (rep int64) {

	ninst := nodeInstance(getNodeContract(addr))
	rep, err := ninst.GetReputation(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	return
}*/

func GetEvent(eid uint64) *types.Event {

	ce, err := _cinst.Events(&bind.CallOpts{From: _from.Address}, eid)
	utils.CheckError(err, utils.WarningMode)

	// Convert binding struct to native struct
	e := types.Event(ce)

	return &e
}

func GetEventReplyCount(eid uint64) int {

	c, err := _cinst.GetEventReplyCount(&bind.CallOpts{From: _from.Address}, eid)
	utils.CheckError(err, utils.WarningMode)

	return int(c.Int64())
}

func GetEventReplies(eid uint64) (r []types.EventReply) {

	for i := 0; i < GetEventReplyCount(eid); i++ {
		raddr, rss, rat, err := _cinst.GetEventReply(&bind.CallOpts{From: _from.Address}, eid, uint64(i))
		utils.CheckError(err, utils.WarningMode)

		r = append(r, types.EventReply{Replier: raddr, RepScores: rss, RepliedAt: rat})
	}

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

func GetApplicationContainers(appid uint64) (ctrs []uint64) {

	ctrs, err := _cinst.GetApplicationContainers(&bind.CallOpts{From: _from.Address}, appid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func GetApplicationContainersData(appid uint64) map[uint64]*types.Container {

	ac, err := _cinst.GetApplicationContainers(&bind.CallOpts{From: _from.Address}, appid)
	utils.CheckError(err, utils.WarningMode)

	ctrs := make(map[uint64]*types.Container)
	for _, rcid := range ac {
		ctrs[rcid] = GetContainer(rcid)
	}

	return ctrs
}

/*func GetContainersFromCurrentEvents() map[uint64]*types.Container {

	ce, err := _cinst.GetCurrentEvents(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	ctrs := make(map[uint64]*types.Container)
	for _, eid := range ce {
		rcid := GetEvent(eid).Rcid
		if rcid > 0 {
			ctrs[rcid] = GetContainer(rcid)
		}
	}

	return ctrs
}*/

func GetActiveApplications() map[uint64]*types.Application {

	aa, err := _cinst.GetActiveApplications(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	apps := make(map[uint64]*types.Application)
	for _, appid := range aa {
		apps[appid] = GetApplication(appid)
	}

	return apps
}

func GetActiveApplicationsLength() int {

	aa, err := _cinst.GetActiveApplications(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	return len(aa)
}

func GetActiveContainers() map[uint64]*types.Container {

	ac, err := _cinst.GetActiveContainers(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	ctrs := make(map[uint64]*types.Container)
	for _, rcid := range ac {
		ctrs[rcid] = GetContainer(rcid)
	}

	return ctrs
}

func GetActiveContainersLength() int {

	ac, err := _cinst.GetActiveContainers(&bind.CallOpts{From: _from.Address})
	utils.CheckError(err, utils.WarningMode)

	return len(ac)
}

// Helpers //
func IsNodeRegistered(addr common.Address) (r bool) {

	r, err := _cinst.IsNodeRegistered(&bind.CallOpts{From: _from.Address}, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

/*func hasNodeReputation(addr common.Address, lrep int64) (r bool) {

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
}*/

func HasRequiredCount(th uint8, count uint64) (r bool) {

	r, err := _cinst.HasRequiredCount(&bind.CallOpts{From: _from.Address}, th, count)
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

func existEvent(eid uint64) (r bool) {

	r, err := _cinst.ExistEvent(&bind.CallOpts{From: _from.Address}, eid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isEventSolved(eid uint64) (r bool) {

	r, err := _cinst.IsEventSolved(&bind.CallOpts{From: _from.Address}, eid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func canSolveEvent(eid uint64, addr common.Address) (r bool) {

	r, err := _cinst.CanSolveEvent(&bind.CallOpts{From: _from.Address}, eid, addr)
	utils.CheckError(err, utils.WarningMode)

	return
}

func existApplication(appid uint64) (r bool) {

	r, err := _cinst.ExistApplication(&bind.CallOpts{From: _from.Address}, appid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func isApplicationActive(appid uint64) (r bool) {

	r, err := _cinst.IsApplicationActive(&bind.CallOpts{From: _from.Address}, appid)
	utils.CheckError(err, utils.WarningMode)

	return
}

func IsApplicationUnregistered(appid uint64) (r bool) {

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

/*func isContainerInCurrentEvent(rcid uint64) (r bool) {

	r, err := _cinst.IsContainerInCurrentEvent(&bind.CallOpts{From: _from.Address}, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}*/

func isContainerUnregistered(rcid uint64) (r bool) {

	r, err := _cinst.IsContainerUnregistered(&bind.CallOpts{From: _from.Address}, rcid)
	utils.CheckError(err, utils.WarningMode)

	return
}
