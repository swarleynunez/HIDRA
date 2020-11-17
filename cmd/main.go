package main

import (
	"context"
	"github.com/swarleynunez/superfog/core/daemons"
	"github.com/swarleynunez/superfog/core/docker"
	"github.com/swarleynunez/superfog/core/eth"
	"github.com/swarleynunez/superfog/core/managers"
	"github.com/swarleynunez/superfog/core/utils"
	"os"
)

func main() {

	// Load .env configuration
	utils.LoadEnv()

	var (
		nodeDir = os.Getenv("ETH_NODE_DIR")
		addr    = os.Getenv("NODE_ADDR")
		pass    = os.Getenv("NODE_ADDR_PASS")
	)

	// Connect to the Ethereum node
	ethc := eth.Connect(utils.FormatPath(nodeDir, "geth.ipc"))

	// Connect to the Docker node
	dcli := docker.Connect()

	// Load keystore
	keypath := utils.FormatPath(nodeDir, "keystore")
	ks := eth.LoadKeystore(keypath)

	// Load and unlock an account
	from := eth.LoadAccount(ks, addr, pass)

	// Initialize and configure node
	managers.InitNode(ethc, dcli, ks, from)

	// Main function empty context
	ctx := context.Background()

	// Main loop
	daemons.StartMonitor(ctx)

	///////////////
	//// Other ////
	///////////////
	// Send ether
	//wei := new(big.Int)
	//wei.SetString("10000000000000000000", 10)
	//to := common.HexToAddress("0x5cb50d3E5a4666FD90c4E6226942EE47eF400348")
	//eth.SendEther(ethc, from, to, wei, ks, pass)
}
