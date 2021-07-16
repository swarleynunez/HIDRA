package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/superfog/core/managers"
	"time"
)

var registerCmd = &cobra.Command{
	Use:                   "register",
	Short:                 "Register node in the configured controller smart contract",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize and configure node
		managers.InitNode(ctx, false)

		// Register node if it has not done yet
		if !managers.IsNodeRegistered(managers.GetFromAccount()) {
			managers.RegisterNode()

			// Debug
			fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Node registered\n")
		} else {
			// Debug
			fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Node is already registered\n")
		}
	},
}
