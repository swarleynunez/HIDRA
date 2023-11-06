package managers

/*func init() {

	InitNode(context.Background(), true)

	// New controller smart contract for each test execution
	caddr := DeployController()
	utils.SetEnvKey("CONTROLLER_ADDR", caddr.String())

	time.Sleep(100 * time.Millisecond)
	InitNode(context.Background(), false)
}

func TestRegisterNodes(t *testing.T) {

	RegisterNode()
	time.Sleep(100 * time.Millisecond)
	if !IsNodeRegistered(_from.Address) {
		t.Fatal("ERROR:", t.Name())
	}

	_from = eth.LoadAccount(_ks, "0x5cb50d3e5a4666fd90c4e6226942ee47ef400348", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	RegisterNode()
	time.Sleep(100 * time.Millisecond)
	if !IsNodeRegistered(_from.Address) {
		t.Fatal("ERROR:", t.Name())
	}

	_from = eth.LoadAccount(_ks, "0x10a99a5329b9f3ccae04a8a1afad03d4d2e9b1ff", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	RegisterNode()
	time.Sleep(100 * time.Millisecond)
	if !IsNodeRegistered(_from.Address) {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestHasNodeReputation(t *testing.T) {

	_from = eth.LoadAccount(_ks, "0xbfdcef8b53327344018a2c569d288f8249c4ff89", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)

	if !hasNodeReputation(_from.Address, 99) {
		t.Fatal("ERROR:", t.Name())
	}

	if !hasNodeReputation(_from.Address, 100) {
		t.Fatal("ERROR:", t.Name())
	}

	if hasNodeReputation(_from.Address, 101) {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestSendEvent(t *testing.T) {

	etype := types.EventType{
		RequiredTask:     types.PingNodeTask,
		Resource: types.NoResource,
	}
	SendEvent(&etype, 0, GetState())

	time.Sleep(100 * time.Millisecond)
	event := GetEvent(getClusterState().NextEventId - 1)
	if event.EType == "" ||
		event.Sender != _from.Address ||
		!utils.EmptyEthAddress(event.Solver.String()) ||
		event.Rcid > 0 ||
		event.SentAt.Uint64() == 0 ||
		event.SolvedAt.Uint64() != 0 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestSendReply(t *testing.T) {

	eid := getClusterState().NextEventId - 1

	_from = eth.LoadAccount(_ks, "0x5cb50d3e5a4666fd90c4e6226942ee47ef400348", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	SendReply(eid, GetState())

	_from = eth.LoadAccount(_ks, "0x10a99a5329b9f3ccae04a8a1afad03d4d2e9b1ff", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	SendReply(eid, GetState())

	time.Sleep(100 * time.Millisecond)
	replies := GetEventReplies(eid)
	if len(replies) != 3 ||
		replies[2].Replier != _from.Address ||
		replies[2].NodeState == "" ||
		len(replies[2].Voters) != 0 ||
		replies[2].RepliedAt.Uint64() == 0 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestVoteSolver(t *testing.T) {

	_from = eth.LoadAccount(_ks, "0xbfdcef8b53327344018a2c569d288f8249c4ff89", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)

	eid := getClusterState().NextEventId - 1

	VoteSolver(eid, _from.Address)

	_from = eth.LoadAccount(_ks, "0x5cb50d3e5a4666fd90c4e6226942ee47ef400348", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	VoteSolver(eid, common.HexToAddress("0xbfdcef8b53327344018a2c569d288f8249c4ff89"))

	_from = eth.LoadAccount(_ks, "0x10a99a5329b9f3ccae04a8a1afad03d4d2e9b1ff", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	VoteSolver(eid, common.HexToAddress("0xbfdcef8b53327344018a2c569d288f8249c4ff89"))

	time.Sleep(100 * time.Millisecond)
	if GetEvent(eid).Solver != common.HexToAddress("0xbfdcef8b53327344018a2c569d288f8249c4ff89") ||
		len(GetEventReplies(eid)[0].Voters) != 3 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestSolveEvent(t *testing.T) {

	_from = eth.LoadAccount(_ks, "0xbfdcef8b53327344018a2c569d288f8249c4ff89", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)

	eid := getClusterState().NextEventId - 1

	SolveEvent(eid)

	time.Sleep(100 * time.Millisecond)
	if GetEvent(eid).SolvedAt.Uint64() == 0 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestRegisterApplication(t *testing.T) {

	RegisterApplication(&inputs.AppInfo, []types.ContainerInfo{inputs.CtrInfo}, false)

	time.Sleep(100 * time.Millisecond)
	appid := getClusterState().NextAppId - 1
	rcid := getClusterState().NextCtrId - 1
	app := GetApplication(appid)
	ctr := GetContainer(rcid)
	if getClusterState().NextAppId != 2 ||
		app.Owner != common.HexToAddress("0xbfdcef8b53327344018a2c569d288f8249c4ff89") ||
		app.Info == "" ||
		len(GetApplicationContainers(appid)) != 1 ||
		app.RegisteredAt.Uint64() == 0 ||
		app.UnregisteredAt.Uint64() != 0 ||
		len(getActiveApplications()) != 1 ||
		getClusterState().NextCtrId != 2 ||
		ctr.Appid != appid ||
		ctr.Info == "" ||
		len(GetContainerInstances(rcid)) > 0 ||
		ctr.RegisteredAt.Uint64() == 0 ||
		ctr.UnregisteredAt.Uint64() != 0 ||
		len(GetActiveContainers()) != 0 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestClusterState(t *testing.T) {

	if existEvent(0) ||
		existApplication(0) ||
		existContainer(0) {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestRegisterContainer(t *testing.T) {

	RegisterContainer(getClusterState().NextAppId-1, &inputs.CtrInfo, false)

	time.Sleep(100 * time.Millisecond)
	if getClusterState().NextCtrId != 3 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestActivateContainer(t *testing.T) {

	RegisterContainer(getClusterState().NextAppId-1, &inputs.CtrInfo, true)

	time.Sleep(100 * time.Millisecond)
	rcid := getClusterState().NextCtrId - 1
	ActivateContainer(rcid)

	time.Sleep(100 * time.Millisecond)
	insts := GetContainerInstances(rcid)
	if len(insts) != 1 ||
		insts[0].Host != _from.Address ||
		insts[0].StartedAt.Uint64() == 0 ||
		!isContainerActive(rcid) ||
		len(GetActiveContainers()) != 1 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestUpdateContainerInfo(t *testing.T) {

	rcid := getClusterState().NextCtrId - 1
	inputs.CtrInfo.ServiceType = types.FrameworkServ
	inputs.CtrInfo.Impact = 5
	inputs.CtrInfo.ImageTag = "python"
	inputs.CtrInfo.Volumes = nil
	inputs.CtrInfo.Ports = nil

	UpdateContainerInfo(rcid, &inputs.CtrInfo)

	time.Sleep(100 * time.Millisecond)
	var cinfo types.ContainerInfo
	utils.UnmarshalJSON(GetContainer(rcid).Info, &cinfo)
	if cinfo.ServiceType != types.FrameworkServ ||
		cinfo.Impact != 5 ||
		cinfo.ImageTag != "python" ||
		cinfo.Volumes != nil ||
		cinfo.Ports != nil {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestUnregisterApplication(t *testing.T) {

	appid := getClusterState().NextAppId - 1

	UnregisterApplication(appid)

	time.Sleep(100 * time.Millisecond)
	app := GetApplication(appid)
	if app.UnregisteredAt.Uint64() == 0 ||
		len(GetApplicationContainers(appid)) != 0 ||
		len(getActiveApplications()) != 0 ||
		len(GetActiveContainers()) != 0 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestUnregisterContainer(t *testing.T) {

	RegisterApplication(&inputs.AppInfo, []types.ContainerInfo{inputs.CtrInfo}, true)

	time.Sleep(100 * time.Millisecond)
	rcid := getClusterState().NextCtrId - 1
	ActivateContainer(rcid)

	time.Sleep(100 * time.Millisecond)
	UnregisterContainer(rcid)

	time.Sleep(100 * time.Millisecond)
	ctr := GetContainer(rcid)
	appid := getClusterState().NextAppId - 1
	if ctr.UnregisteredAt.Uint64() == 0 ||
		isContainerActive(rcid) ||
		len(GetApplicationContainers(appid)) != 0 ||
		len(GetActiveContainers()) != 0 {
		t.Fatal("ERROR:", t.Name())
	}
}

func TestEventWithContainer(t *testing.T) {

	RegisterContainer(getClusterState().NextAppId-1, &inputs.CtrInfo, false)

	// Send event
	time.Sleep(100 * time.Millisecond)
	etype := types.EventType{
		RequiredTask:     types.MigrateContainerTask,
		Resource: types.CpuResource,
	}
	rcid := getClusterState().NextCtrId - 1
	SendEvent(&etype, rcid, GetState())

	time.Sleep(100 * time.Millisecond)
	eid := getClusterState().NextEventId - 1
	event := GetEvent(eid)
	if event.Rcid == 0 {
		t.Fatal("ERROR:", t.Name())
	}

	// Send replies
	_from = eth.LoadAccount(_ks, "0x5cb50d3e5a4666fd90c4e6226942ee47ef400348", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	SendReply(eid, GetState())
	_from = eth.LoadAccount(_ks, "0x10a99a5329b9f3ccae04a8a1afad03d4d2e9b1ff", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	SendReply(eid, GetState())

	// Vote solver
	time.Sleep(100 * time.Millisecond)
	_from = eth.LoadAccount(_ks, "0xbfdcef8b53327344018a2c569d288f8249c4ff89", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	VoteSolver(eid, common.HexToAddress("0x10a99a5329b9f3ccae04a8a1afad03d4d2e9b1ff"))
	_from = eth.LoadAccount(_ks, "0x5cb50d3e5a4666fd90c4e6226942ee47ef400348", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	VoteSolver(eid, common.HexToAddress("0x10a99a5329b9f3ccae04a8a1afad03d4d2e9b1ff"))
	_from = eth.LoadAccount(_ks, "0x10a99a5329b9f3ccae04a8a1afad03d4d2e9b1ff", "12345678")
	_nonce, _ = _ethc.PendingNonceAt(context.Background(), _from.Address)
	VoteSolver(eid, common.HexToAddress("0x10a99a5329b9f3ccae04a8a1afad03d4d2e9b1ff"))

	time.Sleep(100 * time.Millisecond)
	event = GetEvent(eid)
	insts := GetContainerInstances(rcid)
	if utils.EmptyEthAddress(event.Solver.String()) ||
		len(insts) != 1 ||
		utils.EmptyEthAddress(insts[0].Host.String()) ||
		insts[0].StartedAt.Uint64() != 0 {
		t.Fatal("ERROR:", t.Name())
	}

	// Solve event
	time.Sleep(100 * time.Millisecond)
	SolveEvent(eid)

	time.Sleep(100 * time.Millisecond)
	event = GetEvent(eid)
	insts = GetContainerInstances(rcid)
	if event.SolvedAt.Uint64() == 0 ||
		len(insts) != 1 ||
		utils.EmptyEthAddress(insts[0].Host.String()) ||
		insts[0].StartedAt.Uint64() == 0 ||
		!isContainerActive(rcid) {
		t.Fatal("ERROR:", t.Name())
	}
}*/
