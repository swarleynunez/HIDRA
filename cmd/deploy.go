package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/superfog/core/managers"
	"github.com/swarleynunez/superfog/core/utils"
	"time"
)

var deployCmd = &cobra.Command{
	Use:                   "deploy",
	Short:                 "Deploy and configure a new controller smart contract",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize and configure node
		managers.InitNode(ctx, true)

		// Deploy a new controller
		caddr := managers.DeployController()

		// Save the controller address
		utils.SetEnvKey("CONTROLLER_ADDR", caddr.String())

		// Debug
		fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Controller address: ", caddr.String(), "\n")
	},
}
