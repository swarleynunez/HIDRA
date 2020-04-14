package main

import (
	"fmt"
	"github.com/swarleynunez/superfog/core/daemon"
	"github.com/swarleynunez/superfog/core/eth"
	"github.com/swarleynunez/superfog/core/utils"
	"time"
)

const (
	EthNodeIp      = "127.0.0.1"
	EthNodePort    = 7545
	EthNodeDir     = ".ethereum"
	EthKeystoreDir = "keystore"

	AccountPassphrase = "12345678"

	MonitorInterval = 1 * time.Second
)

func init() {

}

func main() {

	fmt.Println(time.Now().Unix())

	// Connect to the Ethereum node
	url := utils.FormatUrl(EthNodeIp, EthNodePort, utils.HttpMode)
	client := eth.Connect(url)
	_ = client

	// Gets the path to keystore directory
	keydir := utils.FormatPath(EthNodeDir, EthKeystoreDir)

	// Load the keystore
	ks := eth.LoadKeystore(keydir)
	_ = ks

	// Create an account
	//account0 := eth.CreateAccount(ks, AccountPassphrase)
	//fmt.Println(account0)

	// Sending ether
	//toAddress := common.HexToAddress("0x124dfc4Fba0eB7EB185b3CdcB0F91Dc273826AC6")
	//eth.SendTxn(client, ks, toAddress, AccountPassphrase, 1000000000000000000)

	// Initialize host state (specs and first state)
	hostSpecs, hostState := daemon.InitState()

	// Update host state
	hostState = daemon.UpdateState()
	_, _ = hostSpecs, hostState

	//for {
	//	select {}
	//}

	fmt.Println(time.Now().Unix())
}
