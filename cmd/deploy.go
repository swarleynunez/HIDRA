package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/superfog/core/managers"
	"github.com/swarleynunez/superfog/core/utils"
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

		// Save the controller address
		utils.SetEnv("CONTROLLER_ADDR", caddr.String())

		fmt.Println("--> Controller address: ", caddr.String())
	},
}
