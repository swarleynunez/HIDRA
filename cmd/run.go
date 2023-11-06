package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/hidra/core/daemons"
	"github.com/swarleynunez/hidra/core/managers"
	"os"
)

const runShortMsg = "Run orchestrator daemons (monitor, enforcer and watchers)"

var runCmd = &cobra.Command{
	Use:                   "run interface",
	Short:                 runShortMsg,
	Long:                  title + "\n\n" + "Info:\n  " + runShortMsg,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize and configure node
		managers.InitNode(ctx, false)

		// Check if node is registered
		if !managers.IsNodeRegistered(managers.GetFromAccount()) {
			fmt.Println("--> Node not registered at loaded controller contract")
			os.Exit(0)
		}

		// Main loop
		daemons.Run(ctx, args[0])
	},
}
