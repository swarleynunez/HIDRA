package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swarleynunez/hidra/core/managers"
	"github.com/swarleynunez/hidra/core/utils"
	"strconv"
)

const registerShortMsg = "Register fog node in the configured controller smart contract"

var registerCmd = &cobra.Command{
	Use:                   "register port",
	Short:                 registerShortMsg,
	Long:                  title + "\n\n" + "Info:\n  " + registerShortMsg,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize and configure node
		managers.InitNode(ctx, false)

		// Register node if it has not done yet
		if !managers.IsNodeRegistered(managers.GetFromAccount()) {
			port, err := strconv.ParseUint(args[0], 10, 16)
			utils.CheckError(err, utils.FatalMode)

			managers.RegisterNode(ctx, uint16(port))

			fmt.Println("--> Node registered")
		} else {
			fmt.Println("--> Node is already registered")
		}
	},
}
