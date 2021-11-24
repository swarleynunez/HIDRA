package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/superfog/core/managers"
)

const registerShortMsg = "Register node in the configured controller smart contract"

var registerCmd = &cobra.Command{
	Use:                   "register",
	Short:                 registerShortMsg,
	Long:                  title + "\n\n" + "Info:\n  " + registerShortMsg,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize and configure node
		managers.InitNode(ctx, false)

		// Register node if it has not done yet
		if !managers.IsNodeRegistered(managers.GetFromAccount()) {
			managers.RegisterNode(ctx)

			fmt.Println("--> Node registered")
		} else {
			fmt.Println("--> Node is already registered")
		}
	},
}
