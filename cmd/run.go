package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/superfog/core/daemons"
	"github.com/swarleynunez/superfog/core/managers"
	"os"
	"time"
)

var runCmd = &cobra.Command{
	Use:                   "run",
	Short:                 "Run orchestrator daemons (monitor, enforcer and watchers)",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize and configure node
		managers.InitNode(ctx, false)

		// Check if node is registered
		if !managers.IsNodeRegistered(managers.GetFromAccount().Address) {
			// Debug
			fmt.Print("[", time.Now().Format("15:04:05.000000"), "] ", "Node not registered at loaded controller address\n")
			os.Exit(0)
		}

		// Main loop
		daemons.Run(ctx)
	},
}
