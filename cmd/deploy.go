package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/hidra/core/managers"
	"github.com/swarleynunez/hidra/core/utils"
)

const deployShortMsg = "Deploy and configure a new controller smart contract"

var deployCmd = &cobra.Command{
	Use:                   "deploy",
	Short:                 deployShortMsg,
	Long:                  title + "\n\n" + "Info:\n  " + deployShortMsg,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize and configure node
		managers.InitNode(ctx, true)

		// Deploy a new controller
		caddr := managers.DeployController(ctx)

		// Save the controller contract address
		utils.SetEnv("CONTROLLER_ADDR", caddr.String())

		fmt.Println("--> Controller contract:", caddr.String())
	},
}
