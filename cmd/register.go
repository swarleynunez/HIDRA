package cmd

import (
	"github.com/spf13/cobra"
	"github.com/swarleynunez/superfog/core/managers"
)

var registerCmd = &cobra.Command{
	Use:                   "register",
	Short:                 "Register node in the configured controller smart contract",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize and configure node
		managers.InitNode(ctx, false)

		// Register node if it has not done yet
		managers.RegisterNode()
	},
}
