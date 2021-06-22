package cmd

import (
	"github.com/spf13/cobra"
	"github.com/swarleynunez/superfog/core/managers"
)

var deployCmd = &cobra.Command{
	Use:                   "deploy",
	Short:                 "Deploy and configure a new controller smart contract",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize and configure node
		managers.InitNode(ctx, true)

		// Deploy a new controller
		managers.DeployController()
	},
}
